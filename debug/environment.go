package debug

import (
	"os"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

const AttributesKey string = "attributes"

func MarshalAttributes(attrs []attribute.KeyValue) []byte {
	sb := strings.Builder{}

	for _, kv := range attrs {
		sb.WriteString(string(kv.Key))
		sb.WriteString(": ")
		sb.WriteString(kv.Value.AsString())
		sb.WriteString("\n")
	}

	return []byte(sb.String())
}

func EnvironmentData() []attribute.KeyValue {

	if v := os.Getenv("GITHUB_ACTIONS"); v == "true" {
		return []attribute.KeyValue{
			semconv.VCSRefHeadName(os.Getenv("GITHUB_REF_NAME")),
			semconv.VCSRefHeadRevision(os.Getenv("GITHUB_SHA")),
			semconv.VCSRefHeadName(os.Getenv("GITHUB_REF_TYPE")),
			attribute.String("github.run.attempt", os.Getenv("GITHUB_RUN_ATTEMPT")),
			attribute.String("github.run.id", os.Getenv("GITHUB_RUN_ID")),
			attribute.String("github.run.number", os.Getenv("GITHUB_RUN_NUMBER")),
			attribute.String("github.workflow.name", os.Getenv("GITHUB_WORKFLOW")),
			attribute.String("github.step.name", os.Getenv("GITHUB_JOB")),
		}
	}

	return nil
}
