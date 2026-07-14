// Package config loads service configuration from the environment into a
// typed struct, per the platform-wide contract (ECD-002 internal/config.go
// row; RFC-014 G.0 "Config: env-only, typed struct, validated at boot,
// printed (secrets redacted) at startup"): env -> struct -> Validate() ->
// Redacted() printed. Each service's internal/config package wraps Load/
// MustLoad with its own concrete Config type, per ADR-0001's config-first
// wiring order.
package config

import (
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
)

// Validator is implemented by config structs that need fail-fast validation
// at boot (ECD-002: "Validate() error (fail-fast)").
type Validator interface {
	Validate() error
}

// Redactor is implemented by config structs that print a secret-redacted
// summary at startup (ECD-002: "Redacted() string").
type Redactor interface {
	Redacted() string
}

// Load parses environment variables (struct tag `envconfig:"NAME"`) into a
// new T and validates it if T implements Validator. It does not print
// anything; use MustLoad in cmd/<name>/main.go for the full boot contract.
func Load[T any]() (T, error) {
	var cfg T
	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, fmt.Errorf("config: load: %w", err)
	}
	if v, ok := any(&cfg).(Validator); ok {
		if err := v.Validate(); err != nil {
			return cfg, fmt.Errorf("config: validate: %w", err)
		}
	}
	return cfg, nil
}

// MustLoad implements the full boot contract: env -> struct -> Validate() ->
// Redacted() printed. It panics on load/validation failure, since a service
// cannot run without valid config (ADR-0001: config is wired first in main).
func MustLoad[T any]() T {
	cfg, err := Load[T]()
	if err != nil {
		panic(err)
	}
	if r, ok := any(&cfg).(Redactor); ok {
		log.Printf("config loaded: %s", r.Redacted())
	} else {
		log.Printf("config loaded: %+v", cfg)
	}
	return cfg
}
