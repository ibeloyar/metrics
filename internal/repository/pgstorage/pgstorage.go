package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/ibeloyar/metrics/internal/model"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type PGStorage struct {
	db *sql.DB
}

func New(connStr string) (*PGStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_migrations",
		SchemaName:      "public",
	})
	if err != nil {
		return nil, err
	}

	migrationsPath := "./migrations"
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+absPath, "postgres", driver)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	return &PGStorage{
		db: db,
	}, nil
}

func (s *PGStorage) Ping() error {
	return s.db.Ping()
}

func (s *PGStorage) GetMetric(name string) *model.Metrics {
	var m model.Metrics
	query := `SELECT id, mtype, delta, value, hash FROM metrics WHERE id = $1`
	row := s.db.QueryRow(query, name)
	err := row.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		// Обработка ошибок по необходимости
		return nil
	}
	return &m
}

func (s *PGStorage) GetMetrics() map[string]model.Metrics {
	result := make(map[string]model.Metrics)

	query := `SELECT id, mtype, delta, value, hash FROM metrics`

	rows, err := s.db.Query(query)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var m model.Metrics
		err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
		if err == nil {
			result[m.ID] = m
		}
	}

	if err := rows.Err(); err != nil {
		return make(map[string]model.Metrics)
	}

	return result
}

func (s *PGStorage) SetMetric(metric model.Metrics) error {
	// ON CONFLICT (id) DO UPDATE SET - если строка с id уже сущевствует, сделать UPDATE
	query := `INSERT INTO metrics (id, mtype, delta, value, hash) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
      		mtype = EXCLUDED.mtype,
      		delta = EXCLUDED.delta,
      		value = EXCLUDED.value,
      		hash = EXCLUDED.hash
    `

	_, err := s.db.Exec(query,
		metric.ID,
		metric.MType,
		metric.Delta,
		metric.Value,
		metric.Hash,
	)
	return err
}

func (s *PGStorage) SetMetrics(metrics []model.Metrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	for _, metric := range metrics {
		// ON CONFLICT (id) DO UPDATE SET - если строка с id уже сущевствует, сделать UPDATE
		query := `INSERT INTO metrics (id, mtype, delta, value, hash) 
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET
				mtype = EXCLUDED.mtype,
				delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta,
				value = EXCLUDED.value,
				hash = EXCLUDED.hash
		`

		_, err := tx.ExecContext(ctx, query,
			metric.ID,
			metric.MType,
			metric.Delta,
			metric.Value,
			metric.Hash,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *PGStorage) IncrementCountMetricValue(name string, delta *int64) error {
	if delta == nil {
		return errors.New("cannot increment metric value, delta is nil")
	}

	query := `INSERT INTO metrics (id, delta, mtype, hash) VALUES ($2, $1, 'counter', '')
		ON CONFLICT (id) DO UPDATE
		SET delta = COALESCE(metrics.delta, 0) + $1
		WHERE metrics.mtype = 'counter';`

	_, err := s.db.Exec(query, *delta, name)

	return err
}

func (s *PGStorage) Shutdown() error {
	return s.db.Close()
}

// Задание по треку «Сервис сбора метрик и алертинга»
// Сервер:
//	Добавьте новый хендлер POST /updates/, принимающий в теле запроса множество метрик в формате: []Metrics (списка метрик).
// Агент:
//	Научите агент работать с использованием нового API (отправлять метрики батчами).
//
// Стоит помнить, что:
// 	- нужно соблюдать обратную совместимость;
//  - отправлять пустые батчи не нужно;
//  - вы умеете сжимать контент по алгоритму gzip;
//  - изменение в базе можно выполнять в рамках одной транзакции или одного запроса;
//  - необходимо избегать формирования условий для возникновения состояния гонки (race condition).
