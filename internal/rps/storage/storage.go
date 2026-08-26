package storage

import (
	"context"
	"database/sql"
	"url_shortener/internal/rps/model"
)

type RPSStorage interface {
	Insert(ctx context.Context, payload string, ts int64, duration int64) (int64, error)
	SelectJoin(ctx context.Context, limit int) ([]model.JoinRow, error)
	SelectSimple(ctx context.Context, limit int) ([]model.JoinRow, error)
	Update(ctx context.Context) error
}

type RPS struct {
	db *sql.DB
}

func NewRPS(db *sql.DB) *RPS {
	return &RPS{db: db}
}

func (s *RPS) Insert(ctx context.Context, payload string, ts int64, duration int64) (int64, error) {
	res, err := s.db.ExecContext(ctx, "INSERT INTO rps_log(payload, ts, duration) VALUES (?, ?, ?)", payload, ts, duration)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *RPS) SelectJoin(ctx context.Context, limit int) ([]model.JoinRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT rps_log.id, rps_log.payload, rps_log.ts, rps_log.duration,
		       rps_meta.key, rps_meta.value
		FROM rps_log
		LEFT JOIN rps_meta ON rps_log.id = rps_meta.log_id
		ORDER BY rps_log.id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.JoinRow, 0, limit)
	for rows.Next() {
		var r model.JoinRow
		var mk, mv sql.NullString
		if err := rows.Scan(&r.ID, &r.Payload, &r.Ts, &r.Duration, &mk, &mv); err != nil {
			return nil, err
		}
		r.MetaKey = mk.String
		r.MetaValue = mv.String
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *RPS) SelectSimple(ctx context.Context, limit int) ([]model.JoinRow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, payload, ts, duration
		FROM rps_log
		ORDER BY id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]model.JoinRow, 0, limit)
	for rows.Next() {
		var r model.JoinRow
		if err := rows.Scan(&r.ID, &r.Payload, &r.Ts, &r.Duration); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *RPS) Update(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "UPDATE rps_log WHERE id=1 SET duration = duration + 1")
	return err
}
