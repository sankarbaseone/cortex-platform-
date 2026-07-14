package nyxhttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleHealthz_AlwaysOK(t *testing.T) {
	rec := httptest.NewRecorder()
	handleHealthz(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestHandleReadyz_AllChecksPass(t *testing.T) {
	h := handleReadyz([]NamedCheck{
		{Name: "pg", Check: func(context.Context) error { return nil }},
		{Name: "kafka", Check: func(context.Context) error { return nil }},
	})
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleReadyz_OneCheckFails(t *testing.T) {
	h := handleReadyz([]NamedCheck{
		{Name: "pg", Check: func(context.Context) error { return nil }},
		{Name: "kafka", Check: func(context.Context) error { return errors.New("down") }},
	})
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503, body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandleReadyz_RespectsBudget(t *testing.T) {
	h := handleReadyz([]NamedCheck{
		{Name: "slow", Check: func(ctx context.Context) error {
			select {
			case <-time.After(time.Hour):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}},
	})
	rec := httptest.NewRecorder()
	start := time.Now()
	h(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if elapsed := time.Since(start); elapsed > readyBudget+time.Second {
		t.Fatalf("readyz took %v, want bounded by the %v budget", elapsed, readyBudget)
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 for a budget-exceeding check", rec.Code)
	}
}

func TestOpsServer_RunStopsOnContextCancel(t *testing.T) {
	o := Ops(Config{Addr: "127.0.0.1:0"})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- o.Run(ctx) }()

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error on graceful shutdown = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not return within 2s of context cancellation")
	}
}
