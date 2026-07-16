//go:build ignore

// Semgrep test fixture only (see tools/semgrep/rules/no-payload-logging.yaml).
package tests

import "log/slog"

func handle(log *slog.Logger, body []byte, id string) {
	// ruleid: go-no-payload-in-logs
	log.Info("received", "payload", body)

	// ruleid: go-no-payload-in-logs
	log.Error("decode failed", "raw_ir", body)

	// ok: go-no-payload-in-logs
	log.Info("received", "id", id, "size", len(body))
}
