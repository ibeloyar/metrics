// Package memstorage provides in-memory metrics storage with optional periodic
// file persistence and restore capabilities.
//
// Supports gauge and counter metrics with type-safe storage. Automatically
// restores metrics from file on startup if configured. Saves metrics to file
// either periodically (via ticker) or synchronously on each write if no ticker.
//
// Usage:
//
//	storage := memstorage.New(jsonFileStorage, 30, true)  // 30s interval, restore on start
//	storage.Init()
//	defer storage.Shutdown()
//
//	// Single operations
//	storage.SetMetric(model.Metrics{ID: "cpu", MType: "gauge", Value: ptr(1.5)})
//
//	// Batch operations with counter accumulation
//	storage.SetMetrics([]model.Metrics{{ID: "requests", MType: "counter", Delta: ptr(100)}})
package memstorage
