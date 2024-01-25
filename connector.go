package spangaugemetricsconnector

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

type connectorImp struct {
	config              Config
	resourceAttributes  map[string]bool
	scopeAttributes     map[string]bool
	spanAttributes      map[string]bool
	spanEventAttributes map[string]bool

	metricsConsumer consumer.Metrics

	logger *zap.Logger

	component.StartFunc
	component.ShutdownFunc
}

func newConnector(logger *zap.Logger, config component.Config) (*connectorImp, error) {
	logger.Info("Building spangaugemetrics connector")
	cfg := config.(*Config)

	resourceAttributes := make(map[string]bool)
	for _, attr := range cfg.Resource.Attributes {
		resourceAttributes[attr] = true
	}

	scopeAttributes := make(map[string]bool)
	for _, attr := range cfg.Scope.Attributes {
		scopeAttributes[attr] = true
	}

	spanAttributes := make(map[string]bool)
	for _, attr := range cfg.Span.Attributes {
		spanAttributes[attr] = true
	}
	if cfg.Span.TraceID {
		spanAttributes["span_trace_id"] = true
	}
	if cfg.Span.SpanID {
		spanAttributes["span_id"] = true
	}
	if cfg.Span.ParentSpanID {
		spanAttributes["span_parent_id"] = true
	}
	if cfg.Span.StartTs {
		spanAttributes["span_start_ts"] = true
	}
	if cfg.Span.EndTs {
		spanAttributes["span_end_ts"] = true
	}
	if cfg.Span.Name {
		spanAttributes["span_name"] = true
	}

	spanEventAttributes := make(map[string]bool)
	for _, attr := range cfg.SpanEvents.Attributes {
		spanEventAttributes[attr] = true
	}
	if cfg.SpanEvents.TraceID {
		spanEventAttributes["span_trace_id"] = true
	}
	if cfg.SpanEvents.SpanID {
		spanEventAttributes["span_id"] = true
	}
	if cfg.SpanEvents.Name {
		spanEventAttributes["span_event_name"] = true
	}

	return &connectorImp{
		config:              *cfg,
		resourceAttributes:  resourceAttributes,
		scopeAttributes:     scopeAttributes,
		spanAttributes:      spanAttributes,
		spanEventAttributes: spanEventAttributes,

		logger: logger,
	}, nil
}

// Capabilities implements the consumer interface.
func (c *connectorImp) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// ConsumeTraces implements the consumer.Traces interface.
func (c *connectorImp) ConsumeTraces(ctx context.Context, traces ptrace.Traces) error {
	m := pmetric.NewMetrics()

	for i := 0; i < traces.ResourceSpans().Len(); i++ {
		resourceSpan := traces.ResourceSpans().At(i)

		rm := m.ResourceMetrics().AppendEmpty()
		extractAttributes(resourceSpan.Resource().Attributes(), c.resourceAttributes).
			CopyTo(rm.Resource().Attributes())

		for j := 0; j < resourceSpan.ScopeSpans().Len(); j++ {
			scopeSpan := resourceSpan.ScopeSpans().At(j)

			sm := rm.ScopeMetrics().AppendEmpty()
			sm.Scope().SetName("spangaugemetricsconnector")
			extractAttributes(scopeSpan.Scope().Attributes(), c.scopeAttributes).
				CopyTo(sm.Scope().Attributes())

			for k := 0; k < scopeSpan.Spans().Len(); k++ {
				span := scopeSpan.Spans().At(k)

				spm := sm.Metrics().AppendEmpty()
				spm.SetName(c.config.Namespace + ".span_duration")
				spm.SetUnit("ms")

				dp := spm.SetEmptyGauge().DataPoints().AppendEmpty()
				dp.SetDoubleValue(float64(span.EndTimestamp().AsTime().Sub(span.StartTimestamp().AsTime()).Nanoseconds()) / float64(time.Millisecond))

				spanAttr := span.Attributes()
				if c.config.Span.TraceID {
					spanAttr.PutStr("span_trace_id", span.TraceID().String())
				}
				if c.config.Span.SpanID {
					spanAttr.PutStr("span_id", span.SpanID().String())
				}
				if c.config.Span.ParentSpanID {
					spanAttr.PutStr("span_parent_id", span.ParentSpanID().String())
				}
				if c.config.Span.StartTs {
					spanAttr.PutInt("span_start_ts", span.StartTimestamp().AsTime().UnixMilli())
				}
				if c.config.Span.EndTs {
					spanAttr.PutInt("span_end_ts", span.EndTimestamp().AsTime().UnixMilli())
				}
				if c.config.Span.Name {
					spanAttr.PutStr("span_name", span.Name())
				}

				extractAttributes(spanAttr, c.spanAttributes).CopyTo(dp.Attributes())

				for l := 0; l < span.Events().Len(); l++ {
					spanEvent := span.Events().At(l)

					spm := sm.Metrics().AppendEmpty()
					spm.SetName(c.config.Namespace + ".span_event")

					dp := spm.SetEmptyGauge().DataPoints().AppendEmpty()
					dp.SetIntValue(1)

					spanEventAttr := spanEvent.Attributes()
					if c.config.SpanEvents.TraceID {
						spanEventAttr.PutStr("span_trace_id", span.TraceID().String())
					}
					if c.config.SpanEvents.SpanID {
						spanEventAttr.PutStr("span_id", span.SpanID().String())
					}
					if c.config.SpanEvents.Name {
						spanEventAttr.PutStr("span_event_name", spanEvent.Name())
					}

					extractAttributes(spanEventAttr, c.spanEventAttributes).CopyTo(dp.Attributes())
				}
			}
		}
	}

	if err := c.metricsConsumer.ConsumeMetrics(ctx, m); err != nil {
		c.logger.Error("Failed ConsumeMetrics", zap.Error(err))
	}

	return nil
}

func extractAttributes(data pcommon.Map, cfg map[string]bool) pcommon.Map {
	output := pcommon.NewMap()
	data.CopyTo(output)

	output.RemoveIf(func(k string, v pcommon.Value) bool {
		return !cfg[k]
	})

	return output
}
