package constants

type ContextKey string

const (
	HeaderKeyRequestID string = "X-REQUEST-ID"

	ContextKeyRequestID ContextKey = "requestid"
	ContextKeyTraceID   ContextKey = "traceid"
	ContextKeySpanID    ContextKey = "spanid"
)
