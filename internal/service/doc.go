// Package service provides business logic layer for metrics operations
// with validation, audit logging, and storage abstraction.
//
// Implements REST API error responses and metric type validation.
// Supports both single and batch metric operations with audit events.
//
// Usage:
//
//	service := service.New(pgStorage, audit.New())
//	err := service.SetMetric(model.Metrics{ID: "cpu", MType: "gauge", Value: ptr(1.5)})
package service
