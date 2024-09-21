package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"go.opentelemetry.io/otel/trace"
)

type CustomLogger struct {
	logger      zerolog.Logger
	serviceName string
}

const (
	CORRELATION_ID = "correlation_id"
)

func NewOtelLogger(otelContext *OtelContext) *CustomLogger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Create a new logger
	logger := getLogger()
	logger = logger.Hook(otelContext.GetZerologHook())

	log.Logger = logger

	return &CustomLogger{
		logger:      logger,
		serviceName: otelContext.GetServiceName(),
	}
}

func getLogger() zerolog.Logger {
	// Format: [LEVEL] [TIME] [CORRELATION ID] [CALLER] Message
	return log.Logger.Output(zerolog.ConsoleWriter{
		Out: os.Stdout,
		FormatLevel: func(i any) string {
			str := strings.ToUpper(fmt.Sprintf("%s", i))
			return fmt.Sprintf("[%s]", str)
		},
		FormatMessage: func(i any) string {
			return fmt.Sprintf("%s", i)
		},
		PartsOrder: []string{
			"level",
			"time",
			"Service",
			CORRELATION_ID,
			"message",
		},
		FormatFieldValue: func(i any) string {
			if i == nil {
				return ""
			}
			return fmt.Sprintf("[%s]", i)
		},
		FieldsExclude: []string{CORRELATION_ID, "Service", "span_id", "trace_id"},
		FormatTimestamp: func(i any) string {
			return time.Now().Format("[2006-01-02 15:04:05]")
		},
	})
}

func (cl *CustomLogger) Info(ctx context.Context, message string) {
	cl.handleGeneralLog(ctx, message, false)
}

func (cl *CustomLogger) Error(ctx context.Context, message string) {
	cl.handleGeneralLog(ctx, message, true)
}

func contextExtractor(ctx context.Context) context.Context {
	switch ctx := ctx.(type) {
	case *gin.Context:
		return ctx.Request.Context()
	case context.Context:
		return ctx
	default:
		return context.TODO()
	}
}

func (cl *CustomLogger) handleGeneralLog(ctx context.Context, message string, isError bool) {
	correlationId := ctx.Value(CORRELATION_ID)
	if correlationId == nil {
		correlationId = "N/A"
	}

	reqCtx := contextExtractor(ctx)

	// Extract traceId and spanId from the context
	spanContext := trace.SpanContextFromContext(reqCtx)
	traceId := spanContext.TraceID().String()
	spanId := spanContext.SpanID().String()

	if isError {
		cl.logger.Error().
			Str("Service", cl.serviceName).
			Str(CORRELATION_ID, fmt.Sprintf("%s", correlationId)).
			Str("trace_idd", traceId).
			Str("span_id", spanId).
			Msg(message)
	} else {
		cl.logger.Info().
			Str("Service", cl.serviceName).
			Str(CORRELATION_ID, fmt.Sprintf("%s", correlationId)).
			Str("trace_id", traceId).
			Str("span_id", spanId).
			Msg(message)
	}
}
