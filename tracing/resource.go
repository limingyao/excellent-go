package tracing

import (
	"go.opentelemetry.io/otel/attribute"
)

const (
	SessionKey = attribute.Key("session_id")
	RequestKey = attribute.Key("request_id")
)
