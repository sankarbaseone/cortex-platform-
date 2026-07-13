// artifacts/go-examples/kernelregistry_exemplar.go
// ============================================================================
// EXEMPLAR IMPLEMENTATION PACKAGE (ECD-004 §4.2) — kernel-registry service.
// This file concatenates, for review convenience, what lands in the repo as:
//   services/kernel-registry/internal/domain/register.go
//   services/kernel-registry/internal/ports/ports.go
//   services/kernel-registry/internal/adapters/pg/repo.go
//   services/kernel-registry/internal/adapters/kafka/consumer.go
//   services/kernel-registry/cmd/kernel-registry/main.go   (DI wiring)
// Claude Code splits it at the ===== FILE markers verbatim.
// DI STANDARD (platform-wide, ECD-004 §4.1): manual constructor injection,
// wired ONLY in cmd/<name>/main.go. No DI framework (wire/fx rejected:
// magic obscures the dependency graph depguard enforces; codegen wire adds
// a build step for zero benefit at 16 services).
// ============================================================================

// ===== FILE services/kernel-registry/internal/domain/register.go =====
package domain

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
)

var hashRe = regexp.MustCompile(`^[a-f0-9]{64}$`)

var (
	ErrInvalidHash      = errors.New("invalid kernel hash")
	ErrIllegalTransition = errors.New("illegal status transition")
	ErrNotFound         = errors.New("kernel not found")
)

type KernelStatus string

const (
	StatusScored   KernelStatus = "SCORED"
	StatusStatic   KernelStatus = "STATIC_ESTIMATE"
	StatusMeasOnly KernelStatus = "MEASUREMENT_ONLY"
)

// legal forward-only transitions (ECD-003 §3.3).
var legal = map[KernelStatus]map[KernelStatus]bool{
	StatusMeasOnly: {StatusMeasOnly: true, StatusStatic: true, StatusScored: true},
	StatusStatic:   {StatusStatic: true, StatusScored: true},
	StatusScored:   {StatusScored: true}, // rescore allowed
}

type Score struct {
	Value        float64
	Components   map[string]float64
	ModelVersion string
	Confidence   float64
}

type Kernel struct {
	Hash, Family, Arch, ToolchainFP string
	Tenant                          string // uuid string; newtype omitted in exemplar for brevity of file split
	Status                          KernelStatus
	Score                           *Score
	FirstSeen, LastSeen             time.Time
}

type Clock interface{ Now() time.Time }

// Registrar is the single domain use-case object for kernel upserts.
type Registrar struct {
	repo  KernelRepo
	bus   Publisher
	clock Clock
}

func NewRegistrar(repo KernelRepo, bus Publisher, clock Clock) *Registrar {
	return &Registrar{repo: repo, bus: bus, clock: clock}
}

// RecordScored upserts a scored kernel, enforcing hash validity and the
// status state machine, and publishes compiler.kernel.scored via outbox.
//
// Business rules (ECD-004 function spec F-KR-01):
//   - hash & family must match ^[a-f0-9]{64}$          → ErrInvalidHash
//   - transition must be legal per table               → ErrIllegalTransition
//   - SCORED requires non-nil Score with kes∈[0,100]   → ErrInvalidScore path
//   - first_seen preserved on upsert; last_seen = now
// Concurrency: safe for concurrent calls; repo upsert is the serialization point.
// Performance target: p95 < 5ms excluding network (bench required).
func (r *Registrar) RecordScored(ctx context.Context, k Kernel) error {
	if !hashRe.MatchString(k.Hash) || !hashRe.MatchString(k.Family) {
		return fmt.Errorf("record scored: %w", ErrInvalidHash)
	}
	if k.Status == StatusScored {
		if k.Score == nil || k.Score.Value < 0 || k.Score.Value > 100 {
			return fmt.Errorf("record scored: score required in [0,100]: %w", ErrIllegalTransition)
		}
	}
	prev, err := r.repo.Get(ctx, k.Tenant, k.Hash)
	switch {
	case errors.Is(err, ErrNotFound):
		k.FirstSeen = r.clock.Now()
	case err != nil:
		return fmt.Errorf("record scored: get: %w", err)
	default:
		if !legal[prev.Status][k.Status] {
			return fmt.Errorf("%s -> %s: %w", prev.Status, k.Status, ErrIllegalTransition)
		}
		k.FirstSeen = prev.FirstSeen
	}
	k.LastSeen = r.clock.Now()
	if err := r.repo.Upsert(ctx, k); err != nil {
		return fmt.Errorf("record scored: upsert: %w", err)
	}
	return r.bus.PublishKernelScored(ctx, k) // outbox-backed; same tx in pg adapter
}

// ===== FILE services/kernel-registry/internal/ports/ports.go =====
package ports // (in repo; exemplar keeps single package for compile demo)

type KernelRepo interface {
	Get(ctx context.Context, tenant, hash string) (Kernel, error) // ErrNotFound sentinel
	Upsert(ctx context.Context, k Kernel) error                   // idempotent by (hash)
}

type Publisher interface {
	PublishKernelScored(ctx context.Context, k Kernel) error
}

// ===== FILE services/kernel-registry/internal/adapters/pg/repo.go =====
// (sqlc-style; hand-shown here so the exemplar is self-contained)
/*
package pg

const upsertSQL = `
INSERT INTO kernels (kernel_hash, tenant_id, family_hash, arch, status, kes,
  kes_model_version, confidence, toolchain_fp, first_seen, last_seen)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (kernel_hash) DO UPDATE SET
  status=$5, kes=$6, kes_model_version=$7, confidence=$8,
  toolchain_fp=$9, last_seen=$11`

func (r *Repo) Upsert(ctx context.Context, k domain.Kernel) error {
	return r.pool.WithTenantTx(ctx, k.Tenant, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, upsertSQL, args(k)...); err != nil { return err }
		return r.outbox.Enqueue(ctx, tx, "nydux.compiler.kernel.scored.v1",
			k.Tenant+":"+k.Hash, encodeScored(k)) // SAME TRANSACTION = outbox pattern
	})
}
*/

// ===== FILE services/kernel-registry/internal/adapters/kafka/consumer.go =====
/*
package kafka

// HandleCompiled consumes nydux.compiler.kernel.compiled.v1.
// Idempotency: upsert-by-natural-key (topics.yaml). Duplicate/out-of-order safe.
func (c *Consumer) HandleCompiled(ctx context.Context, env commonv1.EventEnvelope) error {
	var msg compilerv1.KernelCompiled
	if err := proto.Unmarshal(env.Payload, &msg); err != nil {
		return nyxbus.Poison(err) // straight to DLQ, envelope intact (RFC-005 B.4)
	}
	return c.registrar.RecordCompiled(nyxbus.WithTenant(ctx, env.TenantId), toDomain(msg))
}
*/

// ===== FILE services/kernel-registry/cmd/kernel-registry/main.go =====
/*
package main

func main() {
	cfg := config.MustLoad()                 // env → struct → Validate() → Redacted() print
	shut := nyxotel.MustInit(cfg.OTel)       // traces+metrics+logger
	defer shut()

	pool := nyxpg.MustPool(cfg.PG)           // sets nydux.tenant GUC per call
	outbox := nyxpg.NewOutbox(pool)
	bus := nyxbus.NewOutboxPublisher(outbox) // relay goroutine: poll 250ms, batch 500
	repo := pg.NewRepo(pool, outbox)

	registrar := domain.NewRegistrar(repo, bus, nyxclock.Real{})

	grpcSrv := grpcadapter.New(cfg.GRPC, registrar) // interceptor chain fixed order (ECD-002 NSFS)
	consumers := kafka.NewConsumers(cfg.Kafka, registrar)
	ops := nyxhttp.Ops(cfg.HTTPOps)          // /metrics /healthz /readyz on :9090

	nyxrun.Group(grpcSrv, consumers, bus.Relay(), ops).RunUntilSignal(25 * time.Second)
}
*/
