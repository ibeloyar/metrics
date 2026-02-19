package repository

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"

	"github.com/shirou/gopsutil/v4/mem"
)

func TestRepository_Get_Set(t *testing.T) {
	r := NewRepository()

	r.set("test", 123.45)
	val, ok := r.Get("test")
	if !ok || val != 123.45 {
		t.Errorf("Get() failed: got (%f, %t) want (123.45, true)", val, ok)
	}
}

func TestRepository_GetAll(t *testing.T) {
	r := NewRepository()
	r.set("cpu", 0.75)
	r.set("memory", 85.5)

	all := r.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() wrong length: got %d want 2", len(all))
	}
	if all["cpu"] != 0.75 || all["memory"] != 85.5 {
		t.Errorf("GetAll() wrong values: %v", all)
	}
}

func TestRepository_PollCounter(t *testing.T) {
	r := NewRepository()

	r.IncrementPollCounter()
	r.IncrementPollCounter()

	if counter := r.GetPollCounter(); counter != 2 {
		t.Errorf("PollCounter wrong: got %d want 2", counter)
	}

	r.ResetPollCounter()
	if counter := r.GetPollCounter(); counter != 0 {
		t.Errorf("ResetPollCounter failed: got %d want 0", counter)
	}
}

func TestRepository_Concurrency(t *testing.T) {
	r := NewRepository()

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			r.set(fmt.Sprintf("key%d", id), float64(id))
		}(i)
	}
	wg.Wait()

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			_, ok := r.Get(fmt.Sprintf("key%d", id))
			if !ok {
				t.Errorf("Concurrent read failed for key%d", id)
			}
		}(i)
	}
	wg.Wait()
}

func TestRepository_SetFromMemStats(t *testing.T) {
	r := NewRepository()

	var stats runtime.MemStats
	stats.Alloc = 1024
	stats.HeapAlloc = 2048
	stats.NumGC = 5

	r.SetFromMemStats(stats)

	if val, ok := r.Get("Alloc"); !ok || val != 1024 {
		t.Errorf("SetFromMemStats Alloc failed: got %f want 1024", val)
	}
	if val, ok := r.Get("HeapAlloc"); !ok || val != 2048 {
		t.Errorf("SetFromMemStats HeapAlloc failed: got %f want 2048", val)
	}
	if val, ok := r.Get("NumGC"); !ok || val != 5 {
		t.Errorf("SetFromMemStats NumGC failed: got %f want 5", val)
	}
}

func TestRepository_SetGopsutilMetrics(t *testing.T) {
	r := NewRepository()

	memMetrics := &mem.VirtualMemoryStat{
		Total: 8589934592, // 8GB
		Free:  4294967296, // 4GB
	}
	cpuPercents := []float64{25.5, 30.2, 15.8}

	r.SetGopsutilMetrics(memMetrics, cpuPercents)

	if val, ok := r.Get("TotalMemory"); !ok || val != 8589934592.0 {
		t.Errorf("TotalMemory failed: got %f want 8589934592", val)
	}

	if val, ok := r.Get("CPU0"); !ok || val != 0.255 {
		t.Errorf("CPU0 failed: got %f want 0.255", val)
	}
	if val, ok := r.Get("CPU1"); !ok || val != 0.302 {
		t.Errorf("CPU1 failed: got %f want 0.302", val)
	}
}

func TestRepository_Get_NonExistent(t *testing.T) {
	r := NewRepository()

	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get() should return false for non-existent key")
	}
}

func TestRepository_MemStats_AllFields(t *testing.T) {
	r := NewRepository()

	var stats runtime.MemStats
	testValues := map[string]uint64{
		"Alloc":        1,
		"BuckHashSys":  2,
		"Frees":        3,
		"GCSys":        4,
		"HeapAlloc":    5,
		"HeapIdle":     6,
		"HeapInuse":    7,
		"HeapObjects":  8,
		"HeapReleased": 9,
		"HeapSys":      10,
		"LastGC":       11,
		"Lookups":      12,
		"MCacheInuse":  13,
		"MCacheSys":    14,
		"MSpanInuse":   15,
		"MSpanSys":     16,
		"Mallocs":      17,
		"NextGC":       18,
		"NumForcedGC":  19,
		"NumGC":        20,
		"OtherSys":     21,
		"PauseTotalNs": 22,
		"StackInuse":   23,
		"StackSys":     24,
		"Sys":          25,
		"TotalAlloc":   26,
	}

	for field, value := range testValues {
		statsField := reflect.ValueOf(&stats).Elem().FieldByName(field)
		if statsField.IsValid() && statsField.CanSet() {
			statsField.SetUint(value)
		}
	}

	r.SetFromMemStats(stats)

	for field, want := range testValues {
		got, ok := r.Get(field)
		if !ok || got != float64(want) {
			t.Errorf("%s: got %f (%t) want %d", field, got, ok, want)
		}
	}
}

func TestRepository_MultiplePollIncrements(t *testing.T) {
	r := NewRepository()

	for i := 0; i < 1000; i++ {
		r.IncrementPollCounter()
	}

	if counter := r.GetPollCounter(); counter != 1000 {
		t.Errorf("Multiple increments failed: got %d want 1000", counter)
	}
}
