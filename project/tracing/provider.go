package tracing

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func ConfigureTraceProvider() (*tracesdk.TracerProvider, error) {
	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
	var urlPath string

	if jaegerEndpoint == "" {
		gatewayAddr := os.Getenv("GATEWAY_ADDR")
		parsedURL, err := url.Parse(gatewayAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse GATEWAY_ADDR: %w", err)
		}
		jaegerEndpoint = parsedURL.Host
		urlPath = "/jaeger-api/api/traces"
	} else {
		jaegerEndpoint = strings.TrimPrefix(jaegerEndpoint, "http://")
		jaegerEndpoint = strings.TrimPrefix(jaegerEndpoint, "https://")
	}

	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(jaegerEndpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithURLPath(urlPath),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSyncer(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("tickets"),
		)),
	)

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}
