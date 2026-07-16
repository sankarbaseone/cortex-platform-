// Package main implements genmanifests, the tool that derives AI-native
// registry YAMLs from source-of-truth code (ECD-001 tools/genmanifests).
//
// db-gen/db-check cover the database-registry per ECD-006: schema stays SQL,
// the registry is a generated view (store -> tables -> columns/indexes/owner/RLS).
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Column is one field of a table, as declared by a CREATE TABLE statement.
type Column struct {
	Name     string
	Type     string
	Nullable bool
	Check    string // raw CHECK expression attached to this column, if any
}

// Index is a CREATE INDEX statement targeting a table owned by this store.
type Index struct {
	Name    string
	Table   string
	Columns []string
	Unique  bool
}

// Table is one CREATE TABLE, plus everything genmanifests could infer about
// it from the rest of the migration (RLS, indexes).
type Table struct {
	Name    string
	Owner   string // service directory name the migration lives under
	RLS     bool
	Columns []Column
	Indexes []Index
}

// Store groups tables by backing engine (postgres, timescale, clickhouse),
// mirroring database-registry.yaml's top-level `stores` key.
type Store struct {
	Name   string
	Tables map[string]*Table
}

// Registry is the full generated database-registry.yaml payload.
type Registry struct {
	Stores map[string]*Store
}

func newRegistry() *Registry {
	return &Registry{Stores: map[string]*Store{}}
}

func (r *Registry) store(name string) *Store {
	s, ok := r.Stores[name]
	if !ok {
		s = &Store{Name: name, Tables: map[string]*Table{}}
		r.Stores[name] = s
	}
	return s
}

func (s *Store) table(name, owner string) *Table {
	t, ok := s.Tables[name]
	if !ok {
		t = &Table{Name: name, Owner: owner}
		s.Tables[name] = t
	}
	return t
}

// storeDirs are the migration engine subdirectories genmanifests understands.
// Only postgres/timescale use PostgreSQL-family DDL syntax; clickhouse's
// dialect differs enough that we record the table names it declares (for
// ownership/drift purposes) without attempting full column introspection.
var storeDirs = []string{"postgres", "timescale", "clickhouse"}

// BuildRegistry walks <root>/services/*/migrations/<store>/*.up.sql and
// parses each into the Registry. root is typically the repository root.
func BuildRegistry(root string) (*Registry, error) {
	reg := newRegistry()
	servicesDir := filepath.Join(root, "services")
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return reg, nil // nothing built yet - empty registry, not an error
		}
		return nil, fmt.Errorf("read %s: %w", servicesDir, err)
	}

	for _, svc := range entries {
		if !svc.IsDir() {
			continue
		}
		owner := svc.Name()
		for _, storeName := range storeDirs {
			migDir := filepath.Join(servicesDir, owner, "migrations", storeName)
			files, err := os.ReadDir(migDir)
			if err != nil {
				continue // service doesn't use this store
			}
			var names []string
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
					names = append(names, f.Name())
				}
			}
			sort.Strings(names) // apply in migration order (0001_, 0002_, ...)
			for _, name := range names {
				content, err := os.ReadFile(filepath.Join(migDir, name))
				if err != nil {
					return nil, fmt.Errorf("read %s: %w", name, err)
				}
				if storeName == "clickhouse" {
					applyClickhouse(reg.store(storeName), owner, string(content))
				} else {
					applyPostgresFamily(reg.store(storeName), owner, string(content))
				}
			}
		}
	}
	return reg, nil
}

// --- postgres/timescale parsing ---
//
// This is a deliberately narrow, line-oriented reader over the migration
// conventions actually in use (see services/tenant-svc/migrations/postgres/
// 0001_init.up.sql), not a general SQL parser: CREATE TABLE / CREATE INDEX
// statements, plus one heuristic for the FOREACH-over-ARRAY dynamic-SQL
// pattern used to enable RLS on a list of tables in one DO block. Anything
// outside that shape (generated columns, table inheritance, partitioned
// tables, etc.) is silently not represented - acceptable for a drift check
// whose job is catching accidental desync, not validating arbitrary SQL.

var (
	createTableRe  = regexp.MustCompile(`(?is)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?"?(\w+)"?\s*\((.*)\)\s*$`)
	createIndexRe  = regexp.MustCompile(`(?is)CREATE\s+(UNIQUE\s+)?INDEX\s+"?(\w+)"?\s+ON\s+"?(\w+)"?\s*\((.*?)\)`)
	directRLSRe    = regexp.MustCompile(`(?is)ALTER\s+TABLE\s+"?(\w+)"?\s+ENABLE\s+ROW\s+LEVEL\s+SECURITY`)
	arrayLiteralRe = regexp.MustCompile(`ARRAY\s*\[([^]]*)]`)
	arrayElemRe    = regexp.MustCompile(`'([^']*)'`)
	checkExprRe    = regexp.MustCompile(`(?is)CHECK\s*\((.*)\)`)
)

func applyPostgresFamily(store *Store, owner, sql string) {
	for _, stmt := range splitStatements(sql) {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		switch {
		case createTableRe.MatchString(trimmed):
			m := createTableRe.FindStringSubmatch(trimmed)
			t := store.table(m[1], owner)
			t.Columns = append(t.Columns, parseColumns(m[2])...)
		case createIndexRe.MatchString(trimmed):
			m := createIndexRe.FindStringSubmatch(trimmed)
			cols := splitTopLevel(m[4])
			for i := range cols {
				cols[i] = strings.TrimSpace(strings.Trim(cols[i], `"`))
			}
			idx := Index{Name: m[2], Table: m[3], Columns: cols, Unique: strings.TrimSpace(m[1]) != ""}
			if t, ok := store.Tables[m[3]]; ok {
				t.Indexes = append(t.Indexes, idx)
			}
		default:
			applyRLSHeuristics(store, trimmed)
		}
	}
}

func applyRLSHeuristics(store *Store, stmt string) {
	for _, m := range directRLSRe.FindAllStringSubmatch(stmt, -1) {
		if t, ok := store.Tables[m[1]]; ok {
			t.RLS = true
		}
	}
	if !strings.Contains(strings.ToUpper(stmt), "ENABLE ROW LEVEL SECURITY") {
		return
	}
	arr := arrayLiteralRe.FindStringSubmatch(stmt)
	if arr == nil {
		return
	}
	for _, e := range arrayElemRe.FindAllStringSubmatch(arr[1], -1) {
		if t, ok := store.Tables[e[1]]; ok {
			t.RLS = true
		}
	}
}

// clickhouse DDL (`ENGINE = MergeTree() ...`) is a different dialect; record
// table names/ownership only, per the narrowest-reading fallback in GAP-007.
var clickhouseTableRe = regexp.MustCompile(`(?is)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?"?(\w+)"?`)

func applyClickhouse(store *Store, owner, sql string) {
	for _, stmt := range splitStatements(sql) {
		m := clickhouseTableRe.FindStringSubmatch(stmt)
		if m == nil {
			continue
		}
		store.table(m[1], owner)
	}
}

func parseColumns(body string) []Column {
	var cols []Column
	skipKeywords := []string{"PRIMARY KEY", "UNIQUE", "FOREIGN KEY", "CONSTRAINT", "CHECK", "EXCLUDE"}
	for _, seg := range splitTopLevel(body) {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		upper := strings.ToUpper(seg)
		isTableConstraint := false
		for _, kw := range skipKeywords {
			if strings.HasPrefix(upper, kw) {
				isTableConstraint = true
				break
			}
		}
		if isTableConstraint {
			continue
		}
		fields := strings.Fields(seg)
		if len(fields) < 2 {
			continue
		}
		name := strings.Trim(fields[0], `"`)
		typ := fields[1]
		nullable := !strings.Contains(upper, "NOT NULL") && !strings.Contains(upper, "PRIMARY KEY")
		check := ""
		if cm := checkExprRe.FindStringSubmatch(seg); cm != nil {
			check = strings.TrimSpace(cm[1])
		}
		cols = append(cols, Column{Name: name, Type: typ, Nullable: nullable, Check: check})
	}
	return cols
}

// splitTopLevel splits on commas that are not nested inside parens (so
// `CHECK (status IN ('a','b'))` stays one segment).
func splitTopLevel(s string) []string {
	var out []string
	depth := 0
	last := 0
	for i, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, s[last:i])
				last = i + 1
			}
		}
	}
	out = append(out, s[last:])
	return out
}

// splitStatements splits a migration file into top-level SQL statements,
// treating `$$ ... $$` dollar-quoted bodies (DO blocks) as opaque so that
// semicolons inside them don't end the statement early.
func splitStatements(sql string) []string {
	// strip line comments
	var lines []string
	for _, line := range strings.Split(sql, "\n") {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}
		lines = append(lines, line)
	}
	cleaned := strings.Join(lines, "\n")

	var stmts []string
	dollarQuoted := false
	last := 0
	for i := 0; i < len(cleaned); i++ {
		if strings.HasPrefix(cleaned[i:], "$$") {
			dollarQuoted = !dollarQuoted
			i++ // consume second $
			continue
		}
		if cleaned[i] == ';' && !dollarQuoted {
			stmts = append(stmts, cleaned[last:i])
			last = i + 1
		}
	}
	if strings.TrimSpace(cleaned[last:]) != "" {
		stmts = append(stmts, cleaned[last:])
	}
	return stmts
}

// --- rendering ---

// Render produces deterministic YAML for the registry: sorted store names,
// sorted table names within a store, columns in source (declaration) order,
// indexes sorted by name. Determinism is what makes db-check a meaningful
// drift check rather than a source of flaky diffs.
func (r *Registry) Render() string {
	var b strings.Builder
	b.WriteString("# GENERATED by tools/genmanifests db-gen. Do not hand-edit; see ECD-006.\n")
	b.WriteString("version: 1\n")
	b.WriteString("stores:\n")
	for _, storeName := range sortedKeys(r.Stores) {
		store := r.Stores[storeName]
		fmt.Fprintf(&b, "  %s:\n", storeName)
		if len(store.Tables) == 0 {
			b.WriteString("    tables: {}\n")
			continue
		}
		b.WriteString("    tables:\n")
		for _, tableName := range sortedKeysT(store.Tables) {
			t := store.Tables[tableName]
			fmt.Fprintf(&b, "      %s:\n", tableName)
			fmt.Fprintf(&b, "        owner: %s\n", t.Owner)
			fmt.Fprintf(&b, "        rls: %t\n", t.RLS)
			if len(t.Columns) == 0 {
				b.WriteString("        columns: []\n")
			} else {
				b.WriteString("        columns:\n")
				for _, c := range t.Columns {
					fmt.Fprintf(&b, "          - {name: %s, type: %s, nullable: %t", c.Name, c.Type, c.Nullable)
					if c.Check != "" {
						fmt.Fprintf(&b, ", check: %q", c.Check)
					}
					b.WriteString("}\n")
				}
			}
			idx := append([]Index(nil), t.Indexes...)
			sort.Slice(idx, func(i, j int) bool { return idx[i].Name < idx[j].Name })
			if len(idx) == 0 {
				b.WriteString("        indexes: []\n")
			} else {
				b.WriteString("        indexes:\n")
				for _, ix := range idx {
					fmt.Fprintf(&b, "          - {name: %s, columns: [%s], unique: %t}\n",
						ix.Name, strings.Join(ix.Columns, ", "), ix.Unique)
				}
			}
		}
	}
	return b.String()
}

func sortedKeys(m map[string]*Store) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedKeysT(m map[string]*Table) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
