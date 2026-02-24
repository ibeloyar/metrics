// Package handler provides HTTP handlers for metrics API using Chi router.
//
// Supports both query parameter and JSON endpoints for metrics.
// Implements HMAC-SHA256 request validation, HTML metrics page, and structured logging.
//
// Endpoints:
//
//	GET    /api/ping                    - health check
//	GET    /{type}/{name}               - get metric (query params)
//	POST   /api/v1/{type}/{name}/value  - update metric (query params)
//	POST   /value/                      - update metric (JSON + HMAC)
//	POST   /updates/                    - batch update (JSON + HMAC)
//	GET    /                            - HTML metrics table
package handler
