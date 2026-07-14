package config

import (
	"errors"
	"testing"
)

type testConfig struct {
	Host   string `envconfig:"HOST"`
	Port   int    `envconfig:"PORT" default:"8080"`
	Secret string `envconfig:"SECRET"`
}

func (c *testConfig) Validate() error {
	if c.Host == "" {
		return errors.New("HOST is required")
	}
	return nil
}

func (c *testConfig) Redacted() string {
	secret := "<empty>"
	if c.Secret != "" {
		secret = "<redacted>"
	}
	return "Host=" + c.Host + " Secret=" + secret
}

func TestLoad_ParsesEnvAndValidates(t *testing.T) {
	t.Setenv("HOST", "db.internal")
	t.Setenv("SECRET", "s3cr3t")

	cfg, err := Load[testConfig]()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Host != "db.internal" {
		t.Fatalf("Host = %q", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port default = %d, want 8080", cfg.Port)
	}
}

func TestLoad_ValidateFailurePropagates(t *testing.T) {
	_, err := Load[testConfig]()
	if err == nil {
		t.Fatalf("expected validation error for missing HOST")
	}
}

func TestMustLoad_PanicsOnInvalid(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected MustLoad to panic on invalid config")
		}
	}()
	MustLoad[testConfig]()
}

func TestMustLoad_RedactsSecretOnPrint(t *testing.T) {
	t.Setenv("HOST", "db.internal")
	t.Setenv("SECRET", "s3cr3t")
	cfg := MustLoad[testConfig]()
	redacted := cfg.Redacted()
	if redacted == "" {
		t.Fatalf("Redacted() empty")
	}
	if want := "Secret=<redacted>"; !contains(redacted, want) {
		t.Fatalf("Redacted() = %q, want it to contain %q (never the raw secret)", redacted, want)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
