package tracing

import (
	"context"
	"os"
	"time"

	_ "github.com/limingyao/excellent-go/log/logrus"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

func Init(ctx context.Context, serviceName, endpoint, token string) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	}
	// 1. 创建 exporter
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		log.WithError(err).Fatal()
	}

	// 2. 创建 resource
	//  设置 token or 通过设置环境变量 OTEL_RESOURCE_ATTRIBUTES=token=xxx
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
	}
	if len(token) > 0 {
		attrs = append(attrs, attribute.KeyValue{
			Key: "token", Value: attribute.StringValue(token),
		})
	}
	r, err := resource.New(ctx, []resource.Option{resource.WithAttributes(attrs...)}...)
	if err != nil {
		log.WithError(err).Fatal()
	}

	// 3. 创建 TracerProvider
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(r),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	go func() {
		<-ctx.Done()
		sCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := tp.Shutdown(sCtx); err != nil {
			log.WithError(err).Error()
		}
	}()
}

func InitConsole(ctx context.Context, serviceName string) {
	// 1. 创建 exporter
	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(os.Stderr),
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithoutTimestamps(),
	)
	if err != nil {
		log.WithError(err).Fatal()
	}

	// 2. 创建 resource
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
	}
	r, err := resource.Merge(resource.Default(), resource.NewWithAttributes(semconv.SchemaURL, attrs...))
	if err != nil {
		log.WithError(err).Fatal()
	}

	// 3. 创建 TracerProvider
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(r),
	)

	otel.SetTracerProvider(tp)

	go func() {
		<-ctx.Done()
		sCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := tp.Shutdown(sCtx); err != nil {
			log.WithError(err).Error()
		}
	}()
}
