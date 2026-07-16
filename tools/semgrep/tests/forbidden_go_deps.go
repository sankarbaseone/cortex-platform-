//go:build ignore

// Semgrep test fixture only (see tools/semgrep/rules/forbidden-go-deps.yaml).
// go:build ignore keeps this out of `go build ./...`/`go vet` - the imports
// below are intentionally unresolvable modules, not real dependencies.
package tests

// ruleid: go-no-gorm
import "gorm.io/gorm"

// ruleid: go-no-wire-fx
import "go.uber.org/fx"

// ok: go-no-gorm
import "github.com/nydux/platform/libs/go/nyxpg"
