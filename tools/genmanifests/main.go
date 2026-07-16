package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const defaultOut = "tools/genmanifests/registries/database-registry.yaml"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out := filepath.Join(root, defaultOut)

	switch os.Args[1] {
	case "db-gen":
		if err := genDB(root, out); err != nil {
			fmt.Fprintln(os.Stderr, "db-gen:", err)
			os.Exit(1)
		}
		fmt.Println("wrote", defaultOut)
	case "db-check":
		if err := checkDB(root, out); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("db-check: ok,", defaultOut, "matches migrations")
	default:
		usage()
		os.Exit(2)
	}
}

// genDB regenerates the registry file at out from the migrations under root.
func genDB(root, out string) error {
	reg, err := BuildRegistry(root)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(out, []byte(reg.Render()), 0o644)
}

// checkDB is the drift check: it fails if the committed file at out no
// longer matches what BuildRegistry(root) would generate right now.
func checkDB(root, out string) error {
	reg, err := BuildRegistry(root)
	if err != nil {
		return fmt.Errorf("db-check: %w", err)
	}
	want := reg.Render()
	got, err := os.ReadFile(out)
	if err != nil {
		return fmt.Errorf("db-check: read %s: %w", out, err)
	}
	if want != string(got) {
		return fmt.Errorf("db-check: DRIFT DETECTED - %s is out of sync with services/*/migrations/**.\nRun: go run ./tools/genmanifests db-gen", out)
	}
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: genmanifests <db-gen|db-check>")
}
