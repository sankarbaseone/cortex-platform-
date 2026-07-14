package nyxpg

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nydux/platform/libs/go/nyxerr"
)

// WithTenantTx must reject an empty tenant before ever touching the pool
// (RFC-012 O.4: tenant is mandatory input, never defaulted) — checked here
// against a nil pool to prove the guard runs before any DB access.
func TestWithTenantTx_EmptyTenantRejectedBeforeDBAccess(t *testing.T) {
	p := &Pool{pool: nil}
	err := p.WithTenantTx(context.Background(), "", func(context.Context, pgx.Tx) error {
		t.Fatal("fn must not run when tenant is empty")
		return nil
	})
	if err == nil {
		t.Fatal("expected error for empty tenant")
	}
	if !nyxerr.Is(err, nyxerr.InvalidArgument) {
		t.Fatalf("KindOf(err) = %v, want InvalidArgument", nyxerr.KindOf(err))
	}
}

func TestTenantContext_RoundTrip(t *testing.T) {
	ctx := WithTenant(context.Background(), "tenant-42")
	got, ok := TenantFromContext(ctx)
	if !ok || got != "tenant-42" {
		t.Fatalf("TenantFromContext = (%q, %v), want (tenant-42, true)", got, ok)
	}
}

func TestTenantContext_AbsentByDefault(t *testing.T) {
	if _, ok := TenantFromContext(context.Background()); ok {
		t.Fatal("TenantFromContext must report absent on a bare context")
	}
}
