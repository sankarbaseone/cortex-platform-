package nyxotel

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// NewLogger returns the platform structured JSON logger: fields
// {ts,level,svc,msg,...}, matching RFC-014 G.0's log schema. Payload
// contents must NEVER be logged (RFC-014 G.0) — callers are responsible for
// not passing raw event/message payloads as attributes.
func NewLogger(svc string) *slog.Logger {
	return newLogger(os.Stdout, svc)
}

func newLogger(w io.Writer, svc string) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "ts"
			case slog.MessageKey:
				a.Key = "msg"
			case slog.LevelKey:
				a.Key = "level"
			}
			return a
		},
	})
	return slog.New(h).With("svc", svc)
}

// WithTrace returns l with a trace_id attribute set from ctx's active span,
// if any (RFC-005 B.2: envelope trace_id is the W3C traceparent value).
func WithTrace(ctx context.Context, l *slog.Logger) *slog.Logger {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return l
	}
	return l.With("trace_id", sc.TraceID().String())
}

// WithTenant returns l with a tenant attribute set (RFC-014 G.0 log schema).
func WithTenant(l *slog.Logger, tenant string) *slog.Logger {
	return l.With("tenant", tenant)
}
