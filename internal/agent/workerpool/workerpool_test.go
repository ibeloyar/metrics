package workerpool

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ibeloyar/metrics/internal/agent/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"

	mockService "github.com/ibeloyar/metrics/internal/agent/service/mocks"
)

func TestNewWorkerPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mockService.NewMockService(ctrl)
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	pool := New(3, mockSvc, logger.Sugar())

	assert.Equal(t, 3, pool.rateLimit)
	assert.Equal(t, 3, cap(pool.jobs))
	assert.Equal(t, mockSvc, pool.service)
	assert.Equal(t, logger.Sugar(), pool.lg)
}

func TestDispatch_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	pool := New(2, nil, logger.Sugar())

	metrics := []service.SendMetricBody{{ID: "test", MType: "gauge"}}
	pool.Dispatch(metrics)

	select {
	case got := <-pool.jobs:
		assert.Equal(t, metrics, got)
	default:
		t.Fatal("metrics should be dispatched")
	}
}

func TestDispatch_FullQueue(t *testing.T) {
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	pool := New(1, nil, logger.Sugar())

	pool.jobs <- []service.SendMetricBody{{}}

	dropped := []service.SendMetricBody{{ID: "dropped"}}
	pool.Dispatch(dropped)

	assert.Equal(t, 1, len(pool.jobs))
}

func TestWorker_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mockService.NewMockService(ctrl)
	mockSvc.EXPECT().SendMetrics(gomock.Any()).Return(nil).Times(3)

	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	pool := New(3, mockSvc, logger.Sugar())
	pool.Start()
	defer pool.Shutdown()

	for i := 0; i < 3; i++ {
		pool.Dispatch([]service.SendMetricBody{{ID: fmt.Sprintf("test-%d", i)}})
	}

	time.Sleep(100 * time.Millisecond)
}

func TestWorker_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mockService.NewMockService(ctrl)
	mockSvc.EXPECT().SendMetrics(gomock.Any()).Return(fmt.Errorf("send failed")).Times(2)

	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	pool := New(2, mockSvc, logger.Sugar())
	pool.Start()
	defer pool.Shutdown()

	for i := 0; i < 2; i++ {
		pool.Dispatch([]service.SendMetricBody{{ID: fmt.Sprintf("error-%d", i)}})
	}

	time.Sleep(100 * time.Millisecond)
}

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mockService.NewMockService(ctrl)
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	pool := New(2, mockSvc, logger.Sugar())

	pool.Start()

	pool.Shutdown()
}

func TestShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mockService.NewMockService(ctrl)
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	pool := New(2, mockSvc, logger.Sugar())
	pool.Start()

	pool.Shutdown()

	_, ok := <-pool.jobs
	assert.False(t, ok, "channel should be closed")
}

func TestRateLimitZero(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mockService.NewMockService(ctrl)
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	pool := New(0, mockSvc, logger.Sugar())

	assert.Equal(t, 0, pool.rateLimit)
	assert.Equal(t, 0, cap(pool.jobs))

	pool.Start()
	pool.Dispatch([]service.SendMetricBody{{ID: "test"}})

	pool.Shutdown()
}
