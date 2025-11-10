package agent

import (
	"runtime"
	"sync"
)

type SafeMetrics struct {
	metrics map[string]float64
	mu      sync.RWMutex
}

func NewSafeMetrics() *SafeMetrics {
	return &SafeMetrics{
		metrics: make(map[string]float64),
	}
}

func (sm *SafeMetrics) Get(key string) (float64, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	val, ok := sm.metrics[key]

	return val, ok
}

func (sm *SafeMetrics) GetAll() map[string]float64 {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	result := make(map[string]float64)
	for k, v := range sm.metrics {
		result[k] = v
	}

	return result
}

func (sm *SafeMetrics) set(key string, value float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.metrics[key] = value
}

func (sm *SafeMetrics) SetFromMemStats(stats runtime.MemStats) {
	sm.set("Alloc", float64(stats.Alloc))
	sm.set("BuckHashSys", float64(stats.BuckHashSys))
	sm.set("Frees", float64(stats.Frees))
	sm.set("GCCPUFraction", stats.GCCPUFraction)
	sm.set("GCSys", float64(stats.GCSys))
	sm.set("HeapAlloc", float64(stats.HeapAlloc))
	sm.set("HeapIdle", float64(stats.HeapIdle))
	sm.set("HeapInuse", float64(stats.HeapInuse))
	sm.set("HeapObjects", float64(stats.HeapObjects))
	sm.set("HeapReleased", float64(stats.HeapReleased))
	sm.set("HeapSys", float64(stats.HeapSys))
	sm.set("LastGC", float64(stats.LastGC))
	sm.set("Lookups", float64(stats.Lookups))
	sm.set("MCacheInuse", float64(stats.MCacheInuse))
	sm.set("MCacheSys", float64(stats.MCacheSys))
	sm.set("MSpanInuse", float64(stats.MSpanInuse))
	sm.set("MSpanSys", float64(stats.MSpanSys))
	sm.set("Mallocs", float64(stats.Mallocs))
	sm.set("NextGC", float64(stats.NextGC))
	sm.set("NumForcedGC", float64(stats.NumForcedGC))
	sm.set("NumGC", float64(stats.NumGC))
	sm.set("OtherSys", float64(stats.OtherSys))
	sm.set("PauseTotalNs", float64(stats.PauseTotalNs))
	sm.set("StackInuse", float64(stats.StackInuse))
	sm.set("StackSys", float64(stats.StackSys))
	sm.set("Sys", float64(stats.Sys))
	sm.set("TotalAlloc", float64(stats.TotalAlloc))
}
