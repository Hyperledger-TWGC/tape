package basic

import (
	"fmt"
	"io"
	"sync"
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
			LogSpans: false,
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

const (
	TRANSCATION         = "TRANSCATION"
	TRANSCATIONSTART    = "TRANSCATIONSTART"
	SIGN_PROPOSAL       = "SIGN_PROPOSAL"
	ENDORSEMENT         = "ENDORSEMENT"
	ENDORSEMENT_AT_PEER = "ENDORSEMENT_AT_PEER"
	COLLECT_ENDORSEMENT = "COLLECT_ENDORSEMENT"
	SIGN_ENVELOP        = "SIGN_ENVELOP"
	BROADCAST           = "BROADCAST"
	CONSESUS            = "CONSESUS"
	COMMIT_AT_NETWORK   = "COMMIT_AT_NETWORK"
	COMMIT_AT_PEER      = "COMMIT_AT_PEER"
	COMMIT_AT_ALL_PEERS = "COMMIT_AT_ALL_PEERS"
)

var TapeSpan *TracingSpans
var ProcessMod int

type TracingSpans struct {
	Spans map[string]opentracing.Span
	Lock  sync.Mutex
}

func (TS *TracingSpans) MakeSpan(txid, address, event string, parent opentracing.Span) opentracing.Span {
	str := fmt.Sprintf(event + address)
	if parent == nil {
		return opentracing.GlobalTracer().StartSpan(str, opentracing.Tag{Key: "txid", Value: txid})
	} else {
		return opentracing.GlobalTracer().StartSpan(str, opentracing.ChildOf(parent.Context()), opentracing.Tag{Key: "txid", Value: txid})
	}
}

func (TS *TracingSpans) GetSpan(txid, address, event string) opentracing.Span {
	TS.Lock.Lock()
	defer TS.Lock.Unlock()

	str := fmt.Sprintf(event + txid + address)
	span, ok := TS.Spans[str]
	if ok {
		return span
	}
	return nil
}

func (TS *TracingSpans) SpanIntoMap(txid, address, event string, parent opentracing.Span) opentracing.Span {
	TS.Lock.Lock()
	defer TS.Lock.Unlock()

	str := fmt.Sprintf(event + txid + address)
	span, ok := TS.Spans[str]
	if !ok {
		span = TS.MakeSpan(txid, address, event, parent)
		TS.Spans[str] = span
	}
	return span
}

func (TS *TracingSpans) FinishWithMap(txid, address, event string) {
	TS.Lock.Lock()
	defer TS.Lock.Unlock()

	str := fmt.Sprintf(event + txid + address)
	span, ok := TS.Spans[str]
	if ok {
		span.Finish()
		delete(TS.Spans, str)
	}
}

func GetGlobalSpan() *TracingSpans {
	return TapeSpan
}

func InitSpan() *TracingSpans {
	Spans := make(map[string]opentracing.Span)

	TapeSpan = &TracingSpans{
		Spans: Spans,
	}

	return GetGlobalSpan()
}

func SetMod(mod int) {
	ProcessMod = mod
}

func GetMod() int {
	return ProcessMod
}
