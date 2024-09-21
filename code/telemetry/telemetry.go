package telemetry

import (
	"context"
	"fmt"
	glog "log"

	"github.com/agoda-com/opentelemetry-go/otelzerolog"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	otlpLogsHttp "github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogshttp"
	otlpLogsSdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"go.mongodb.org/mongo-driver/event"

	"github.com/gin-gonic/gin"

	gcpdetector "go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	otlpTraceHttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type (
	OtelContext struct {
		appName  string
		resource *resource.Resource
		tracing
		logging
	}

	tracing struct {
		tracer   *trace.Tracer
		provider *sdktrace.TracerProvider
		exporter *otlptrace.Exporter
	}

	logging struct {
		provider *otlpLogsSdk.LoggerProvider
		exporter *otlplogs.Exporter
	}
)

func NewOtelContext(endpoint, token, service string) (*OtelContext, error) {
	newResource, err := resource.New(context.TODO(),
		resource.WithDetectors(gcpdetector.NewDetector()),
		resource.WithAttributes(semconv.ServiceName(service)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}
	resource, err := resource.Merge(resource.Default(), newResource)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}

	otlpLogsHttpClient := otlpLogsHttp.NewClient(
		otlpLogsHttp.WithEndpoint(endpoint),
		otlpLogsHttp.WithInsecure(),
		otlpLogsHttp.WithHeaders(map[string]string{"Authorization": "Bearer " + token}),
		otlpLogsHttp.WithURLPath("/otlp_http/v1/logs"),
	)
	logsExporter, err := otlplogs.NewExporter(context.TODO(), otlplogs.WithClient(otlpLogsHttpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create collector exporter: %v", err)
	}

	loggerProvider := otlpLogsSdk.NewLoggerProvider(
		otlpLogsSdk.WithBatcher(logsExporter),
		otlpLogsSdk.WithResource(resource),
	)

	traceExporter, err := otlpTraceHttp.New(context.TODO(),
		otlpTraceHttp.WithEndpoint(endpoint),
		otlpTraceHttp.WithInsecure(),
		otlpTraceHttp.WithHeaders(map[string]string{"Authorization": "Bearer " + token}),
		otlpTraceHttp.WithURLPath("/otlp_http/v1/traces"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %v", err)
	}

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(traceExporter),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := otel.Tracer(service)

	return &OtelContext{
		appName:  service,
		resource: resource,
		tracing: tracing{
			tracer:   &tracer,
			provider: traceProvider,
			exporter: traceExporter,
		},
		logging: logging{
			provider: loggerProvider,
			exporter: logsExporter,
		},
	}, nil
}

func (o *OtelContext) Shutdown() {
	if err := o.logging.provider.Shutdown(context.Background()); err != nil {
		glog.Printf("failed to shutdown logger provider: %v", err)
	}
	if err := o.tracing.provider.Shutdown(context.Background()); err != nil {
		glog.Printf("failed to shutdown trace provider: %v", err)
	}
}

func (o *OtelContext) GetTracer() *trace.Tracer {
	return o.tracing.tracer
}

func (o *OtelContext) GetResource() *resource.Resource {
	return o.resource
}

func (o *OtelContext) GetZerologHook() *otelzerolog.Hook {
	return otelzerolog.NewHook(o.logging.provider)
}

func (o *OtelContext) GetGinMiddleware() gin.HandlerFunc {
	return otelgin.Middleware(o.appName)
}

func (o *OtelContext) GetMongoTraceqlHook() *event.CommandMonitor {
	return otelmongo.NewMonitor(
		otelmongo.WithTracerProvider(o.tracing.provider),
		otelmongo.WithCommandAttributeDisabled(false),
	)
}

func (o *OtelContext) GetMongoDefaultHook() *event.CommandMonitor {
	return otelmongo.NewMonitor()
}

func (o *OtelContext) GetServiceName() string {
	return o.appName
}
