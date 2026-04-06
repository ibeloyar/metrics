package server

import (
	"context"
	"errors"
	"net"

	"github.com/ibeloyar/metrics/internal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
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

		p, ok := peer.FromContext(ctx)
		if !ok || p.Addr == nil {
			mc.lg.Warn("no peer info in context")
			return nil, status.Errorf(codes.PermissionDenied, "no peer info")
		}

		ipStr := p.Addr.String()
		host, _, err := net.SplitHostPort(ipStr)
		if err != nil {
			mc.lg.Warnf("invalid peer addr %q: %v", ipStr, err)
			return nil, status.Errorf(codes.PermissionDenied, "invalid peer addr")
		}

		clientIP := net.ParseIP(host)
		if clientIP == nil {
			mc.lg.Warnf("invalid client IP %q", host)
			return nil, status.Errorf(codes.PermissionDenied, "invalid client IP")
		}

		if !mc.trustedSubnet.Contains(clientIP) {
			mc.lg.Warnf("client IP %q not in trusted subnet %q", clientIP, mc.trustedSubnet.String())
			return nil, status.Errorf(codes.PermissionDenied, "IP not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
