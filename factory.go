package spangaugemetricsconnector

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"
)

const (
	typeStr = "spangaugemetrics"
)

// NewFactory creates a factory for the spangaugemetrics connector.
func NewFactory() connector.Factory {
	return connector.NewFactory(
		typeStr,
		createDefaultConfig,
		connector.WithTracesToMetrics(createTracesToMetricsConnector, component.StabilityLevelAlpha),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Resource: ResourceConfig{
			Attributes: []string{
				"service.name",
			},
		},
		Scope: ScopeConfig{
			Attributes: make([]string, 0),
		},
		Span: SpanConfig{
			TraceID:      true,
			SpanID:       true,
			ParentSpanID: true,
			StartTs:      true,
			Name:         true,
		},
		SpanEvents: SpanEventConfig{
			TraceID:    true,
			SpanID:     true,
			Name:       true,
			Attributes: make([]string, 0),
		},
	}
}

func createTracesToMetricsConnector(ctx context.Context, params connector.CreateSettings, cfg component.Config, nextConsumer consumer.Metrics) (connector.Traces, error) {
	c, err := newConnector(params.Logger, cfg)
	if err != nil {
		return nil, err
	}
	c.metricsConsumer = nextConsumer
	return c, nil
}
