package pgstorage

import (
	"database/sql"
	"errors"

	"github.com/ibeloyar/metrics/internal/model"
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

/*
Доработайте сервис и добавьте поддержку СУБД PostgreSQL для хранения метрик.

1) Сервису нужно самостоятельно создать все необходимые таблицы в базе данных. Схема и формат хранения остаются на ваше усмотрение.
2) Используйте инструмент миграций для создания и изменения схемы базы данных.
	Для хранения значений gauge рекомендуется использовать тип double precision.
3) При отсутствии переменной окружения DATABASE_DSN или флага командной строки -d либо при их пустых значениях, вернитесь последовательно к:
	хранению метрик в файле — при наличии соответствующей переменной окружения или флага командной строки;
	хранению метрик в памяти.
*/
