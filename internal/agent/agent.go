package agent

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func Run() {
	var m runtime.MemStats
	var mu sync.Mutex

	pollCount := 0

	go func() {
		client := &http.Client{
			Timeout: time.Second * 1,
		}

		for {
			time.Sleep(reportInterval)

			mu.Lock()
			metrics := map[string]float64{
				"Alloc":         float64(m.Alloc),
				"BuckHashSys":   float64(m.BuckHashSys),
				"Frees":         float64(m.Frees),
				"GCCPUFraction": m.GCCPUFraction,
				"GCSys":         float64(m.GCSys),
				"HeapAlloc":     float64(m.HeapAlloc),
				"HeapIdle":      float64(m.HeapIdle),
				"HeapInuse":     float64(m.HeapInuse),
				"HeapObjects":   float64(m.HeapObjects),
				"HeapReleased":  float64(m.HeapReleased),
				"HeapSys":       float64(m.HeapSys),
				"LastGC":        float64(m.LastGC),
				"Lookups":       float64(m.Lookups),
				"MCacheInuse":   float64(m.MCacheInuse),
				"MCacheSys":     float64(m.MCacheSys),
				"MSpanInuse":    float64(m.MSpanInuse),
				"MSpanSys":      float64(m.MSpanSys),
				"Mallocs":       float64(m.Mallocs),
				"NextGC":        float64(m.NextGC),
				"NumForcedGC":   float64(m.NumForcedGC),
				"NumGC":         float64(m.NumGC),
				"OtherSys":      float64(m.OtherSys),
				"PauseTotalNs":  float64(m.PauseTotalNs),
				"StackInuse":    float64(m.StackInuse),
				"StackSys":      float64(m.StackSys),
				"Sys":           float64(m.Sys),
				"TotalAlloc":    float64(m.TotalAlloc),
			}
			mu.Unlock()

			for name, value := range metrics {
				request, err := http.NewRequest(
					http.MethodPost,
					fmt.Sprintf("http://localhost:8080/update/gauge/%s/%f", name, value),
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

			response, err := client.Post(fmt.Sprintf("http://localhost:8080/update/counter/PollCount/%d", pollCount), "text/plain", nil)
			if err != nil {
				panic(err)
			}
			response.Body.Close()

			response, err = client.Post(fmt.Sprintf("http://localhost:8080/update/gauge/RandomValue/%f", rand.Float64()), "text/plain", nil)
			if err != nil {
				panic(err)
			}
			response.Body.Close()
		}
	}()

	for {
		mu.Lock()
		runtime.ReadMemStats(&m)
		mu.Unlock()

		pollCount++

		time.Sleep(pollInterval)
	}
}
