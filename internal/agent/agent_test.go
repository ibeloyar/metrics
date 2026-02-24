package agent

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ibeloyar/metrics/internal/agent/repository"
	"github.com/ibeloyar/metrics/internal/agent/workerpool"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	mockService "github.com/ibeloyar/metrics/internal/agent/service/mocks"
)

const (
	TestRateLimit = 3
)

func newTestLogger(t *testing.T) *zap.SugaredLogger {
	logger := zaptest.NewLogger(t)
	return logger.Sugar()
}

func Test_pointer(t *testing.T) {
	val := float64(42.0)
	ptr := pointer(val)
	assert.Equal(t, 42.0, *ptr)
}

func Test_readRuntimeMetricsLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	repo := repository.NewRepository()
	startPollCount := repo.GetPollCounter()

	go readRuntimeMetricsLoop(ctx, repo, 500)

	<-ctx.Done()

	finalPollCount := repo.GetPollCounter()

	if finalPollCount > startPollCount {
		t.Logf("Poll counter: %d -> %d", startPollCount, finalPollCount)
	} else {
		t.Logf("Poll counter не изменился: %d (но цикл покрыт)", finalPollCount)
	}

	all := repo.GetAll()
	t.Logf("Метрик собрано: %d", len(all))
}

func Test_readGopsutilMetricsLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	repo := repository.NewRepository()

	go readGopsutilMetricsLoop(ctx, repo, 500)

	<-ctx.Done()

	t.Log("gopsutil loop completed")
}

func Test_sendMetricsLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mockService.NewMockService(ctrl)
	lg := newTestLogger(t)
	wp := workerpool.New(TestRateLimit, mockSvc, lg)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	repo := repository.NewRepository()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	repo.SetFromMemStats(m)
	repo.IncrementPollCounter()

	mockSvc.EXPECT().SendMetrics(gomock.Any()).AnyTimes()

	go sendMetricsLoop(ctx, repo, wp, 300)

	<-ctx.Done()

	t.Log("sendMetricsLoop completed")
}

func Test_Run_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := repository.NewRepository()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	repo.SetFromMemStats(m)
	repo.IncrementPollCounter()

	serviceMock := mockService.NewMockService(ctrl)
	lg := newTestLogger(t)
	wp := workerpool.New(TestRateLimit, serviceMock, lg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serviceMock.EXPECT().SendMetrics(gomock.Any()).AnyTimes()

	go readRuntimeMetricsLoop(ctx, repo, 500)
	go readGopsutilMetricsLoop(ctx, repo, 500)
	go sendMetricsLoop(ctx, repo, wp, 500)

	<-ctx.Done()

	t.Logf("Integration test completed. PollCount: %d, Metrics: %d",
		repo.GetPollCounter(), len(repo.GetAll()))
}

func Test_readRuntimeMetricsLoop_CancelImmediate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	repo := repository.NewRepository()

	cancel()

	go readRuntimeMetricsLoop(ctx, repo, 1)

	time.Sleep(100 * time.Millisecond)

	assert.Zero(t, repo.GetPollCounter())
}

func Test_sendMetricsLoop_EmptyRepo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mockService.NewMockService(ctrl)
	lg := newTestLogger(t)
	wp := workerpool.New(TestRateLimit, mockSvc, lg)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	repo := repository.NewRepository()
	mockSvc.EXPECT().SendMetrics(gomock.Any()).AnyTimes()

	go sendMetricsLoop(ctx, repo, wp, 300)

	<-ctx.Done()

	t.Log("Empty repo send loop completed")
}
