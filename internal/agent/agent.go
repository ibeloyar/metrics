package agent

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	config "github.com/ibeloyar/metrics/internal/config/agent"
)

func Run(config config.Config) error {
	var m runtime.MemStats

	safeMetrics := NewSafeMetrics()

	pollCount := 0

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	readMetricTicker := time.NewTicker(time.Duration(config.PollIntervalSec) * time.Second)
	defer readMetricTicker.Stop()

	sendMetricTicker := time.NewTicker(time.Duration(config.ReportIntervalSec) * time.Second)
	defer sendMetricTicker.Stop()

	for {
		select {
		case <-readMetricTicker.C:
			// Сбор метрик
			runtime.ReadMemStats(&m)

			safeMetrics.SetFromMemStats(m)

			pollCount++
		case <-sendMetricTicker.C:
			// Отправка метрик
			client := &http.Client{
				Timeout: time.Second * 1,
			}

			metrics := safeMetrics.GetAll()

			for name, value := range metrics {
				request, err := http.NewRequest(
					http.MethodPost,
					fmt.Sprintf("http://%s/update/gauge/%s/%f", config.Addr, name, value),
					nil,
				)
				if err != nil {
					return err
				}

				request.Header.Set("Content-Type", "text/plain")

				response, err := client.Do(request)
				if err != nil {
					return err
				}
				response.Body.Close()
			}

			response, err := client.Post(fmt.Sprintf("http://%s/update/counter/PollCount/%d", config.Addr, pollCount), "text/plain", nil)
			if err != nil {
				return err
			}
			response.Body.Close()

			response, err = client.Post(fmt.Sprintf("http://%s/update/gauge/RandomValue/%f", config.Addr, rand.Float64()), "text/plain", nil)
			if err != nil {
				return err
			}
			response.Body.Close()
		case <-ctx.Done():
			return nil
		}
	}
}
