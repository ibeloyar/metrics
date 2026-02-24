package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ibeloyar/metrics/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	migrationsTable = "schema_migrations"
	schemaName      = "public"
	migrationsPath  = "./migrations"

	maxAttempts = 3
)

// PGStorage represents PostgreSQL storage for metrics with retry logic.
type PGStorage struct {
	db         *sql.DB
	classifier *PostgresErrorClassifier
}

// New creates new PGStorage instance with automatic schema migrations.
//
// Automatically runs migrations from ./migrations directory.
// Connection pool is created with pgxpool under the hood.
//
// connStr should be a valid PostgreSQL connection string.
//
// Returns error if:
//   - cannot connect to database
//   - migration fails (except migrate.ErrNoChange)
//   - absolute path resolution fails
func New(connStr string) (*PGStorage, error) {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	db := stdlib.OpenDBFromPool(pool)

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: migrationsTable,
		SchemaName:      schemaName,
	})
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+absPath, "postgres", driver)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	return &PGStorage{
		db:         db,
		classifier: NewPostgresErrorClassifier(),
	}, nil
}

// Ping checks database connectivity.
//
// Returns nil if connection is healthy, otherwise returns the error.
func (s *PGStorage) Ping() error {
	return s.db.Ping()
}

// GetMetric retrieves single metric by ID.
//
// Returns nil if metric doesn't exist (sql.ErrNoRows).
// Other errors are silently converted to nil (after retry attempts).
func (s *PGStorage) GetMetric(name string) *model.Metrics {
	var m model.Metrics
	query := `SELECT id, mtype, delta, value, hash FROM metrics WHERE id = $1`

	err := s.executeWithRetryConnection(func(db *sql.DB) error {
		row := db.QueryRow(query, name)
		return row.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash)
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return nil // Другие ошибки (включая исчерпанные retry)
	}
	return &m
}

// GetMetrics returns all metrics as ID -> Metrics map.
//
// Returns empty map on any error (after retry attempts).
func (s *PGStorage) GetMetrics() map[string]model.Metrics {
	result := make(map[string]model.Metrics)

	err := s.executeWithRetryConnection(func(db *sql.DB) error {
		query := `SELECT id, mtype, delta, value, hash FROM metrics`
		rows, err := db.Query(query)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m model.Metrics
			if err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value, &m.Hash); err != nil {
				return err
			}
			result[m.ID] = m
		}
		return rows.Err()
	})

	if err != nil {
		return make(map[string]model.Metrics) // Возвращаем пустую карту
	}
	return result
}

// SetMetric stores or updates single metric.
//
// Uses UPSERT (INSERT ... ON CONFLICT DO UPDATE) semantics.
// All fields are overwritten on conflict.
func (s *PGStorage) SetMetric(metric *model.Metrics) error {
	// ON CONFLICT (id) DO UPDATE SET - если строка с id уже сущевствует, сделать UPDATE
	query := `INSERT INTO metrics (id, mtype, delta, value, hash) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
      		mtype = EXCLUDED.mtype,
      		delta = EXCLUDED.delta,
      		value = EXCLUDED.value,
      		hash = EXCLUDED.hash
    `

	return s.executeWithRetryConnection(func(db *sql.DB) error {
		_, err := db.Exec(query,
			metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash,
		)
		return err
	})

}

// SetMetrics stores or updates multiple metrics atomically.
//
// Uses transaction with 6s timeout. Counter metrics accumulate delta values:
//
//	delta = COALESCE(existing.delta, 0) + new.delta
//
// Gauge metrics overwrite all fields. Returns error if transaction fails.
func (s *PGStorage) SetMetrics(metrics []model.Metrics) error {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	return s.executeWithRetryConnection(func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback() // гарантированный откат

		query := `INSERT INTO metrics (id, mtype, delta, value, hash) 
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET
                mtype = EXCLUDED.mtype,
                delta = COALESCE(metrics.delta, 0) + EXCLUDED.delta,
                value = EXCLUDED.value,
                hash = EXCLUDED.hash`

		for _, metric := range metrics {
			_, err := tx.ExecContext(ctx, query,
				metric.ID, metric.MType, metric.Delta, metric.Value, metric.Hash,
			)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	})
}

// IncrementCountMetricValue increments counter metric delta.
//
// Only works with 'counter' type metrics. Accumulates delta:
//
//	new_delta = COALESCE(existing.delta, 0) + delta
//
// Panics if delta is nil. Only updates counters (ignores other types).
func (s *PGStorage) IncrementCountMetricValue(name string, delta *int64) error {
	if delta == nil {
		return errors.New("cannot increment metric value, delta is nil")
	}

	return s.executeWithRetryConnection(func(db *sql.DB) error {
		query := `INSERT INTO metrics (id, delta, mtype, hash) VALUES ($2, $1, 'counter', '')
            ON CONFLICT (id) DO UPDATE
            SET delta = COALESCE(metrics.delta, 0) + $1
            WHERE metrics.mtype = 'counter'`

		_, err := db.Exec(query, *delta, name)
		return err
	})
}

// Shutdown closes database connection gracefully.
//
// Should be called before application exit.
func (s *PGStorage) Shutdown() error {
	return s.db.Close()
}

func (s *PGStorage) executeWithRetryConnection(operation func(*sql.DB) error) error {
	err := operation(s.db)
	if err == nil {
		return nil
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if s.classifier.Classify(err) != Retriable {
			return err
		}

		delay := getAttemptDelay(attempt)
		time.Sleep(delay)

		err = operation(s.db)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return lastErr // Возвращаем последнюю ошибку после 3 попыток
}

func getAttemptDelay(attempt int) time.Duration {
	switch attempt {
	case 0:
		return 1 * time.Second
	case 1:
		return 3 * time.Second
	default: // attempt >= 2
		return 5 * time.Second
	}
}
