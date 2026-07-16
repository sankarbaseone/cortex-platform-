package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(wd, "testdata", "fixture")
}

func TestBuildRegistry_ParsesColumnsIndexesAndRLS(t *testing.T) {
	reg, err := BuildRegistry(fixtureRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	store := reg.Stores["postgres"]
	if store == nil {
		t.Fatal("expected postgres store")
	}

	widgets := store.Tables["widgets"]
	if widgets == nil {
		t.Fatal("expected widgets table")
	}
	if widgets.Owner != "svc-a" {
		t.Errorf("owner = %q, want svc-a", widgets.Owner)
	}
	if !widgets.RLS {
		t.Error("widgets should be RLS-enabled via the FOREACH ARRAY heuristic")
	}
	if len(widgets.Columns) != 4 {
		t.Fatalf("widgets columns = %d, want 4: %+v", len(widgets.Columns), widgets.Columns)
	}
	byName := map[string]Column{}
	for _, c := range widgets.Columns {
		byName[c.Name] = c
	}
	if byName["widget_id"].Nullable {
		t.Error("widget_id (PRIMARY KEY) should not be nullable")
	}
	if byName["tenant_id"].Nullable {
		t.Error("tenant_id (NOT NULL) should not be nullable")
	}
	if got := byName["status"].Check; !strings.Contains(got, "active") {
		t.Errorf("status.Check = %q, want it to contain the CHECK expression", got)
	}
	if len(widgets.Indexes) != 1 || widgets.Indexes[0].Name != "widgets_tenant_idx" {
		t.Errorf("widgets.Indexes = %+v, want [widgets_tenant_idx]", widgets.Indexes)
	}

	tags := store.Tables["widget_tags"]
	if tags == nil {
		t.Fatal("expected widget_tags table")
	}
	if tags.RLS {
		t.Error("widget_tags was never named in the RLS ARRAY literal - must not be RLS-enabled")
	}
}

func TestBuildRegistry_MissingServicesDirIsEmptyNotError(t *testing.T) {
	reg, err := BuildRegistry(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.Stores) != 0 {
		t.Errorf("expected empty registry, got %+v", reg.Stores)
	}
}

func TestRender_IsDeterministicAcrossRuns(t *testing.T) {
	root := fixtureRoot(t)
	reg1, err := BuildRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	reg2, err := BuildRegistry(root)
	if err != nil {
		t.Fatal(err)
	}
	if reg1.Render() != reg2.Render() {
		t.Error("Render() must be byte-identical across independent parses of the same input")
	}
}

// TestDriftCheck_DetectsIntentionalDesync is the concrete proof required by
// T-005's DoD ("drift check fails on intentional registry desync test"): it
// drives the same genDB/checkDB functions main() uses, end to end, against
// real files on disk.
func TestDriftCheck_DetectsIntentionalDesync(t *testing.T) {
	root := fixtureRoot(t)
	out := filepath.Join(t.TempDir(), "database-registry.yaml")

	if err := genDB(root, out); err != nil {
		t.Fatalf("genDB: %v", err)
	}
	if err := checkDB(root, out); err != nil {
		t.Fatalf("checkDB on a freshly generated file should pass: %v", err)
	}

	// Intentionally desync the committed registry: flip a fact (RLS) that
	// the migrations do not support, without touching the source SQL.
	generated, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	desynced := strings.Replace(string(generated), "rls: true", "rls: false", 1)
	if desynced == string(generated) {
		t.Fatal("test setup bug: desync edit did not change anything - fixture no longer contains rls: true")
	}
	if err := os.WriteFile(out, []byte(desynced), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := checkDB(root, out); err == nil {
		t.Fatal("checkDB must fail against an intentionally desynced registry")
	}

	// Regenerating (db-check's suggested fix) must clear the drift.
	if err := genDB(root, out); err != nil {
		t.Fatalf("genDB: %v", err)
	}
	if err := checkDB(root, out); err != nil {
		t.Fatalf("checkDB should pass again after regenerating: %v", err)
	}
}
