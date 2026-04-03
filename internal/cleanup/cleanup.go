package cleanup

import (
	"context"
	"fmt"
	"time"
)

type sessionCleaner interface {
	DeleteExpired(ctx context.Context) (int64, error)
}

type attemptCleaner interface {
	Cleanup(ctx context.Context) (int64, error)
}

// Run starts background goroutines that periodically clean up expired sessions
// and old login attempts. It blocks until ctx is cancelled.
func Run(ctx context.Context, sessions sessionCleaner, attempts attemptCleaner) {
	sessionTicker := time.NewTicker(time.Hour)
	attemptTicker := time.NewTicker(10 * time.Minute)
	defer sessionTicker.Stop()
	defer attemptTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sessionTicker.C:
			n, err := sessions.DeleteExpired(ctx)
			if err != nil {
				fmt.Printf("cleanup: delete expired sessions: %v\n", err)
			} else {
				fmt.Printf("cleanup: deleted %d expired sessions\n", n)
			}
		case <-attemptTicker.C:
			n, err := attempts.Cleanup(ctx)
			if err != nil {
				fmt.Printf("cleanup: delete old login attempts: %v\n", err)
			} else {
				fmt.Printf("cleanup: deleted %d old login attempts\n", n)
			}
		}
	}
}
