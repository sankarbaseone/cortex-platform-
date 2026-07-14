// Package main is not a real service — it is the T-003 compile-proof that
// libs/go/{config,nyxotel,nyxpg,nyxbus,nyxclock,nyxhttp,nyxrun,nyxerr} wire
// together correctly, in the ADR-0001 order every real service's
// cmd/<name>/main.go follows: config -> otel -> stores -> outbox/bus ->
// repos -> domain use-cases -> transports -> nyxrun.Group(...).
// RunUntilSignal(25s). A real service's domain/ports/adapters (kernel-
// registry etc.) are built out under services/* starting at T-101; this
// file intentionally has no business logic.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/nydux/platform/libs/go/config"
	"github.com/nydux/platform/libs/go/nyxbus"
	"github.com/nydux/platform/libs/go/nyxclock"
	"github.com/nydux/platform/libs/go/nyxerr"
	"github.com/nydux/platform/libs/go/nyxhttp"
	"github.com/nydux/platform/libs/go/nyxotel"
	"github.com/nydux/platform/libs/go/nyxpg"
	"github.com/nydux/platform/libs/go/nyxrun"
)

// Config demonstrates the platform-wide config contract (ECD-002
// internal/config.go row): each service embeds the shared lib sub-configs
// and adds its own fields alongside them.
type Config struct {
	OTel    nyxotel.Config
	PG      nyxpg.Config
	Kafka   nyxbus.Config
	HTTPOps nyxhttp.Config
}

func (c *Config) Validate() error { return nil } // sub-configs self-validate via `required:"true"` tags

func (c *Config) Redacted() string {
	return fmt.Sprintf("otel.service=%s pg.dsn=<redacted> kafka.brokers=%v http.addr=%s",
		c.OTel.ServiceName, c.Kafka.Brokers, c.HTTPOps.Addr)
}

func main() {
	// 1. config
	cfg := config.MustLoad[Config]()

	// 2. otel
	shutdown := nyxotel.MustInit(cfg.OTel)
	defer func() { _ = shutdown(context.Background()) }()

	// 3. stores
	pool := nyxpg.MustPool(context.Background(), cfg.PG)
	defer pool.Close()
	outbox := nyxpg.NewOutbox(pool)

	// 4. outbox/bus
	producer, err := nyxbus.NewProducer(cfg.Kafka)
	if err != nil {
		panic(err)
	}
	defer producer.Close()
	bus := nyxbus.NewOutboxPublisher(outbox, producer)

	// 5. repos / 6. domain use-cases — service-specific (T-101+); this
	// exemplar only proves nyxclock's Clock is wireable into a use-case.
	clock := nyxclock.Real{}
	_ = clock.Now()

	// 7. transports
	ops := nyxhttp.Ops(cfg.HTTPOps, nyxhttp.NamedCheck{
		Name: "pg",
		Check: func(ctx context.Context) error {
			if err := pool.Raw().Ping(ctx); err != nil {
				return nyxerr.Wrap(nyxerr.Unavailable, "readyz.pg", err)
			}
			return nil
		},
	})

	// 8. nyxrun.Group(...).RunUntilSignal(25s) — ADR-0001's fixed wiring tail.
	if err := nyxrun.Group(bus.Relay(), ops).RunUntilSignal(25 * time.Second); err != nil {
		panic(err)
	}
}
