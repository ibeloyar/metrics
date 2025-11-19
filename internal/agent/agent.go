package agent

import (
	"context"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ibeloyar/metrics/internal/agent/config"
	"github.com/ibeloyar/metrics/internal/agent/repository"
	"github.com/ibeloyar/metrics/internal/agent/service"
)

func Run(config config.Config) error {
	var m runtime.MemStats

	repo := repository.NewRepository()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	readMetricTicker := time.NewTicker(time.Duration(config.PollIntervalSec) * time.Second)
	defer readMetricTicker.Stop()

	sendMetricTicker := time.NewTicker(time.Duration(config.ReportIntervalSec) * time.Second)
	defer sendMetricTicker.Stop()

	for {
		select {
		case <-readMetricTicker.C:
			runtime.ReadMemStats(&m)

			repo.SetFromMemStats(m)

			repo.IncrementPollCounter()
		case <-sendMetricTicker.C:
			as := service.NewService(config.Addr)
			metrics := repo.GetAll()
			pollCounter := repo.GetPollCounter()

			for name, value := range metrics {
				if err := as.SendGaugeMetric(name, value); err != nil {
					return err
				}
			}

			if err := as.SendPollCounter(pollCounter); err != nil {
				return err
			}
			repo.ResetPollCounter()

			if err := as.SendRandomValue(); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}
