// Package pgstorage provides PostgreSQL storage implementation for metrics
// with automatic schema migrations, connection retry logic, and ACID-compliant
// batch operations.
//
// The storage supports both gauge and counter metrics with upsert semantics.
// Counters accumulate delta values on conflicts. All operations include
// automatic connection retry with exponential backoff for transient failures.
//
// Usage:
//
//	storage, err := pgstorage.New("postgres://user:pass@localhost/db")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer storage.Shutdown()
//
//	// Single metric operations
//	err = storage.SetMetric(model.Metrics{ID: "cpu_load", MType: "gauge", Value: &value})
//	metric := storage.GetMetric("cpu_load")
//
//	// Batch operations with transactions
//	err = storage.SetMetrics([]model.Metrics{...})
package pgstorage
