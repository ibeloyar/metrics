package agent

import (
	"context"
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

//1. Перепланируйте архитектуру агента таким образом, чтобы сбор метрик (опрос runtime) и их отправка осуществлялись в разных горутинах.
//2. При этом количество одновременно исходящих запросов на сервер нужно ограничивать «сверху».
//   Соответствующее значение должно задаваться аргументами: Через флаг -l=<ЗНАЧЕНИЕ> и переменную окружения RATE_LIMIT.
//   Совет: Используйте паттерн worker pool.
//
//3. Добавьте ещё одну горутину, которая будет использовать пакет gopsutil и собирать дополнительные метрики типа gauge:
//   TotalMemory,
//   FreeMemory,
//   CPUutilization1 (точное количество — по числу CPU, определяемому во время исполнения)

func pointer[T any](v T) *T {
	return &v
}

func Run(config config.Config) error {
	lg, err := logger.New()
	if err != nil {
		return err
	}

	repo := repository.NewRepository()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s := service.NewService(config.Addr, config.Key)
	wp := workerpool.New(config.RateLimit, s, lg)
	wp.Start()

	go readRuntimeMetricsLoop(ctx, repo, config.PollIntervalSec)
	go readGopsutilMetricsLoop(ctx, repo, config.PollIntervalSec)
	go sendMetricsLoop(ctx, repo, wp, config.ReportIntervalSec)

	<-ctx.Done()

	wp.Shutdown()
	lg.Info("Agent shutdown")

	return nil
}

func readRuntimeMetricsLoop(ctx context.Context, repo *repository.Repository, pollIntervalSec int) {
	var m runtime.MemStats

	ticker := time.NewTicker(time.Duration(pollIntervalSec) * time.Second)
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

func readGopsutilMetricsLoop(ctx context.Context, repo *repository.Repository, pollIntervalSec int) {
	ticker := time.NewTicker(time.Duration(pollIntervalSec) * time.Second)
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

func sendMetricsLoop(ctx context.Context, repo *repository.Repository, wp *workerpool.WorkerPool, reportIntervalSec int) {
	ticker := time.NewTicker(time.Duration(reportIntervalSec) * time.Second)
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
