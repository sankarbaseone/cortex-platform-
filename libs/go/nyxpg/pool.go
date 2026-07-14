// Package nyxpg wraps pgxpool with the platform's multi-tenancy contract:
// every query runs with the RLS tenant GUC set for the current transaction
// (RFC-007: "All tenant tables get RLS ... services set GUC per request";
// RFC-009 I.6: JWT tenant claim -> request GUC -> RLS is isolation layer
// 1-2 of defense-in-depth). It also implements the transactional outbox
// pattern (RFC-012 O.4) so a domain write and its event commit atomically.
package nyxpg

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nydux/platform/libs/go/nyxerr"
)

// Config is embedded (as PG) in each service's own config struct.
type Config struct {
	DSN         string        `envconfig:"PG_DSN" required:"true"`
	MaxConns    int32         `envconfig:"PG_MAX_CONNS" default:"10"`
	MinConns    int32         `envconfig:"PG_MIN_CONNS" default:"0"`
	ConnTimeout time.Duration `envconfig:"PG_CONN_TIMEOUT" default:"5s"`
}

// Pool is a tenant-aware Postgres connection pool.
type Pool struct {
	pool *pgxpool.Pool
}

// NewPool parses cfg.DSN, opens a pool sized per cfg, and pings it.
func NewPool(ctx context.Context, cfg Config) (*Pool, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("nyxpg: parse config: %w", err)
	}
	if cfg.MaxConns > 0 {
		pgxCfg.MaxConns = cfg.MaxConns
	}
	pgxCfg.MinConns = cfg.MinConns

	connCtx, cancel := context.WithTimeout(ctx, cfg.ConnTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connCtx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("nyxpg: connect: %w", err)
	}
	if err := pool.Ping(connCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("nyxpg: ping: %w", err)
	}
	return &Pool{pool: pool}, nil
}

// MustPool is NewPool for main.go wiring (ADR-0001): panics on failure since
// a service cannot run without its store.
func MustPool(ctx context.Context, cfg Config) *Pool {
	p, err := NewPool(ctx, cfg)
	if err != nil {
		panic(err)
	}
	return p
}

// Close releases all pooled connections.
func (p *Pool) Close() { p.pool.Close() }

// Raw exposes the underlying *pgxpool.Pool for adapters needing direct
// access (e.g. sqlc-generated queries taking a pgxpool.Pool/pgx.Tx).
func (p *Pool) Raw() *pgxpool.Pool { return p.pool }

// WithTenantTx runs fn inside a transaction with the RLS tenant GUC set for
// that transaction only. Uses `SELECT set_config(..., true)` rather than
// `SET LOCAL x = $1`, since SET does not accept bind parameters — set_config
// does, and its third argument (true) scopes the setting to the current
// transaction (RFC-007: RLS policy reads current_setting('nydux.tenant')).
func (p *Pool) WithTenantTx(ctx context.Context, tenant string, fn func(ctx context.Context, tx pgx.Tx) error) error {
	if tenant == "" {
		return nyxerr.New(nyxerr.InvalidArgument, "nyxpg.WithTenantTx", "tenant is required")
	}
	return pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT set_config('nydux.tenant', $1, true)`, tenant); err != nil {
			return fmt.Errorf("nyxpg: set tenant GUC: %w", err)
		}
		return fn(ctx, tx)
	})
}

type tenantKey struct{}

// WithTenant stores tenant in ctx for adapters that read it out-of-band
// (ECD-002: "RLS GUC set per call via nyxpg.WithTenant(ctx)").
func WithTenant(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenant)
}

// TenantFromContext retrieves the tenant set by WithTenant, if any.
func TenantFromContext(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(tenantKey{}).(string)
	return t, ok
}
