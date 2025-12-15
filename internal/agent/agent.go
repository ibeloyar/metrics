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
	"github.com/ibeloyar/metrics/internal/logger"
)

func pointer[T any](v T) *T {
	return &v
}

func Run(config config.Config) error {
	var m runtime.MemStats

	lg, err := logger.New()
	if err != nil {
		return err
	}

	repo := repository.NewRepository()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	readMetricTicker := time.NewTicker(time.Duration(config.PollIntervalSec) * time.Second)
	defer readMetricTicker.Stop()

	sendMetricTicker := time.NewTicker(time.Duration(config.ReportIntervalSec) * time.Second)
	defer sendMetricTicker.Stop()

	lg.Info("Starting metric agent...")
	for {
		select {
		case <-readMetricTicker.C:
			runtime.ReadMemStats(&m)

			repo.SetFromMemStats(m)

			repo.IncrementPollCounter()
		case <-sendMetricTicker.C:
			var allMetrics []service.SendMetricBody

			as := service.NewService(config.Addr, config.Key)
			metrics := repo.GetAll()
			pollCounter := repo.GetPollCounter()

			for name, value := range metrics {
				allMetrics = append(allMetrics, service.SendMetricBody{
					ID:    name,
					MType: "gauge",
					Value: pointer(value),
				})
			}

			allMetrics = append(allMetrics, service.SendMetricBody{
				ID:    "PollCount",
				MType: "counter",
				Delta: pointer(pollCounter),
			})

			randomValue := rand.Float64()
			allMetrics = append(allMetrics, service.SendMetricBody{
				ID:    "RandomValue",
				MType: "gauge",
				Value: pointer(randomValue),
			})

			if err := as.SendMetrics(allMetrics); err != nil {
				lg.Error("Error sending metrics: ", err)
				return err
			}

			lg.Info("Metrics sent successfully")
		case <-ctx.Done():
			return nil
		}
	}
}
