package memory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	_ "modernc.org/sqlite"
)

// ─────────────────────────────────────────────────────────────
// Test Helpers
// ─────────────────────────────────────────────────────────────

func setupGraph(t *testing.T) (*Graph, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_graph.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	g := NewGraph(db)
	if g == nil {
		t.Fatal("NewGraph returned nil")
	}
	if err := g.InitDB(); err != nil {
		t.Fatalf("InitDB: %v", err)
	}

	cleanup := func() {
		db.Close()
	}
	return g, cleanup
}

func setupGraphWithFile(t *testing.T) (*Graph, string, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_graph.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	g := NewGraph(db)
	if g == nil {
		t.Fatal("NewGraph returned nil")
	}
	if err := g.InitDB(); err != nil {
		t.Fatalf("InitDB: %v", err)
	}

	cleanup := func() {
		db.Close()
	}
	return g, dbPath, cleanup
}

// ─────────────────────────────────────────────────────────────
// 1. NewGraph with nil DB
// ─────────────────────────────────────────────────────────────

func TestNewGraphNilDB(t *testing.T) {
	g := NewGraph(nil)
	if g != nil {
		t.Error("expected nil Graph for nil DB")
	}
}

// ─────────────────────────────────────────────────────────────
// 2. InitDB creates tables
// ─────────────────────────────────────────────────────────────

func TestInitDBCreatesTables(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// Verify tables exist by checking schema
	var count int
	err := g.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('entities','triples','attributes')`).Scan(&count)
	if err != nil {
		t.Fatalf("query schema: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 tables, got %d", count)
	}
}

// ─────────────────────────────────────────────────────────────
// 3. AddEntity basic
// ─────────────────────────────────────────────────────────────

func TestAddEntityBasic(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	e := Entity{
		Name: "Alice",
		Type: "person",
	}
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("AddEntity: %v", err)
	}

	got, err := g.GetEntity("Alice")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}
	if got.Name != "Alice" {
		t.Errorf("expected name Alice, got %q", got.Name)
	}
	if got.Type != "person" {
		t.Errorf("expected type person, got %q", got.Type)
	}
}

// ─────────────────────────────────────────────────────────────
// 4. AddEntity idempotent (upsert)
// ─────────────────────────────────────────────────────────────

func TestAddEntityUpsert(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	e := Entity{Name: "Bob", Type: "person"}
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("first AddEntity: %v", err)
	}

	// Update type
	e.Type = "organization"
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("second AddEntity: %v", err)
	}

	got, err := g.GetEntity("Bob")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}
	if got.Type != "organization" {
		t.Errorf("expected type organization after upsert, got %q", got.Type)
	}
}

// ─────────────────────────────────────────────────────────────
// 5. AddEntity with attributes
// ─────────────────────────────────────────────────────────────

func TestAddEntityWithAttributes(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	e := Entity{
		Name: "ProjectX",
		Type: "project",
		Attributes: map[string]string{
			"language": "Go",
			"version":  "1.0.0",
		},
	}
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("AddEntity: %v", err)
	}

	got, err := g.GetEntity("ProjectX")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}
	if got.Attributes["language"] != "Go" {
		t.Errorf("expected language Go, got %q", got.Attributes["language"])
	}
	if got.Attributes["version"] != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %q", got.Attributes["version"])
	}
}

// ─────────────────────────────────────────────────────────────
// 6. GetEntity not found
// ─────────────────────────────────────────────────────────────

func TestGetEntityNotFound(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	_, err := g.GetEntity("NonExistent")
	if err == nil {
		t.Error("expected error for nonexistent entity")
	}
}

// ─────────────────────────────────────────────────────────────
// 7. FindEntities by type
// ─────────────────────────────────────────────────────────────

func TestFindEntitiesByType(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	entities := []Entity{
		{Name: "Alice", Type: "person"},
		{Name: "Bob", Type: "person"},
		{Name: "ProjectX", Type: "project"},
	}
	for _, e := range entities {
		if err := g.AddEntity(e); err != nil {
			t.Fatalf("AddEntity %s: %v", e.Name, err)
		}
	}

	people, err := g.FindEntities("person")
	if err != nil {
		t.Fatalf("FindEntities: %v", err)
	}
	if len(people) != 2 {
		t.Errorf("expected 2 people, got %d", len(people))
	}

	all, err := g.FindEntities("")
	if err != nil {
		t.Fatalf("FindEntities all: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 total entities, got %d", len(all))
	}
}

// ─────────────────────────────────────────────────────────────
// 8. AddAlias
// ─────────────────────────────────────────────────────────────

func TestAddAlias(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	if err := g.AddEntity(Entity{Name: "Alice", Type: "person"}); err != nil {
		t.Fatalf("AddEntity: %v", err)
	}

	if err := g.AddAlias("Alice", "Alicia"); err != nil {
		t.Fatalf("AddAlias: %v", err)
	}

	got, err := g.GetEntity("Alice")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}

	found := false
	for _, a := range got.Aliases {
		if a == "Alicia" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected alias 'Alicia', got %v", got.Aliases)
	}
}

// ─────────────────────────────────────────────────────────────
// 9. AddAlias idempotent
// ─────────────────────────────────────────────────────────────

func TestAddAliasIdempotent(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	if err := g.AddEntity(Entity{Name: "Bob", Type: "person"}); err != nil {
		t.Fatalf("AddEntity: %v", err)
	}

	if err := g.AddAlias("Bob", "Robert"); err != nil {
		t.Fatalf("first AddAlias: %v", err)
	}
	if err := g.AddAlias("Bob", "Robert"); err != nil {
		t.Fatalf("second AddAlias (idempotent): %v", err)
	}

	got, err := g.GetEntity("Bob")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}

	count := 0
	for _, a := range got.Aliases {
		if a == "Robert" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected alias 'Robert' exactly once, got %v", got.Aliases)
	}
}

// ─────────────────────────────────────────────────────────────
// 10. AddAlias for nonexistent entity
// ─────────────────────────────────────────────────────────────

func TestAddAliasNotFound(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	err := g.AddAlias("NonExistent", "alias")
	if err == nil {
		t.Error("expected error for alias on nonexistent entity")
	}
}

// ─────────────────────────────────────────────────────────────
// 11. SetAttribute
// ─────────────────────────────────────────────────────────────

func TestSetAttribute(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	if err := g.AddEntity(Entity{Name: "ProjectX", Type: "project"}); err != nil {
		t.Fatalf("AddEntity: %v", err)
	}

	if err := g.SetAttribute("ProjectX", "language", "Go"); err != nil {
		t.Fatalf("SetAttribute: %v", err)
	}

	got, err := g.GetEntity("ProjectX")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}
	if got.Attributes["language"] != "Go" {
		t.Errorf("expected language Go, got %q", got.Attributes["language"])
	}

	// Update attribute
	if err := g.SetAttribute("ProjectX", "language", "Rust"); err != nil {
		t.Fatalf("SetAttribute update: %v", err)
	}

	got2, err := g.GetEntity("ProjectX")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}
	if got2.Attributes["language"] != "Rust" {
		t.Errorf("expected language Rust after update, got %q", got2.Attributes["language"])
	}
}

// ─────────────────────────────────────────────────────────────
// 12. AddTriple & Query
// ─────────────────────────────────────────────────────────────

func TestAddTripleAndQuery(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	triple := Triple{
		Subject:    "Alice",
		Predicate:  "works_on",
		Object:     "ProjectX",
		Confidence: 0.9,
	}
	if err := g.AddTriple(triple); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}

	// Query by subject
	results, err := g.Query("Alice", "", "")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 triple, got %d", len(results))
	}
	if results[0].Predicate != "works_on" {
		t.Errorf("expected predicate works_on, got %q", results[0].Predicate)
	}

	// Query by predicate
	results2, err := g.Query("", "works_on", "")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results2) != 1 {
		t.Errorf("expected 1 triple by predicate, got %d", len(results2))
	}

	// Query all (wildcard)
	results3, err := g.Query("", "", "")
	if err != nil {
		t.Fatalf("Query all: %v", err)
	}
	if len(results3) != 1 {
		t.Errorf("expected 1 triple wildcard, got %d", len(results3))
	}
}

// ─────────────────────────────────────────────────────────────
// 13. DeleteTriple
// ─────────────────────────────────────────────────────────────

func TestDeleteTriple(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	tID := "test_triple_001"
	if err := g.AddTriple(Triple{ID: tID, Subject: "A", Predicate: "rel", Object: "B"}); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}

	if err := g.DeleteTriple(tID); err != nil {
		t.Fatalf("DeleteTriple: %v", err)
	}

	results, err := g.Query("", "", "")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 triples after delete, got %d", len(results))
	}
}

// ─────────────────────────────────────────────────────────────
// 14. DeleteTriple not found
// ─────────────────────────────────────────────────────────────

func TestDeleteTripleNotFound(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	err := g.DeleteTriple("nonexistent")
	if err == nil {
		t.Error("expected error deleting nonexistent triple")
	}
}

// ─────────────────────────────────────────────────────────────
// 15. QueryValid — temporal validity
// ─────────────────────────────────────────────────────────────

func TestQueryValidTemporal(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// Triple valid from 2026-01-01 to 2026-06-30
	if err := g.AddTriple(Triple{
		Subject:   "Alice",
		Predicate: "located_at",
		Object:    "New York",
		ValidFrom: "2026-01-01",
		ValidTo:   "2026-06-30",
	}); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}

	// Triple valid from 2026-07-01 onward (no end)
	if err := g.AddTriple(Triple{
		Subject:   "Alice",
		Predicate: "located_at",
		Object:    "London",
		ValidFrom: "2026-07-01",
	}); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}

	// Query at 2026-03-15 — should return New York
	mar, err := g.QueryValid("2026-03-15")
	if err != nil {
		t.Fatalf("QueryValid: %v", err)
	}
	if len(mar) != 1 {
		t.Fatalf("expected 1 valid triple in March, got %d", len(mar))
	}
	if mar[0].Object != "New York" {
		t.Errorf("expected New York, got %q", mar[0].Object)
	}

	// Query at 2026-08-01 — should return London
	aug, err := g.QueryValid("2026-08-01")
	if err != nil {
		t.Fatalf("QueryValid: %v", err)
	}
	if len(aug) != 1 {
		t.Fatalf("expected 1 valid triple in August, got %d", len(aug))
	}
	if aug[0].Object != "London" {
		t.Errorf("expected London, got %q", aug[0].Object)
	}

	// Query at 2025-12-01 — before any validity
	dec, err := g.QueryValid("2025-12-01")
	if err != nil {
		t.Fatalf("QueryValid: %v", err)
	}
	if len(dec) != 0 {
		t.Errorf("expected 0 valid triples in Dec 2025, got %d", len(dec))
	}
}

// ─────────────────────────────────────────────────────────────
// 16. GetRelated
// ─────────────────────────────────────────────────────────────

func TestGetRelated(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	triples := []Triple{
		{Subject: "Alice", Predicate: "works_on", Object: "ProjectX"},
		{Subject: "Alice", Predicate: "manages", Object: "TeamA"},
		{Subject: "Bob", Predicate: "works_on", Object: "Alice"},
	}
	for _, tr := range triples {
		if err := g.AddTriple(tr); err != nil {
			t.Fatalf("AddTriple: %v", err)
		}
	}

	related, err := g.GetRelated("Alice")
	if err != nil {
		t.Fatalf("GetRelated: %v", err)
	}
	if len(related) != 3 {
		t.Errorf("expected 3 related triples for Alice, got %d", len(related))
	}
}

// ─────────────────────────────────────────────────────────────
// 17. BFS Path Finding — basic path
// ─────────────────────────────────────────────────────────────

func TestGetPathBasic(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// A -> B -> C -> D
	triples := []Triple{
		{Subject: "A", Predicate: "knows", Object: "B"},
		{Subject: "B", Predicate: "knows", Object: "C"},
		{Subject: "C", Predicate: "knows", Object: "D"},
	}
	for _, tr := range triples {
		if err := g.AddTriple(tr); err != nil {
			t.Fatalf("AddTriple: %v", err)
		}
	}

	paths, err := g.GetPath("A", "D", 5)
	if err != nil {
		t.Fatalf("GetPath: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("expected at least 1 path from A to D")
	}

	// First path should be A -> B -> C -> D
	path := paths[0]
	if len(path) != 4 {
		t.Errorf("expected path length 4, got %d: %v", len(path), path)
	}
	expected := []string{"A", "B", "C", "D"}
	for i, v := range expected {
		if path[i] != v {
			t.Errorf("path[%d] = %q, expected %q", i, path[i], v)
		}
	}
}

// ─────────────────────────────────────────────────────────────
// 18. BFS Path Finding — no path
// ─────────────────────────────────────────────────────────────

func TestGetPathNoPath(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	triples := []Triple{
		{Subject: "A", Predicate: "knows", Object: "B"},
		{Subject: "C", Predicate: "knows", Object: "D"},
	}
	for _, tr := range triples {
		if err := g.AddTriple(tr); err != nil {
			t.Fatalf("AddTriple: %v", err)
		}
	}

	paths, err := g.GetPath("A", "D", 5)
	if err != nil {
		t.Fatalf("GetPath: %v", err)
	}
	if paths != nil {
		t.Errorf("expected nil paths, got %v", paths)
	}
}

// ─────────────────────────────────────────────────────────────
// 19. BFS Path Finding — maxHops limit
// ─────────────────────────────────────────────────────────────

func TestGetPathMaxHops(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// A -> B -> C -> D
	triples := []Triple{
		{Subject: "A", Predicate: "knows", Object: "B"},
		{Subject: "B", Predicate: "knows", Object: "C"},
		{Subject: "C", Predicate: "knows", Object: "D"},
	}
	for _, tr := range triples {
		if err := g.AddTriple(tr); err != nil {
			t.Fatalf("AddTriple: %v", err)
		}
	}

	// maxHops=2 should not reach D (needs 3 hops)
	paths, err := g.GetPath("A", "D", 2)
	if err != nil {
		t.Fatalf("GetPath: %v", err)
	}
	if paths != nil {
		t.Errorf("expected nil paths with maxHops=2, got %v", paths)
	}

	// maxHops=3 should find it
	paths3, err := g.GetPath("A", "D", 3)
	if err != nil {
		t.Fatalf("GetPath: %v", err)
	}
	if len(paths3) == 0 {
		t.Error("expected paths with maxHops=3")
	}
}

// ─────────────────────────────────────────────────────────────
// 20. BFS Path Finding — multiple paths
// ─────────────────────────────────────────────────────────────

func TestGetPathMultiplePaths(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// A -> B -> D and A -> C -> D
	triples := []Triple{
		{Subject: "A", Predicate: "knows", Object: "B"},
		{Subject: "A", Predicate: "knows", Object: "C"},
		{Subject: "B", Predicate: "knows", Object: "D"},
		{Subject: "C", Predicate: "knows", Object: "D"},
	}
	for _, tr := range triples {
		if err := g.AddTriple(tr); err != nil {
			t.Fatalf("AddTriple: %v", err)
		}
	}

	paths, err := g.GetPath("A", "D", 3)
	if err != nil {
		t.Fatalf("GetPath: %v", err)
	}
	if len(paths) < 2 {
		t.Errorf("expected at least 2 paths, got %d", len(paths))
	}

	// Collect and sort path strings for comparison
	pathStrs := make([]string, len(paths))
	for i, p := range paths {
		pathStrs[i] = joinPath(p)
	}
	sort.Strings(pathStrs)

	if pathStrs[0] != "A-B-D" && pathStrs[0] != "A-C-D" {
		t.Errorf("unexpected path: %s", pathStrs[0])
	}
}

func joinPath(path []string) string {
	result := ""
	for i, p := range path {
		if i > 0 {
			result += "-"
		}
		result += p
	}
	return result
}

// ─────────────────────────────────────────────────────────────
// 21. Export/Import round-trip
// ─────────────────────────────────────────────────────────────

func TestExportImportRoundTrip(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// Add entities
	entities := []Entity{
		{Name: "Alice", Type: "person", Attributes: map[string]string{"role": "engineer"}},
		{Name: "ProjectX", Type: "project", Attributes: map[string]string{"lang": "Go"}},
	}
	for _, e := range entities {
		if err := g.AddEntity(e); err != nil {
			t.Fatalf("AddEntity %s: %v", e.Name, err)
		}
	}

	// Add alias
	if err := g.AddAlias("Alice", "Alicia"); err != nil {
		t.Fatalf("AddAlias: %v", err)
	}

	// Add triples
	triples := []Triple{
		{Subject: "Alice", Predicate: "works_on", Object: "ProjectX", Confidence: 0.95},
		{Subject: "Alice", Predicate: "located_at", Object: "New York", ValidFrom: "2026-01-01", ValidTo: "2026-12-31"},
	}
	for _, tr := range triples {
		if err := g.AddTriple(tr); err != nil {
			t.Fatalf("AddTriple: %v", err)
		}
	}

	// Export
	data, err := g.Export()
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	// Verify JSON is valid
	var exported GraphExport
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("unmarshal export: %v", err)
	}
	if len(exported.Entities) != 2 {
		t.Errorf("expected 2 entities in export, got %d", len(exported.Entities))
	}
	if len(exported.Triples) != 2 {
		t.Errorf("expected 2 triples in export, got %d", len(exported.Triples))
	}

	// Import into a NEW graph
	g2, cleanup2 := setupGraph(t)
	defer cleanup2()

	if err := g2.Import(data); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// Verify imported data
	gotEntity, err := g2.GetEntity("Alice")
	if err != nil {
		t.Fatalf("GetEntity Alice after import: %v", err)
	}
	if gotEntity.Type != "person" {
		t.Errorf("expected type person, got %q", gotEntity.Type)
	}
	if gotEntity.Attributes["role"] != "engineer" {
		t.Errorf("expected role engineer, got %q", gotEntity.Attributes["role"])
	}

	foundAlias := false
	for _, a := range gotEntity.Aliases {
		if a == "Alicia" {
			foundAlias = true
			break
		}
	}
	if !foundAlias {
		t.Errorf("expected alias 'Alicia' after import, got %v", gotEntity.Aliases)
	}

	importedTriples, err := g2.Query("", "", "")
	if err != nil {
		t.Fatalf("Query after import: %v", err)
	}
	if len(importedTriples) != 2 {
		t.Errorf("expected 2 triples after import, got %d", len(importedTriples))
	}
}

// ─────────────────────────────────────────────────────────────
// 22. Stats accuracy
// ─────────────────────────────────────────────────────────────

func TestStatsAccuracy(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// Add known entities and triples
	entities := []Entity{
		{Name: "Alice", Type: "person"},
		{Name: "Bob", Type: "person"},
		{Name: "ProjectX", Type: "project"},
	}
	for _, e := range entities {
		if err := g.AddEntity(e); err != nil {
			t.Fatalf("AddEntity: %v", e.Name)
		}
	}

	triples := []Triple{
		{Subject: "Alice", Predicate: "works_on", Object: "ProjectX"},
		{Subject: "Bob", Predicate: "works_on", Object: "ProjectX"},
		{Subject: "Alice", Predicate: "knows", Object: "Bob"},
	}
	for _, tr := range triples {
		if err := g.AddTriple(tr); err != nil {
			t.Fatalf("AddTriple: %v", err)
		}
	}

	stats, err := g.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}

	if stats["total_entities"] != 3 {
		t.Errorf("expected 3 total entities, got %d", stats["total_entities"])
	}
	if stats["total_triples"] != 3 {
		t.Errorf("expected 3 total triples, got %d", stats["total_triples"])
	}
	if stats["entity_type:person"] != 2 {
		t.Errorf("expected 2 person entities, got %d", stats["entity_type:person"])
	}
	if stats["entity_type:project"] != 1 {
		t.Errorf("expected 1 project entity, got %d", stats["entity_type:project"])
	}
	if stats["predicate:works_on"] != 2 {
		t.Errorf("expected 2 works_on triples, got %d", stats["predicate:works_on"])
	}
	if stats["predicate:knows"] != 1 {
		t.Errorf("expected 1 knows triple, got %d", stats["predicate:knows"])
	}
}

// ─────────────────────────────────────────────────────────────
// 23. Concurrent access — 50 goroutines
// ─────────────────────────────────────────────────────────────

func TestConcurrentAccess(t *testing.T) {
	g, _, cleanup := setupGraphWithFile(t)
	defer cleanup()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()

			name := fmt.Sprintf("Entity_%d", id)
			e := Entity{Name: name, Type: "concept"}
			if err := g.AddEntity(e); err != nil {
				errCh <- fmt.Errorf("goroutine %d AddEntity: %w", id, err)
				return
			}

			triple := Triple{
				Subject:   name,
				Predicate: "related_to",
				Object:    "Root",
			}
			if err := g.AddTriple(triple); err != nil {
				errCh <- fmt.Errorf("goroutine %d AddTriple: %w", id, err)
				return
			}

			// Read back
			_, err := g.GetEntity(name)
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d GetEntity: %w", id, err)
				return
			}

			// Query
			_, err = g.Query(name, "", "")
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d Query: %w", id, err)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		t.Errorf("%d concurrent errors:", len(errors))
		for i, err := range errors {
			if i < 5 { // Show first 5
				t.Logf("  %v", err)
			}
		}
	}

	// Verify final count
	stats, err := g.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats["total_entities"] != goroutines {
		t.Errorf("expected %d entities after concurrent writes, got %d", goroutines, stats["total_entities"])
	}
	if stats["total_triples"] != goroutines {
		t.Errorf("expected %d triples after concurrent writes, got %d", goroutines, stats["total_triples"])
	}
}

// ─────────────────────────────────────────────────────────────
// 24. DetectEntity — person
// ─────────────────────────────────────────────────────────────

func TestDetectEntityPerson(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"email", "alice@example.com", "person"},
		{"title_case_name", "John Smith", "person"},
		{"three_word_name", "Nguyen Van An", "person"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectEntity(tc.input)
			if got != tc.expected {
				t.Errorf("DetectEntity(%q) = %q, expected %q", tc.input, got, tc.expected)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// 25. DetectEntity — project, task, location, org, event
// ─────────────────────────────────────────────────────────────

func TestDetectEntityVarious(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my/project", "project"},
		{"github.com/user/repo", "project"},
		{"TODO_fix_bug", "task"},
		{"FIXME_crash", "task"},
		{"task#123", "task"},
		{"London", "location"},
		{"Tokyo", "location"},
		{"Acme Inc", "organization"},
		{"TechCorp LLC", "organization"},
		{"DevConf 2026 conference", "event"},
		{"hackathon_may", "event"},
		{"random_thing", "concept"},
		{"", "concept"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := DetectEntity(tc.input)
			if got != tc.expected {
				t.Errorf("DetectEntity(%q) = %q, expected %q", tc.input, got, tc.expected)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// 26. Triple auto-ID is deterministic
// ─────────────────────────────────────────────────────────────

func TestTripleAutoID(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	t1 := Triple{Subject: "A", Predicate: "rel", Object: "B"}
	if err := g.AddTriple(t1); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}

	results, err := g.Query("A", "rel", "B")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 triple, got %d", len(results))
	}

	// Re-add same triple (should overwrite due to same ID)
	t2 := Triple{Subject: "A", Predicate: "rel", Object: "B", Confidence: 0.5}
	if err := g.AddTriple(t2); err != nil {
		t.Fatalf("AddTriple again: %v", err)
	}

	results2, err := g.Query("A", "rel", "B")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results2) != 1 {
		t.Errorf("expected 1 triple after re-add, got %d", len(results2))
	}
	if results2[0].Confidence != 0.5 {
		t.Errorf("expected confidence 0.5, got %f", results2[0].Confidence)
	}
}

// ─────────────────────────────────────────────────────────────
// 27. Entity auto-type detection in AddEntity
// ─────────────────────────────────────────────────────────────

func TestAddEntityAutoType(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	e := Entity{Name: "John Smith"} // no type specified
	if err := g.AddEntity(e); err != nil {
		t.Fatalf("AddEntity: %v", err)
	}

	got, err := g.GetEntity("John Smith")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}
	if got.Type != "person" {
		t.Errorf("expected auto-detected type 'person', got %q", got.Type)
	}
}

// ─────────────────────────────────────────────────────────────
// 28. Query with exact match on all fields
// ─────────────────────────────────────────────────────────────

func TestQueryExactMatch(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	triples := []Triple{
		{Subject: "Alice", Predicate: "works_on", Object: "ProjectX"},
		{Subject: "Alice", Predicate: "manages", Object: "TeamA"},
		{Subject: "Bob", Predicate: "works_on", Object: "ProjectX"},
	}
	for _, tr := range triples {
		if err := g.AddTriple(tr); err != nil {
			t.Fatalf("AddTriple: %v", err)
		}
	}

	// Query with all fields specified
	results, err := g.Query("Alice", "works_on", "ProjectX")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 exact match, got %d", len(results))
	}

	// Query with subject + predicate
	results2, err := g.Query("Alice", "works_on", "")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results2) != 1 {
		t.Errorf("expected 1 triple for Alice+works_on, got %d", len(results2))
	}
}

// ─────────────────────────────────────────────────────────────
// 29. GetEntity case-insensitive lookup
// ─────────────────────────────────────────────────────────────

func TestGetEntityCaseInsensitive(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	if err := g.AddEntity(Entity{Name: "Alice", Type: "person"}); err != nil {
		t.Fatalf("AddEntity: %v", err)
	}

	// Lookup with different case
	got, err := g.GetEntity("alice")
	if err != nil {
		t.Fatalf("GetEntity lowercase: %v", err)
	}
	if got.Name != "Alice" {
		t.Errorf("expected name Alice, got %q", got.Name)
	}
}

// ─────────────────────────────────────────────────────────────
// 30. Export on empty graph
// ─────────────────────────────────────────────────────────────

func TestExportEmptyGraph(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	data, err := g.Export()
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	var exp GraphExport
	if err := json.Unmarshal(data, &exp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(exp.Entities) != 0 {
		t.Errorf("expected 0 entities, got %d", len(exp.Entities))
	}
	if len(exp.Triples) != 0 {
		t.Errorf("expected 0 triples, got %d", len(exp.Triples))
	}
}

// ─────────────────────────────────────────────────────────────
// 31. SetAttribute on nonexistent entity
// ─────────────────────────────────────────────────────────────

func TestSetAttributeNotFound(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	err := g.SetAttribute("NonExistent", "key", "value")
	if err == nil {
		t.Error("expected error setting attribute on nonexistent entity")
	}
}

// ─────────────────────────────────────────────────────────────
// 32. Import invalid JSON
// ─────────────────────────────────────────────────────────────

func TestImportInvalidJSON(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	err := g.Import([]byte(`{invalid json`))
	if err == nil {
		t.Error("expected error importing invalid JSON")
	}
}

// ─────────────────────────────────────────────────────────────
// 33. Confidence clamping
// ─────────────────────────────────────────────────────────────

func TestTripleConfidenceClamping(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// Confidence 0 should clamp to 1.0
	if err := g.AddTriple(Triple{Subject: "A", Predicate: "rel", Object: "B", Confidence: 0}); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}
	results, _ := g.Query("A", "rel", "B")
	if len(results) != 1 {
		t.Fatalf("expected 1 triple")
	}
	if results[0].Confidence != 1.0 {
		t.Errorf("expected clamped confidence 1.0, got %f", results[0].Confidence)
	}

	// Confidence > 1 should clamp to 1.0
	if err := g.AddTriple(Triple{Subject: "C", Predicate: "rel", Object: "D", Confidence: 1.5}); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}
	results2, _ := g.Query("C", "rel", "D")
	if len(results2) != 1 {
		t.Fatalf("expected 1 triple")
	}
	if results2[0].Confidence != 1.0 {
		t.Errorf("expected clamped confidence 1.0, got %f", results2[0].Confidence)
	}
}

// ─────────────────────────────────────────────────────────────
// 34. GetPath maxHops=0 returns nil
// ─────────────────────────────────────────────────────────────

func TestGetPathZeroHops(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	if err := g.AddTriple(Triple{Subject: "A", Predicate: "rel", Object: "B"}); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}

	paths, err := g.GetPath("A", "B", 0)
	if err != nil {
		t.Fatalf("GetPath: %v", err)
	}
	if paths != nil {
		t.Errorf("expected nil for maxHops=0, got %v", paths)
	}
}

// ─────────────────────────────────────────────────────────────
// 35. Persistent storage across connections
// ─────────────────────────────────────────────────────────────

func TestPersistentStorage(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "persistent.db")

	// First connection: add data
	db1, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db1: %v", err)
	}
	g1 := NewGraph(db1)
	if err := g1.InitDB(); err != nil {
		t.Fatalf("InitDB1: %v", err)
	}
	if err := g1.AddEntity(Entity{Name: "PersistentEntity", Type: "concept"}); err != nil {
		t.Fatalf("AddEntity: %v", err)
	}
	if err := g1.AddTriple(Triple{Subject: "PersistentEntity", Predicate: "is", Object: "persistent"}); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}
	db1.Close()

	// Second connection: verify data
	db2, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db2: %v", err)
	}
	defer db2.Close()
	g2 := NewGraph(db2)
	if err := g2.InitDB(); err != nil {
		t.Fatalf("InitDB2: %v", err)
	}

	got, err := g2.GetEntity("PersistentEntity")
	if err != nil {
		t.Fatalf("GetEntity after reconnect: %v", err)
	}
	if got.Name != "PersistentEntity" {
		t.Errorf("expected PersistentEntity, got %q", got.Name)
	}

	triples, err := g2.Query("", "", "")
	if err != nil {
		t.Fatalf("Query after reconnect: %v", err)
	}
	if len(triples) != 1 {
		t.Errorf("expected 1 triple after reconnect, got %d", len(triples))
	}
}

// ─────────────────────────────────────────────────────────────
// 36. Triple with empty validity (always valid)
// ─────────────────────────────────────────────────────────────

func TestTripleAlwaysValid(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// Triple with no validity constraints
	if err := g.AddTriple(Triple{
		Subject:   "Math",
		Predicate: "is",
		Object:    "universal",
	}); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}

	// Should be valid at any date
	for _, date := range []string{"2000-01-01", "2026-06-15", "2099-12-31"} {
		results, err := g.QueryValid(date)
		if err != nil {
			t.Fatalf("QueryValid %s: %v", date, err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 valid triple at %s, got %d", date, len(results))
		}
	}
}

// ─────────────────────────────────────────────────────────────
// 37. BFS with disconnected start/end
// ─────────────────────────────────────────────────────────────

func TestGetPathNonexistentEntities(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// Add some unrelated triples
	if err := g.AddTriple(Triple{Subject: "X", Predicate: "rel", Object: "Y"}); err != nil {
		t.Fatalf("AddTriple: %v", err)
	}

	// Try path between entities not in the graph
	paths, err := g.GetPath("A", "Z", 5)
	if err != nil {
		t.Fatalf("GetPath: %v", err)
	}
	if paths != nil {
		t.Errorf("expected nil paths for nonexistent entities, got %v", paths)
	}
}

// ─────────────────────────────────────────────────────────────
// 38. Multiple aliases for same entity
// ─────────────────────────────────────────────────────────────

func TestMultipleAliases(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	if err := g.AddEntity(Entity{Name: "Go", Type: "project"}); err != nil {
		t.Fatalf("AddEntity: %v", err)
	}

	aliases := []string{"Golang", "Go Language", "Gopher Lang"}
	for _, a := range aliases {
		if err := g.AddAlias("Go", a); err != nil {
			t.Fatalf("AddAlias %s: %v", a, err)
		}
	}

	got, err := g.GetEntity("Go")
	if err != nil {
		t.Fatalf("GetEntity: %v", err)
	}

	aliasSet := make(map[string]bool)
	for _, a := range got.Aliases {
		aliasSet[a] = true
	}

	for _, expected := range aliases {
		if !aliasSet[expected] {
			t.Errorf("missing alias %q, got %v", expected, got.Aliases)
		}
	}
}

// ─────────────────────────────────────────────────────────────
// 39. QueryValid with overlapping validity windows
// ─────────────────────────────────────────────────────────────

func TestQueryValidOverlappingWindows(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	// Two overlapping triples
	if err := g.AddTriple(Triple{
		Subject:   "Alice",
		Predicate: "role",
		Object:    "engineer",
		ValidFrom: "2026-01-01",
		ValidTo:   "2026-06-30",
	}); err != nil {
		t.Fatalf("AddTriple 1: %v", err)
	}

	if err := g.AddTriple(Triple{
		Subject:   "Alice",
		Predicate: "role",
		Object:    "manager",
		ValidFrom: "2026-04-01",
		ValidTo:   "2026-12-31",
	}); err != nil {
		t.Fatalf("AddTriple 2: %v", err)
	}

	// During overlap (May) — both valid
	may, err := g.QueryValid("2026-05-15")
	if err != nil {
		t.Fatalf("QueryValid May: %v", err)
	}
	if len(may) != 2 {
		t.Errorf("expected 2 valid triples in May overlap, got %d", len(may))
	}

	// Before overlap (Feb) — only engineer
	feb, err := g.QueryValid("2026-02-15")
	if err != nil {
		t.Fatalf("QueryValid Feb: %v", err)
	}
	if len(feb) != 1 {
		t.Errorf("expected 1 valid triple in Feb, got %d", len(feb))
	}
	if feb[0].Object != "engineer" {
		t.Errorf("expected engineer in Feb, got %q", feb[0].Object)
	}

	// After overlap (Oct) — only manager
	oct, err := g.QueryValid("2026-10-15")
	if err != nil {
		t.Fatalf("QueryValid Oct: %v", err)
	}
	if len(oct) != 1 {
		t.Errorf("expected 1 valid triple in Oct, got %d", len(oct))
	}
	if oct[0].Object != "manager" {
		t.Errorf("expected manager in Oct, got %q", oct[0].Object)
	}
}

// ─────────────────────────────────────────────────────────────
// 40. Stats on empty graph
// ─────────────────────────────────────────────────────────────

func TestStatsEmpty(t *testing.T) {
	g, cleanup := setupGraph(t)
	defer cleanup()

	stats, err := g.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats["total_entities"] != 0 {
		t.Errorf("expected 0 total entities, got %d", stats["total_entities"])
	}
	if stats["total_triples"] != 0 {
		t.Errorf("expected 0 total triples, got %d", stats["total_triples"])
	}
}
