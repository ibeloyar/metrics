package repository

import (
	"runtime"
	"sync"
)

type Repository struct {
	metrics     map[string]float64
	mu          sync.RWMutex
	pollCounter int
}

func NewRepository() *Repository {
	return &Repository{
		metrics:     make(map[string]float64),
		pollCounter: 0,
	}
}

func (r *Repository) Get(key string) (float64, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	val, ok := r.metrics[key]

	return val, ok
}

func (r *Repository) GetAll() map[string]float64 {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make(map[string]float64)
	for k, v := range r.metrics {
		result[k] = v
	}

	return result
}

func (r *Repository) set(key string, value float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.metrics[key] = value
}

func (r *Repository) SetFromMemStats(stats runtime.MemStats) {
	r.set("Alloc", float64(stats.Alloc))
	r.set("BuckHashSys", float64(stats.BuckHashSys))
	r.set("Frees", float64(stats.Frees))
	r.set("GCCPUFraction", stats.GCCPUFraction)
	r.set("GCSys", float64(stats.GCSys))
	r.set("HeapAlloc", float64(stats.HeapAlloc))
	r.set("HeapIdle", float64(stats.HeapIdle))
	r.set("HeapInuse", float64(stats.HeapInuse))
	r.set("HeapObjects", float64(stats.HeapObjects))
	r.set("HeapReleased", float64(stats.HeapReleased))
	r.set("HeapSys", float64(stats.HeapSys))
	r.set("LastGC", float64(stats.LastGC))
	r.set("Lookups", float64(stats.Lookups))
	r.set("MCacheInuse", float64(stats.MCacheInuse))
	r.set("MCacheSys", float64(stats.MCacheSys))
	r.set("MSpanInuse", float64(stats.MSpanInuse))
	r.set("MSpanSys", float64(stats.MSpanSys))
	r.set("Mallocs", float64(stats.Mallocs))
	r.set("NextGC", float64(stats.NextGC))
	r.set("NumForcedGC", float64(stats.NumForcedGC))
	r.set("NumGC", float64(stats.NumGC))
	r.set("OtherSys", float64(stats.OtherSys))
	r.set("PauseTotalNs", float64(stats.PauseTotalNs))
	r.set("StackInuse", float64(stats.StackInuse))
	r.set("StackSys", float64(stats.StackSys))
	r.set("Sys", float64(stats.Sys))
	r.set("TotalAlloc", float64(stats.TotalAlloc))
}

func (r *Repository) GetPollCounter() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.pollCounter
}

func (r *Repository) IncrementPollCounter() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pollCounter++
}

func (r *Repository) ResetPollCounter() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pollCounter = 0
}
