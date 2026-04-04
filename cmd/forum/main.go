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

	"github.com/Pegorino82/lfcru_forum/internal/auth"
	"github.com/Pegorino82/lfcru_forum/internal/cleanup"
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

	// Репозитории для главной страницы
	newsRepo := news.NewRepo(pool)
	matchRepo := match.NewRepo(pool)
	topicRepo := forum.NewRepo(pool)

	// Хэндлер главной страницы
	homeHandler := home.NewHandler(newsRepo, matchRepo, topicRepo)

	// Маршруты
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/", homeHandler.ShowHome)
	auth.NewHandler(authSvc).RegisterRoutes(e)

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
