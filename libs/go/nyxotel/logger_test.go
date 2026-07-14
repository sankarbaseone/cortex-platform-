package nyxotel

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNewLogger_EmitsPlatformSchema(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
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
	})).With("svc", "kernel-registry")

	l = WithTenant(l, "tenant-1")
	l.Info("kernel recorded")

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("log line not valid JSON: %v (%s)", err, buf.String())
	}
	for _, key := range []string{"ts", "level", "svc", "tenant", "msg"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("log line missing field %q: %v", key, got)
		}
	}
	if got["svc"] != "kernel-registry" || got["tenant"] != "tenant-1" || got["msg"] != "kernel recorded" {
		t.Fatalf("unexpected log fields: %v", got)
	}
}

func TestWithTrace_NoActiveSpanIsNoop(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewJSONHandler(&buf, nil))
	got := WithTrace(context.Background(), l)
	if got != l {
		t.Fatalf("WithTrace without an active span must return l unchanged")
	}
}

func TestNewLogger_ServiceNameSet(t *testing.T) {
	var buf bytes.Buffer
	l := newLogger(&buf, "finance-svc")
	l.Info("x")
	if !strings.Contains(buf.String(), `"svc":"finance-svc"`) {
		t.Fatalf("expected svc field in output, got %s", buf.String())
	}
}
