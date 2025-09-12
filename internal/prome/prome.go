package prome

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PromClient struct {
	api v1.API
}

func NewPromClient(url string) (*PromClient, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, err
	}

	return &PromClient{
		api: v1.NewAPI(client),
	}, nil
}

func (c *PromClient) generalRequest(query string, start, end string) (model.Matrix, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}
	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %w", err)
	}

	var step time.Duration

	// 如果时间间隔在半小时以内，则步长设置为8秒；大于半小时则设置15秒间隔
	if endTime.Sub(startTime) <= 30*time.Minute {
		step = 8 * time.Second
	} else if endTime.Sub(startTime) <= 2*time.Hour {
		step = 15 * time.Second
	}
	result, _, err := c.api.QueryRange(ctx, query, v1.Range{Start: startTime, End: endTime, Step: step})
	if err != nil {
		return nil, err
	}
	res, ok := result.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
	return res, nil
}

func (c *PromClient) GetLatencyByTimeRange(start string, end string) (model.Matrix, error) {
	return c.generalRequest("node_network_latency_ms", start, end)
}

func (c *PromClient) GetSingleLatencyByTimeRange(node1, node2 string, start, end string) (model.Matrix, error) {
	q := fmt.Sprintf("node_network_latency_ms{src=~\"%s|%s\", dst=~\"%s|%s\"}",
		node1, node2, node1, node2)
	return c.generalRequest(q, start, end)
}
