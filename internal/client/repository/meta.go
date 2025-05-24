package repository

import (
	"context"
	"database/sql"
	"time"
)

type MetaRepository struct {
	db *sql.DB
}

func NewMetaRepository(db *sql.DB) (*MetaRepository, error) {
	repo := &MetaRepository{db: db}
	if err := repo.init(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *MetaRepository) GetLastSync(ctx context.Context) (time.Time, error) {
	var tsInt int64
	err := r.db.QueryRowContext(ctx, "SELECT last_sync FROM meta WHERE id = 0").Scan(&tsInt)

	return time.Unix(tsInt, 0).UTC(), err
}

func (r *MetaRepository) SetLastSync(ctx context.Context, t time.Time) error {
	_, err := r.db.ExecContext(ctx, "UPDATE meta SET last_sync = ? WHERE id = 0", t.Unix())

	return err
}

func (r *MetaRepository) GetMasterPasswordHash(ctx context.Context) (string, error) {
	var h string
	err := r.db.QueryRowContext(ctx, "SELECT master_password_hash FROM meta WHERE id = 0").Scan(&h)

	return h, err
}

func (r *MetaRepository) SetMasterPasswordHash(ctx context.Context, h string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE meta SET master_password_hash = ? WHERE id = 0", h)

	return err
}

func (r *MetaRepository) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS meta (
		  id INTEGER PRIMARY KEY CHECK (id = 0),
		  last_sync INTEGER NOT NULL,
          master_password_hash TEXT NOT NULL
		)`,
		`INSERT OR IGNORE INTO meta (id, last_sync, master_password_hash) VALUES (0, 0, "")`,
	}

	for _, query := range queries {
		_, err := r.db.Exec(query)

		if err != nil {
			return err
		}
	}

	return nil
}
