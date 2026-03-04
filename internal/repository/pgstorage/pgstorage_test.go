package pgstorage

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ibeloyar/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func int64Ptr(i int64) *int64       { return &i }
func float64Ptr(f float64) *float64 { return &f }

func TestPGStorage_GetMetric_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PGStorage{db: db, classifier: &PostgresErrorClassifier{}}

	rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value", "hash"}).
		AddRow("metric1", "counter", int64(100), nil, "")

	mock.ExpectQuery("SELECT id, mtype, delta, value, hash FROM metrics WHERE id = \\$1").
		WithArgs("metric1").
		WillReturnRows(rows)

	result := store.GetMetric("metric1")

	assert.NotNil(t, result)
	assert.Equal(t, "metric1", result.ID)
	assert.Equal(t, model.Counter, result.MType)
	assert.Equal(t, int64(100), *result.Delta)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_GetMetric_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PGStorage{db: db, classifier: &PostgresErrorClassifier{}}

	mock.ExpectQuery("SELECT id, mtype, delta, value, hash FROM metrics WHERE id = \\$1").
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	result := store.GetMetric("missing")

	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_GetMetrics_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PGStorage{db: db, classifier: &PostgresErrorClassifier{}}

	rows := sqlmock.NewRows([]string{"id", "mtype", "delta", "value", "hash"}).
		AddRow("counter1", "counter", int64(100), nil, "").
		AddRow("gauge1", "gauge", nil, float64(10.5), "")

	mock.ExpectQuery("SELECT id, mtype, delta, value, hash FROM metrics").
		WillReturnRows(rows)

	result := store.GetMetrics()

	assert.Len(t, result, 2)
	assert.Equal(t, int64(100), *result["counter1"].Delta)
	assert.Equal(t, float64(10.5), *result["gauge1"].Value)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_SetMetric_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PGStorage{db: db, classifier: &PostgresErrorClassifier{}}
	metric := &model.Metrics{ID: "test", MType: "counter", Delta: int64Ptr(42)}

	mock.ExpectExec(`INSERT INTO metrics \(id, mtype, delta, value, hash\) 
	                 VALUES \(\$1, \$2, \$3, \$4, \$5\) 
	                 ON CONFLICT \(id\) DO UPDATE SET
	                     mtype = EXCLUDED\.mtype,
	                     delta = EXCLUDED\.delta,
	                     value = EXCLUDED\.value,
	                     hash = EXCLUDED\.hash`).
		WithArgs("test", "counter", int64(42), nil, "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.SetMetric(metric)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_SetMetrics_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PGStorage{db: db, classifier: &PostgresErrorClassifier{}}
	metrics := []model.Metrics{
		{ID: "c1", MType: "counter", Delta: int64Ptr(10)},
		{ID: "g1", MType: "gauge", Value: float64Ptr(5.5)},
	}

	query := `INSERT INTO metrics \(id, mtype, delta, value, hash\) 
	          VALUES \(\$1, \$2, \$3, \$4, \$5\) 
	          ON CONFLICT \(id\) DO UPDATE SET
	              mtype = EXCLUDED\.mtype,
	              delta = COALESCE\(metrics\.delta, 0\) \+ EXCLUDED\.delta,
	              value = EXCLUDED\.value,
	              hash = EXCLUDED\.hash`

	mock.ExpectBegin()
	mock.ExpectExec(query).
		WithArgs("c1", "counter", int64(10), nil, "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(query).
		WithArgs("g1", "gauge", nil, float64(5.5), "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = store.SetMetrics(metrics)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_SetMetrics_TransactionFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PGStorage{db: db, classifier: &PostgresErrorClassifier{}}
	metrics := []model.Metrics{{ID: "fail", MType: "counter", Delta: int64Ptr(1)}}

	query := `INSERT INTO metrics \(id, mtype, delta, value, hash\) 
	          VALUES \(\$1, \$2, \$3, \$4, \$5\) 
	          ON CONFLICT \(id\) DO UPDATE SET
	              mtype = EXCLUDED\.mtype,
	              delta = COALESCE\(metrics\.delta, 0\) \+ EXCLUDED\.delta,
	              value = EXCLUDED\.value,
	              hash = EXCLUDED\.hash`

	mock.ExpectBegin()
	mock.ExpectExec(query).
		WithArgs("fail", "counter", int64(1), nil, "").
		WillReturnError(errors.New("tx fail"))
	mock.ExpectRollback()

	err = store.SetMetrics(metrics)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_IncrementCountMetricValue_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PGStorage{db: db, classifier: &PostgresErrorClassifier{}}
	delta := int64(100)

	mock.ExpectExec(`INSERT INTO metrics \(id, delta, mtype, hash\) VALUES \(\$2, \$1, 'counter', ''\) 
	                 ON CONFLICT \(id\) DO UPDATE
	                 SET delta = COALESCE\(metrics\.delta, 0\) \+ \$1
	                 WHERE metrics\.mtype = 'counter'`).
		WithArgs(delta, "counter1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = store.IncrementCountMetricValue("counter1", &delta)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_IncrementCountMetricValue_NilDelta(t *testing.T) {
	store := &PGStorage{db: nil, classifier: nil}

	err := store.IncrementCountMetricValue("counter", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delta is nil")
}

func TestPGStorage_getAttemptDelay(t *testing.T) {
	tests := []struct {
		name    string
		attempt int
		delay   time.Duration
	}{
		{"first", 0, 1 * time.Second},
		{"second", 1, 3 * time.Second},
		{"third", 2, 5 * time.Second},
		{"fourth", 3, 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.delay, getAttemptDelay(tt.attempt))
		})
	}
}

func TestPGStorage_Shutdown(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectClose()

	store := &PGStorage{db: db, classifier: &PostgresErrorClassifier{}}
	err = store.Shutdown()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_Ping_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	store := &PGStorage{db: db, classifier: &PostgresErrorClassifier{}}
	mock.ExpectPing()

	err = store.Ping()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPGStorage_Ping_Fail_WithRetry(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	classifier := NewPostgresErrorClassifier()
	store := &PGStorage{db: db, classifier: classifier}

	db.Close() // Close connection for error simulate

	err = store.Ping()

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
