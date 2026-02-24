package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() (*zap.SugaredLogger, error) {
	config := zap.NewDevelopmentConfig()

	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

func LoggingMiddleware(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &loggingResponseWriter{
				ResponseWriter: w,
				responseData: &responseData{
					status: http.StatusOK,
					size:   0,
				}}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Int("status", rw.responseData.status),
				zap.Int("size", rw.responseData.size),
				zap.Duration("duration", duration),
			)
		})
	}
}
