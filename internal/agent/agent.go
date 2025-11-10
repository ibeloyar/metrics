package agent

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	config "github.com/ibeloyar/metrics/internal/config/agent"
)

func Run(config config.Config) {
	var m runtime.MemStats

	safeMetrics := NewSafeMetrics()

	pollCount := 0

	go func() {
		client := &http.Client{
			Timeout: time.Second * 1,
		}

		for {
			time.Sleep(time.Duration(config.ReportIntervalSec) * time.Second)

			metrics := safeMetrics.GetAll()

			for name, value := range metrics {
				request, err := http.NewRequest(
					http.MethodPost,
					fmt.Sprintf("http://%s/update/gauge/%s/%f", config.Addr, name, value),
					nil,
				)
				if err != nil {
					panic(err)
				}

				request.Header.Set("Content-Type", "text/plain")

				response, err := client.Do(request)
				if err != nil {
					panic(err)
				}
				response.Body.Close()
			}

			response, err := client.Post(fmt.Sprintf("http://%s/update/counter/PollCount/%d", config.Addr, pollCount), "text/plain", nil)
			if err != nil {
				panic(err)
			}
			response.Body.Close()

			response, err = client.Post(fmt.Sprintf("http://%s/update/gauge/RandomValue/%f", config.Addr, rand.Float64()), "text/plain", nil)
			if err != nil {
				panic(err)
			}
			response.Body.Close()
		}
	}()

	for {
		runtime.ReadMemStats(&m)

		safeMetrics.SetFromMemStats(m)

		pollCount++

		time.Sleep(time.Duration(config.PollIntervalSec) * time.Second)
	}
}
