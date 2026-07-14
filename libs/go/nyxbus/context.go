package nyxbus

import "context"

type tenantKey struct{}

// WithTenant stores tenant in ctx for handlers processing a consumed record
// (used at the point of decode, e.g. c.registrar.RecordCompiled(
// nyxbus.WithTenant(ctx, env.TenantId), ...) — ECD-004 §4.2 exemplar).
func WithTenant(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenant)
}

// TenantFromContext retrieves the tenant set by WithTenant, if any.
func TenantFromContext(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(tenantKey{}).(string)
	return t, ok
}
