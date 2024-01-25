package spangaugemetricsconnector

// Config defines the configuration options for spangaugemetricsconnector.
type Config struct {
	Resource   ResourceConfig  `mapstructure:"resource"`
	Scope      ScopeConfig     `mapstructure:"scope"`
	Span       SpanConfig      `mapstructure:"span"`
	SpanEvents SpanEventConfig `mapstructure:"span_events"`
	Namespace  string          `mapstructure:"namespace"`
}

type ResourceConfig struct {
	Attributes []string `mapstructure:"attributes"`
}

type ScopeConfig struct {
	Attributes []string `mapstructure:"attributes"`
}

type SpanConfig struct {
	TraceID      bool     `mapstructure:"trace_id"`
	SpanID       bool     `mapstructure:"span_id"`
	ParentSpanID bool     `mapstructure:"parent_span_id"`
	StartTs      bool     `mapstructure:"start_ts"`
	EndTs        bool     `mapstructure:"end_ts"`
	Name         bool     `mapstructure:"name"`
	Attributes   []string `mapstructure:"attributes"`
}

type SpanEventConfig struct {
	TraceID    bool     `mapstructure:"trace_id"`
	SpanID     bool     `mapstructure:"span_id"`
	Name       bool     `mapstructure:"name"`
	Attributes []string `mapstructure:"attributes"`
}

func (c *Config) Validate() error {
	return nil
}
