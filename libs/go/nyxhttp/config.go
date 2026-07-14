// Package nyxhttp provides the uniform ops surface every service exposes
// (RFC-014 G.0): /healthz (liveness), /readyz (readiness, 2s dep-ping
// budget), /metrics (Prometheus), all on :9090. It also provides the
// outbound circuit breaker every sync call must be wrapped in (RFC-014 G.0:
// "all outbound sync calls wrapped (gobreaker): 5 consecutive failures ->
// open 30s -> half-open probe").
package nyxhttp

import "time"

// Config is embedded (as HTTPOps) in each service's own config struct.
type Config struct {
	Addr string `envconfig:"HTTP_OPS_ADDR" default:":9090"`
}

// readyBudget bounds total time spent probing readiness checks (RFC-014
// G.0: "readiness: deps ping w/ 2s budget").
const readyBudget = 2 * time.Second
