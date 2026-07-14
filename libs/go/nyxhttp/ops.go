package nyxhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ReadinessCheck pings one dependency; it must respect ctx's deadline.
type ReadinessCheck func(ctx context.Context) error

// NamedCheck pairs a ReadinessCheck with a name for /readyz's response body.
type NamedCheck struct {
	Name  string
	Check ReadinessCheck
}

// OpsServer serves /healthz, /readyz, /metrics on one listener (RFC-014
// G.0). Other services have no HTTP business endpoints on this port.
type OpsServer struct {
	srv *http.Server
}

// Ops constructs the ops HTTP server. checks are run (concurrently, each
// bounded by the shared 2s budget) on every /readyz call.
func Ops(cfg Config, checks ...NamedCheck) *OpsServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/readyz", handleReadyz(checks))
	mux.Handle("/metrics", promhttp.Handler())

	return &OpsServer{srv: &http.Server{Addr: cfg.Addr, Handler: mux}}
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

type readyResult struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

func handleReadyz(checks []NamedCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readyBudget)
		defer cancel()

		type outcome struct {
			name string
			err  error
		}
		results := make(chan outcome, len(checks))
		for _, c := range checks {
			c := c
			go func() { results <- outcome{c.Name, c.Check(ctx)} }()
		}

		res := readyResult{Status: "ok", Checks: map[string]string{}}
		ok := true
		for range checks {
			o := <-results
			if o.err != nil {
				ok = false
				res.Checks[o.name] = o.err.Error()
			} else {
				res.Checks[o.name] = "ok"
			}
		}
		if !ok {
			res.Status = "unready"
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}
}

// Run implements the nyxrun.Runnable shape: serves until ctx is canceled,
// then drains via graceful http.Server.Shutdown.
func (o *OpsServer) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(1)
	errCh := make(chan error, 1)
	go func() {
		defer wg.Done()
		if err := o.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := o.srv.Shutdown(shutdownCtx)
		wg.Wait()
		return err
	case err := <-errCh:
		return err
	}
}
