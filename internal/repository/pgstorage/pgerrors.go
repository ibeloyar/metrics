package pgstorage

import (
	"errors"

	"github.com/lib/pq"
)

// PGErrorClassification classifies PostgreSQL errors as retriable or non-retriable.
type PGErrorClassification int

const (
	// NonRetriable errors should not be retried (syntax errors, constraint violations).
	NonRetriable PGErrorClassification = iota
	// Retriable errors can be safely retried (connection issues, transaction rollbacks).
	Retriable
)

// PostgresErrorClassifier classifies PostgreSQL errors for retry logic.
type PostgresErrorClassifier struct{}

// NewPostgresErrorClassifier creates a new error classifier instance.
func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// Classify determines if an error is retriable.
//
// Uses errors.As to unwrap pq.Error and checks PostgreSQL error codes:
// [PostgreSQL Error Codes](https://www.postgresql.org/docs/current/errcodes-appendix.html)
//
// Returns Retriable for:
//   - Connection errors (08XXX)
//   - Transaction rollback (40XXX)
//   - Operator intervention (57P03)
//
// Returns NonRetriable for:
//   - Data errors (22XXX)
//   - Constraint violations (23XXX)
//   - Syntax errors (42XXX)
func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}

	// Check pq.Error using errors.As
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return classifyPgError(pqErr)
	}

	// Non-pq errors are non-retriable by default
	return NonRetriable
}

// classifyPgError classifies specific PostgreSQL error codes.
//
// Connection errors (08000), transaction rollbacks (40000), and operator intervention (57P03)
// are retriable. Data, constraint, and syntax errors are non-retriable.
//
// [PostgreSQL Error Codes](https://www.postgresql.org/docs/current/errcodes-appendix.html)
func classifyPgError(pqErr *pq.Error) PGErrorClassification {
	switch pqErr.Code {
	// Class 08 - Connection errors
	case "08000", "08001", "08003", "08004", "08006", "08007":
		return Retriable

	// Class 40 - Transaction rollback
	case "40000", "40001", "40P01":
		return Retriable

	// Class 57 - Operator intervention
	case "57P03":
		return Retriable
	}

	// Class 22 - Data errors
	switch pqErr.Code {
	case "22000", "22004":
		return NonRetriable
	}

	// Class 23 - Integrity constraint violations
	switch pqErr.Code {
	case "23000", "23001", "23502", "23503", "23505", "23514":
		return NonRetriable
	}

	// Class 42 - Syntax errors and access rule violations
	switch pqErr.Code {
	case "42601", "42P01", "42703", "42P02", "42P03":
		return NonRetriable
	}

	// Default: non-retriable
	return NonRetriable
}
