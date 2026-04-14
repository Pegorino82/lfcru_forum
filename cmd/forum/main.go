package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Pegorino82/lfcru_forum/internal/admin"
	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/Pegorino82/lfcru_forum/internal/cleanup"
	"github.com/Pegorino82/lfcru_forum/internal/comment"
	"github.com/Pegorino82/lfcru_forum/internal/config"
	"github.com/Pegorino82/lfcru_forum/internal/forum"
	"github.com/Pegorino82/lfcru_forum/internal/home"
	appMiddleware "github.com/Pegorino82/lfcru_forum/internal/middleware"
	"github.com/Pegorino82/lfcru_forum/internal/match"
	"github.com/Pegorino82/lfcru_forum/internal/news"
	"github.com/Pegorino82/lfcru_forum/internal/ratelimit"
	"github.com/Pegorino82/lfcru_forum/internal/session"
	"github.com/Pegorino82/lfcru_forum/internal/tmpl"
	"github.com/Pegorino82/lfcru_forum/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Load()

	// Подключение к PostgreSQL
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to ping database: %v\n", err)
		os.Exit(1)
	}

	// Миграции
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Репозитории
	userRepo := user.NewRepo(pool)
	sessionRepo := session.NewRepo(pool)
	attemptRepo := ratelimit.NewLoginAttemptRepo(pool)
	newsRepo := news.NewRepo(pool)
	matchRepo := match.NewRepo(pool)
	topicRepo := forum.NewRepo(pool)
	commentRepo := comment.NewRepo(pool)

	// Сервис аутентификации
	authCfg := auth.Config{
		BcryptCost:         cfg.BcryptCost,
		SessionLifetime:    cfg.SessionLifetime,
		RateLimitWindow:    cfg.RateLimitWindow,
		RateLimitMax:       cfg.RateLimitMax,
		SessionGracePeriod: cfg.SessionGracePeriod,
		MaxSessionsPerUser: cfg.MaxSessionsPerUser,
		CookieSecure:       cfg.CookieSecure,
	}
	authSvc := auth.NewService(userRepo, sessionRepo, attemptRepo, authCfg)

	// Сервис комментариев
	commentSvc := comment.NewService(commentRepo, userRepo)

	// Сервис форума
	forumSvc := forum.NewService(topicRepo)

	// Шаблоны
	renderer, err := tmpl.New(os.DirFS("templates"), "templates/")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load templates: %v\n", err)
		os.Exit(1)
	}

	// Echo
	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(appMiddleware.CSRFMiddleware())
	e.Use(auth.LoadSession(authSvc))

	// Хэндлеры и маршруты
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/", home.NewHandler(newsRepo, matchRepo, topicRepo).ShowHome)
	auth.NewHandler(authSvc).RegisterRoutes(e)
	news.NewHandler(newsRepo, commentRepo, commentSvc).RegisterRoutes(e)

	// Forum routes
	forumHandler := forum.NewHandler(forumSvc, renderer)

	// Moderator-only routes (require auth + role)
	modGroup := e.Group("", auth.RequireAuth, auth.RequireRole(renderer, "moderator", "admin"))
	modGroup.GET("/forum/sections/new", forumHandler.NewSection)
	modGroup.POST("/forum/sections", forumHandler.CreateSection)
	modGroup.GET("/forum/sections/:id/topics/new", forumHandler.NewTopic)
	modGroup.POST("/forum/sections/:id/topics", forumHandler.CreateTopic)

	// Public routes
	e.GET("/forum", forumHandler.Index)
	e.GET("/forum/sections/:id", forumHandler.ShowSection)
	e.GET("/forum/topics/:id", forumHandler.ShowTopic)

	// Auth-only routes
	authGroup := e.Group("", auth.RequireAuth)
	authGroup.POST("/forum/topics/:id/posts", forumHandler.CreatePost)

	// Admin routes
	imagesRepo := admin.NewImagesRepo(pool)
	imgSvc := admin.NewImageService(cfg.UploadsDir)
	imagesHandler := admin.NewImagesHandler(imagesRepo, imgSvc)
	forumAdminHandler := admin.NewForumHandler(forumSvc)

	adminGroup := e.Group("", admin.RequireAdminOrMod(renderer))
	adminGroup.GET("/admin", admin.NewHandler().Dashboard)
	adminGroup.POST("/admin/articles/:id/images", imagesHandler.Upload)
	adminGroup.DELETE("/admin/articles/:id/images/:image_id", imagesHandler.Delete)

	// Admin forum routes
	adminGroup.GET("/admin/forum/sections", forumAdminHandler.ListSections)
	adminGroup.GET("/admin/forum/sections/new", forumAdminHandler.NewSection)
	adminGroup.POST("/admin/forum/sections", forumAdminHandler.CreateSection)
	adminGroup.GET("/admin/forum/sections/:id/edit", forumAdminHandler.EditSection)
	adminGroup.POST("/admin/forum/sections/:id", forumAdminHandler.UpdateSection)
	adminGroup.GET("/admin/forum/sections/:id/topics", forumAdminHandler.ListTopics)
	adminGroup.GET("/admin/forum/sections/:id/topics/new", forumAdminHandler.NewTopic)
	adminGroup.POST("/admin/forum/sections/:id/topics", forumAdminHandler.CreateTopic)
	adminGroup.GET("/admin/forum/topics/:id/edit", forumAdminHandler.EditTopic)
	adminGroup.POST("/admin/forum/topics/:id", forumAdminHandler.UpdateTopic)

	// Фоновая очистка
	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()
	go cleanup.Run(bgCtx, sessionRepo, attemptRepo)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(":" + cfg.AppPort); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	<-ctx.Done()
	bgCancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown error: %v\n", err)
	}
}

func runMigrations(databaseURL string) error {
	goose.SetBaseFS(nil)

	db, err := goose.OpenDBWithDriver("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		if strings.Contains(err.Error(), "no migration files found") {
			fmt.Println("migrations: no files found, skipping")
			return nil
		}
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}
