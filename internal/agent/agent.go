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
	"github.com/ibeloyar/metrics/internal/agent/grpcservice"
	"github.com/ibeloyar/metrics/internal/agent/repository"
	"github.com/ibeloyar/metrics/internal/agent/service"
	"github.com/ibeloyar/metrics/internal/agent/workerpool"
	"github.com/ibeloyar/metrics/internal/logger"
	"github.com/ibeloyar/metrics/pkg/netlib"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	metricsv1 "github.com/ibeloyar/metrics/proto/metrics/v1"
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

	var grpcClient grpcservice.MetricsClient
	if config.GRPCAddr != "" {
		grpcClient, err = grpcservice.NewGRPCMetricsClient(config.GRPCAddr)
		if err != nil {
			lg.Error("failed to create grpc client", zap.Error(err))
		}

		go sendMetricsGRPC(ctx, repo, grpcClient, lg, time.Duration(config.ReportInterval)*time.Second)
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)

		wp.Shutdown()

		if config.GRPCAddr != "" {
			if shutdownErr := grpcClient.Shutdown(shutdownCtx); shutdownErr != nil {
				lg.Error("failed to shutdown grpc client", zap.Error(shutdownErr))
			}
		}
	}()

	select {
	case <-shutdownCtx.Done():
		if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
			lg.Warn("agent shutdown timeout exceeded, forcing exit")
		}
	case <-done:
		lg.Info("agent graceful shutdown completed")
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

func getAllMetrics(repo *repository.Repository) []service.SendMetricBody {
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

	return allMetrics
}

func sendMetricsLoop(ctx context.Context, repo *repository.Repository, wp *workerpool.WorkerPool, reportInterval time.Duration) {
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			allMetrics := getAllMetrics(repo)

			wp.Dispatch(allMetrics)
		case <-ctx.Done():
			return
		}
	}
}

func sendMetricsGRPC(ctx context.Context, repo *repository.Repository, client grpcservice.MetricsClient, lg *zap.SugaredLogger, reportInterval time.Duration) {
	updateMetricsCtx := ctx

	if localIP := netlib.GetOutboundIP(); localIP != "" {
		md := metadata.Pairs("x-real-ip", localIP)
		updateMetricsCtx = metadata.NewOutgoingContext(ctx, md)
	}

	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			allMetrics := getAllMetrics(repo)
			grpcMetrics := make([]*metricsv1.Metric, 0)

			for _, metric := range allMetrics {
				grpcMetric := &metricsv1.Metric{}

				grpcMetric.SetId(metric.ID)
				if metric.MType == "counter" {
					grpcMetric.SetType(metricsv1.Metric_COUNTER)
				}
				if metric.MType == "gauge" {
					grpcMetric.SetType(metricsv1.Metric_GAUGE)
				}
				if metric.Delta != nil {
					grpcMetric.SetDelta(*metric.Delta)
				}
				if metric.Value != nil {
					grpcMetric.SetValue(*metric.Value)
				}

				grpcMetrics = append(grpcMetrics, grpcMetric)
			}

			if _, err := client.UpdateMetrics(updateMetricsCtx, grpcMetrics); err != nil {
				lg.Error("sending grpc metrics failed", zap.Error(err))
			}

		case <-ctx.Done():
			return
		}
	}
}
