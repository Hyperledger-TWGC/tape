package basic

import (
	"fmt"
	"io"
	"time"

	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func LogEvent(logger *log.Logger, txid, event string) {
	now := time.Now()
	time_str := fmt.Sprintf(now.Format(time.RFC3339Nano))
	logger.Debugf("For txid %s, event %s at %s", txid, event, time_str)
}

// Init returns an instance of Jaeger Tracer that samples 100%
// of traces and logs all spans to stdout.
func Init(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.New(
		service,
		config.Logger(jaeger.StdLogger),
	)
	if err != nil {
		log.Fatalf("ERROR: cannot init Jaeger: %v", err)
	}
	return tracer, closer
}
