package workerpool

import (
	"sync"

	"github.com/ibeloyar/metrics/internal/agent/service"
	"go.uber.org/zap"
)

type WorkerPool struct {
	rateLimit int
	jobs      chan []service.SendMetricBody
	service   *service.Service
	lg        *zap.SugaredLogger
	wg        sync.WaitGroup
}

func New(rateLimit int, srv *service.Service, lg *zap.SugaredLogger) *WorkerPool {
	return &WorkerPool{
		rateLimit: rateLimit,
		jobs:      make(chan []service.SendMetricBody, rateLimit),
		service:   srv,
		lg:        lg,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.rateLimit; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPool) Dispatch(metrics []service.SendMetricBody) {
	select {
	case wp.jobs <- metrics:
	default:
		wp.lg.Warn("worker pool queue full, dropping metrics batch")
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	
	for metrics := range wp.jobs {
		if err := wp.service.SendMetrics(metrics); err != nil {
			wp.lg.Error("worker failed to send metrics: ", err)
		} else {
			wp.lg.Info("metrics batch sent by worker")
		}
	}
}

func (wp *WorkerPool) Shutdown() {
	close(wp.jobs)
	wp.wg.Wait()
}
