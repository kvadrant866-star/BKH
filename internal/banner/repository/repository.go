package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"banner/internal/agregator"
	"banner/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) UpsertBatch(ctx context.Context, deltas []agregator.Delta) error {
	if len(deltas) == 0 {
		return nil
	}
	values := make([]string, 0, len(deltas))
	args := make([]any, 0, len(deltas)*3)
	for i, d := range deltas {
		p1 := i*3 + 1
		values = append(values, fmt.Sprintf("($%d,$%d,$%d)", p1, p1+1, p1+2))
		args = append(args, d.BannerID, d.TSMinute.UTC(), d.Count)
	}
	query := "INSERT INTO banner_minute_stats (banner_id, minute_ts, count) VALUES " +
		strings.Join(values, ",") +
		" ON CONFLICT (banner_id, minute_ts) DO UPDATE SET count = banner_minute_stats.count + EXCLUDED.count"
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *PostgresRepository) SelectStats(ctx context.Context, bannerID int, from, to time.Time) ([]entity.MinuteStat, error) {
	rows, err := r.db.Query(ctx, `
		SELECT minute_ts, count
		FROM banner_minute_stats
		WHERE banner_id = $1 AND minute_ts >= $2 AND minute_ts < $3
		ORDER BY minute_ts
	`, bannerID, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]entity.MinuteStat, 0)
	for rows.Next() {
		var ts time.Time
		var c int64
		if err := rows.Scan(&ts, &c); err != nil {
			return nil, err
		}
		out = append(out, entity.MinuteStat{Ts: ts.UTC(), V: c})
	}
	return out, rows.Err()
}
