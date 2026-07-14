package nyxrun

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fnRunnable func(ctx context.Context) error

func (f fnRunnable) Run(ctx context.Context) error { return f(ctx) }

func TestRun_AllDrainCleanlyOnCancel(t *testing.T) {
	drained := make(chan struct{}, 2)
	r := Group(
		fnRunnable(func(ctx context.Context) error { <-ctx.Done(); drained <- struct{}{}; return nil }),
		fnRunnable(func(ctx context.Context) error { <-ctx.Done(); drained <- struct{}{}; return nil }),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // simulate signal already received

	if err := r.run(ctx, time.Second); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if len(drained) != 2 {
		t.Fatalf("drained %d runnables, want 2", len(drained))
	}
}

func TestRun_OneFailureCancelsTheRest(t *testing.T) {
	boom := errors.New("grpc server crashed")
	otherCanceled := make(chan struct{})

	r := Group(
		fnRunnable(func(ctx context.Context) error { return boom }),
		fnRunnable(func(ctx context.Context) error { <-ctx.Done(); close(otherCanceled); return nil }),
	)

	err := r.run(context.Background(), time.Second)
	if !errors.Is(err, boom) {
		t.Fatalf("run() error = %v, want %v", err, boom)
	}
	select {
	case <-otherCanceled:
	case <-time.After(time.Second):
		t.Fatal("the surviving runnable was never canceled after its sibling failed")
	}
}

func TestRun_GracePeriodExceededForcesReturn(t *testing.T) {
	r := Group(
		fnRunnable(func(ctx context.Context) error {
			<-ctx.Done()
			time.Sleep(time.Hour) // never actually drains in time
			return nil
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	err := r.run(ctx, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected a grace-period-exceeded error")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("run() took %v to force-return, want ~50ms", elapsed)
	}
}

func TestRun_EmptyGroupReturnsPromptlyOnCancel(t *testing.T) {
	r := Group()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := r.run(ctx, time.Second); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}
