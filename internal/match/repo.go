package match

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ pool *pgxpool.Pool }

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// NextUpcoming возвращает ближайший будущий матч (match_date >= asOf)
// или nil, если будущих матчей нет.
func (r *Repo) NextUpcoming(ctx context.Context, asOf time.Time) (*Match, error) {
	var m Match
	err := r.pool.QueryRow(ctx, `
		SELECT id, opponent, match_date, tournament, is_home
		FROM matches
		WHERE match_date >= $1
		ORDER BY match_date ASC
		LIMIT 1
	`, asOf).Scan(&m.ID, &m.Opponent, &m.MatchDate, &m.Tournament, &m.IsHome)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}
