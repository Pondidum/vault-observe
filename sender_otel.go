package main

// "go.opentelemetry.io/otel"
import (
	"context"
	"crypto/rand"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	"go.opentelemetry.io/otel/semconv"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"
)

type OtelSender struct {
	tp *sdktrace.TracerProvider
}

func NewOtelSender() (*OtelSender, error) {
	exporter, err := zipkin.NewRawExporter(
		"http://localhost:9411/api/v2/spans",
		zipkin.WithSDKOptions(sdktrace.WithSampler(sdktrace.AlwaysSample())),
	)
	if err != nil {
		return nil, err
	}

	processor := sdktrace.NewSimpleSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(processor),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.ServiceNameKey.String("vault-observe"),
		)),
		sdktrace.WithIDGenerator(&Generator{}),
	)
	otel.SetTracerProvider(tp)

	return &OtelSender{tp: tp}, nil
}

func (o *OtelSender) Send(typed Event, event map[string]interface{}) error {
	id, err := uuid.Parse(typed.Request.ID)
	if err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), "request_id", id)

	tr := otel.GetTracerProvider().Tracer("main")
	ctx, span := tr.Start(ctx, typed.Type, trace.WithSpanKind(trace.SpanKindServer))

	for key, value := range event {
		span.SetAttributes(attribute.KeyValue{
			Key:   attribute.Key(key),
			Value: attribute.StringValue(fmt.Sprintf("%v", value)),
		})
	}

	if typed.Error != "" {
		span.SetStatus(codes.Error, typed.Error)
	}

	span.End()
	return nil
}

func (o *OtelSender) Shutdown() error {
	return o.tp.Shutdown(context.Background())
}

type Generator struct {
}

func (g *Generator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	val := ctx.Value("request_id").(uuid.UUID)
	tid := trace.TraceID{}
	req, _ := val.MarshalText()
	copy(tid[:], req)

	sid := trace.SpanID{}
	rand.Read(sid[:])

	return tid, sid
}

func (g *Generator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	sid := trace.SpanID{}
	rand.Read(sid[:])

	return sid
}
