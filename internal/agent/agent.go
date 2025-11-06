package agent

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type Config struct {
	addr              string
	reportIntervalSec int
	pollIntervalSec   int
}

func Run() {
	var config Config
	var m runtime.MemStats
	var mu sync.Mutex

	flag.StringVar(&config.addr, "a", ":8080", "The address metric SERVER listen on")
	flag.IntVar(&config.reportIntervalSec, "r", 10, "Send report metrics interval")
	flag.IntVar(&config.pollIntervalSec, "p", 2, "Read metrics interval")

	flag.Parse()

	pollCount := 0

	go func() {
		client := &http.Client{
			Timeout: time.Second * 1,
		}

		for {
			time.Sleep(time.Duration(config.reportIntervalSec) * time.Second)

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
					fmt.Sprintf("http://%s/update/gauge/%s/%f", config.addr, name, value),
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

			response, err := client.Post(fmt.Sprintf("http://%s/update/counter/PollCount/%d", config.addr, pollCount), "text/plain", nil)
			if err != nil {
				panic(err)
			}
			response.Body.Close()

			response, err = client.Post(fmt.Sprintf("http://%s/update/gauge/RandomValue/%f", config.addr, rand.Float64()), "text/plain", nil)
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

		time.Sleep(time.Duration(config.pollIntervalSec) * time.Second)
	}
}
