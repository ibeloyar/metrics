package grpcservice

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	metricsv1 "github.com/ibeloyar/metrics/proto/metrics/v1"
)

type MetricsClient interface {
	UpdateMetrics(ctx context.Context, metrics []*metricsv1.Metric) (*metricsv1.UpdateMetricsResponse, error)
	Shutdown(ctx context.Context) error
}

type metricsClient struct {
	conn   *grpc.ClientConn
	client metricsv1.MetricsClient
}

func NewGRPCMetricsClient(addr string) (MetricsClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &metricsClient{
		conn:   conn,
		client: metricsv1.NewMetricsClient(conn),
	}, nil
}

func (c *metricsClient) UpdateMetrics(ctx context.Context, metrics []*metricsv1.Metric) (*metricsv1.UpdateMetricsResponse, error) {
	req := &metricsv1.UpdateMetricsRequest{}

	req.SetMetrics(metrics)

	return c.client.UpdateMetrics(ctx, req)
}

func (c *metricsClient) Shutdown(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		done <- c.conn.Close()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
