package nyxhttp

import (
	"errors"
	"testing"
)

func TestNewBreaker_TripsAfterFiveConsecutiveFailures(t *testing.T) {
	b := NewBreaker[int]("test-dep")
	boom := errors.New("upstream down")

	for i := 0; i < 5; i++ {
		_, err := b.Execute(func() (int, error) { return 0, boom })
		if !errors.Is(err, boom) {
			t.Fatalf("call %d: err = %v, want the underlying failure", i, err)
		}
	}

	// The 6th call should be rejected by the now-open breaker rather than
	// invoking the wrapped function at all.
	called := false
	_, err := b.Execute(func() (int, error) { called = true; return 0, nil })
	if called {
		t.Fatal("breaker should be open and must not invoke the wrapped call")
	}
	if err == nil {
		t.Fatal("expected an open-circuit error")
	}
}

func TestNewBreaker_StaysClosedOnSuccess(t *testing.T) {
	b := NewBreaker[string]("healthy-dep")
	for i := 0; i < 10; i++ {
		got, err := b.Execute(func() (string, error) { return "ok", nil })
		if err != nil || got != "ok" {
			t.Fatalf("call %d: got=%q err=%v, want ok/nil", i, got, err)
		}
	}
}
