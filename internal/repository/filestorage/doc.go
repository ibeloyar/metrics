// Package filestorage provides JSON file-based persistence for metrics.
//
// Stores metrics map as pretty-printed JSON with automatic directory creation.
// Implements memstorage.FileStorage interface for seamless integration.
//
// File permissions: 0644 (owner: rw, group/other: r).
// Directory permissions: 0777 (all access) for automatic creation.
//
// Usage:
//
//	fs := filestorage.New("/var/lib/metrics/metrics.json")
//	storage := memstorage.New(fs, 30, true)  // save every 30s, restore on start
package filestorage
