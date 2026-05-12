package server

import (
	"context"
	"errors"
	"net"

	"github.com/ibeloyar/metrics/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	metricsv1 "github.com/ibeloyar/metrics/proto/metrics/v1"
)

type Storage interface {
	SetMetrics(metrics []model.Metrics) error
}

type MetricsGRPCController struct {
	metricsv1.UnimplementedMetricsServer

	lg            *zap.SugaredLogger
	storage       Storage
	trustedSubnet *net.IPNet
}

func NewMetricsGRPCController(lg *zap.SugaredLogger, storage Storage, trustedSubnet string) *MetricsGRPCController {
	var subnet *net.IPNet

	if trustedSubnet != "" {
		_, parsedSubnet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			lg.Fatalf("invalid trusted_subnet %q: %v", trustedSubnet, err)
		}
		subnet = parsedSubnet
	}

	return &MetricsGRPCController{
		lg:            lg,
		storage:       storage,
		trustedSubnet: subnet,
	}
}

func (mc *MetricsGRPCController) HandlePanic(p any) error {
	if p != nil {
		mc.lg.Errorf("panic: %v", p)
	}
	return errors.New("grpc internal error")
}

func (mc *MetricsGRPCController) UpdateMetrics(_ context.Context, req *metricsv1.UpdateMetricsRequest) (*metricsv1.UpdateMetricsResponse, error) {
	metrics := make([]model.Metrics, 0)

	for _, metric := range req.GetMetrics() {
		newMetric := model.Metrics{}

		newMetric.ID = metric.GetId()
		if metric.GetType() == metricsv1.Metric_COUNTER {
			newMetric.MType = model.Counter
		}
		if metric.GetType() == metricsv1.Metric_GAUGE {
			newMetric.MType = model.Gauge
		}
		value := metric.GetValue()
		newMetric.Value = &value

		delta := metric.GetDelta()
		newMetric.Delta = &delta

		metrics = append(metrics, newMetric)
	}

	if err := mc.storage.SetMetrics(metrics); err != nil {
		mc.lg.Errorf("error updating metrics: %v", err)

		return nil, err
	}

	return &metricsv1.UpdateMetricsResponse{}, nil
}

func (mc *MetricsGRPCController) TrustedNetsUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if mc.trustedSubnet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			mc.lg.Warn("no metadata in context")
			return nil, status.Errorf(codes.PermissionDenied, "no metadata")
		}

		ips := md.Get("x-real-ip")
		if len(ips) == 0 {
			mc.lg.Warn("x-real-ip header missing")
			return nil, status.Errorf(codes.PermissionDenied, "x-real-ip required")
		}

		host := ips[0]
		clientIP := net.ParseIP(host)
		if clientIP == nil {
			mc.lg.Warnf("invalid x-real-ip %q", host)
			return nil, status.Errorf(codes.PermissionDenied, "invalid x-real-ip")
		}

		if !mc.trustedSubnet.Contains(clientIP) {
			mc.lg.Warnf("client IP %q not in trusted subnet %q", clientIP, mc.trustedSubnet.String())
			return nil, status.Errorf(codes.PermissionDenied, "IP not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
