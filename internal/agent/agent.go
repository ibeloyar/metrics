package agent

import (
	"context"
	"errors"
	"math/rand/v2"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ibeloyar/metrics/internal/agent/config"
	"github.com/ibeloyar/metrics/internal/agent/repository"
	"github.com/ibeloyar/metrics/internal/agent/service"
	"github.com/ibeloyar/metrics/internal/agent/workerpool"
	"github.com/ibeloyar/metrics/internal/logger"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

const ShutdownTimeout = 30 * time.Second

func pointer[T any](v T) *T {
	return &v
}

func Run(config config.Config) error {
	lg, err := logger.New()
	if err != nil {
		return err
	}

	repo := repository.NewRepository()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	s := service.NewService(config.Addr, config.Key, config.CryptoKey)
	wp := workerpool.New(config.RateLimit, s, lg)
	wp.Start()

	go readRuntimeMetricsLoop(ctx, repo, time.Duration(config.PollInterval)*time.Second)
	go readGopsutilMetricsLoop(ctx, repo, time.Duration(config.PollInterval)*time.Second)
	go sendMetricsLoop(ctx, repo, wp, time.Duration(config.ReportInterval)*time.Second)

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	wp.Shutdown()

	select {
	case <-shutdownCtx.Done():
		if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
			lg.Warn("agent shutdown timeout exceeded, forcing exit")
		} else {
			lg.Info("agent graceful shutdown completed")
		}
	}

	return nil
}

func readRuntimeMetricsLoop(ctx context.Context, repo *repository.Repository, pollInterval time.Duration) {
	var m runtime.MemStats

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(&m)
			repo.SetFromMemStats(m)
			repo.IncrementPollCounter()
		case <-ctx.Done():
			return
		}
	}
}

func readGopsutilMetricsLoop(ctx context.Context, repo *repository.Repository, pollInterval time.Duration) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			memStats, _ := mem.VirtualMemory()
			cpuPercents, _ := cpu.Percent(0, false)

			repo.SetGopsutilMetrics(memStats, cpuPercents)
		case <-ctx.Done():
			return
		}
	}
}

func sendMetricsLoop(ctx context.Context, repo *repository.Repository, wp *workerpool.WorkerPool, reportInterval time.Duration) {
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			allMetrics := make([]service.SendMetricBody, 0)

			for name, value := range repo.GetAll() {
				allMetrics = append(allMetrics, service.SendMetricBody{
					ID:    name,
					MType: "gauge",
					Value: pointer(value),
				})
			}

			allMetrics = append(allMetrics, service.SendMetricBody{
				ID:    "PollCount",
				MType: "counter",
				Delta: pointer(repo.GetPollCounter()),
			})

			randomValue := rand.Float64()
			allMetrics = append(allMetrics, service.SendMetricBody{
				ID:    "RandomValue",
				MType: "gauge",
				Value: pointer(randomValue),
			})

			wp.Dispatch(allMetrics)
		case <-ctx.Done():
			return
		}
	}
}
