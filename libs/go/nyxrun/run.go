// Package nyxrun is the composition root every service's cmd/<name>/main.go
// ends on (ADR-0001: manual constructor injection, wired exclusively in
// main.go; wiring order config -> otel -> stores -> outbox/bus -> repos ->
// domain -> transports -> nyxrun.Group(...).RunUntilSignal(25s)).
package nyxrun

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

// Runnable is the shape every long-running component (gRPC server, Kafka
// consumer, outbox relay, ops HTTP server) implements to be composed into a
// Group. Run must return once ctx is canceled (crash-only design, RFC-014
// G.0); Group enforces a hard grace deadline on top for stragglers.
type Runnable interface {
	Run(ctx context.Context) error
}

// Runner runs a fixed set of Runnables together and drains them on signal.
type Runner struct {
	runnables []Runnable
}

// Group constructs a Runner over the given Runnables. This is the last line
// of every service's main.go: nyxrun.Group(grpcSrv, consumers, bus.Relay(),
// ops).RunUntilSignal(25 * time.Second) (ECD-004 §4.2 exemplar).
func Group(runnables ...Runnable) *Runner {
	return &Runner{runnables: runnables}
}

// RunUntilSignal runs every Runnable concurrently until SIGTERM/SIGINT (or
// any one Runnable returns an error, which cancels the rest), then allows up
// to grace for all of them to drain before forcing return (RFC-014 G.0:
// "SIGTERM drain <=25s", matching terminationGracePeriod 30s).
func (r *Runner) RunUntilSignal(grace time.Duration) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	return r.run(ctx, grace)
}

// run is RunUntilSignal's logic parameterized by ctx, so tests can drive
// shutdown without sending real OS signals.
func (r *Runner) run(ctx context.Context, grace time.Duration) error {
	g, gctx := errgroup.WithContext(ctx)
	for _, rn := range r.runnables {
		rn := rn
		g.Go(func() error { return rn.Run(gctx) })
	}

	<-gctx.Done() // signal received, or a Runnable returned a non-nil error

	done := make(chan error, 1)
	go func() { done <- g.Wait() }()

	select {
	case err := <-done:
		return err
	case <-time.After(grace):
		return fmt.Errorf("nyxrun: grace period (%s) exceeded before all runnables drained", grace)
	}
}
