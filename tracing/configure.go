package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

func Configure(ctx context.Context, appName string, version string) (func(ctx context.Context) error, error) {
	traceExporter, err := createTraceExporter(ctx)
	if err != nil {
		return nil, err
	}

	resource, err := createResource(appName, version)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithResource(resource),
		trace.WithBatcher(traceExporter),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp.Shutdown, nil
}

func createResource(appName string, version string) (*resource.Resource, error) {

	env := os.Getenv("DEPLOY_ENV")

	if env == "" {
		host, err := os.Hostname()
		if err != nil {
			host = "unknown-machine"
		}
		env = "local:" + host
	}

	return resource.New(
		context.Background(),
		resource.WithTelemetrySDK(),
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceName(appName),
			semconv.ServiceVersion(version),
			semconv.DeploymentEnvironmentName(env),

			semconv.VCSRepositoryName("crash-reporter"),
			semconv.VCSRepositoryURLFull("https://github.ol.epicgames.net/online-web/crash-reporter"),
			semconv.VCSRefHeadName(os.Getenv("BRANCH_NAME")),
			semconv.VCSRefHeadRevision(os.Getenv("COMMIT_HASH")),
			semconv.VCSRefHeadTypeBranch,
		),
	)

}
func createTraceExporter(ctx context.Context) (trace.SpanExporter, error) {

	exporter := os.Getenv("OTEL_TRACE_EXPORTER")

	switch exporter {
	case "none":
		return nil, nil

	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())

	case "stderr":
		return stdouttrace.New(stdouttrace.WithPrettyPrint(), stdouttrace.WithWriter(os.Stderr))

	default:
		return otlptracegrpc.New(ctx)
	}
}
