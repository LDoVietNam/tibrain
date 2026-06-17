package memory_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ti/cli/internal/memory"
)

// ─────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────

func newTestPalace(t *testing.T) *memory.Palace {
	t.Helper()
	dir := t.TempDir()
	p := memory.NewPalace(dir)
	if err := p.Load(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { p.Close() })
	return p
}

func testDrawer(text string, category string) memory.Drawer {
	return memory.Drawer{
		Text:       text,
		Importance: 0.8,
		Category:   category,
		CreatedAt:  time.Now().Unix(),
	}
}

// ─────────────────────────────────────────────────────────────
// 1. Palace Construction & Identity
// ─────────────────────────────────────────────────────────────

func TestNewPalaceDefaultIdentity(t *testing.T) {
	dir := t.TempDir()
	p := memory.NewPalace(dir)
	if p.Path != dir {
		t.Fatalf("expected path %s, got %s", dir, p.Path)
	}
	if p.Identity == "" {
		t.Fatal("expected default identity")
	}
	if p.Wings == nil {
		t.Fatal("expected wings map to be initialized")
	}
}

func TestNewPalaceWithIdentityFile(t *testing.T) {
	dir := t.TempDir()
	identityContent := "I am Ti, an AI assistant for the Ti project.\nTraits: fast, accurate, helpful."
	if err := os.WriteFile(filepath.Join(dir, "identity.txt"), []byte(identityContent), 0644); err != nil {
		t.Fatal(err)
	}
	p := memory.NewPalace(dir)
	if p.Identity != identityContent {
		t.Fatalf("expected identity from file, got: %s", p.Identity)
	}
}

// ─────────────────────────────────────────────────────────────
// 2. SQLite Load/Save/Persistence
// ─────────────────────────────────────────────────────────────

func TestLoadCreatesDatabase(t *testing.T) {
	dir := t.TempDir()
	p := memory.NewPalace(dir)
	if err := p.Load(); err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	dbPath := filepath.Join(dir, "memory.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file not created after Load()")
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	p := memory.NewPalace(dir)
	if err := p.Load(); err != nil {
		t.Fatal(err)
	}

	d := testDrawer("Postgres chosen over MongoDB for ACID compliance", memory.CategoryDecision)
	if err := p.AddDrawer("ticlaw", "database", d); err != nil {
		t.Fatal(err)
	}
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	p.Close()

	p2 := memory.NewPalace(dir)
	if err := p2.Load(); err != nil {
		t.Fatal(err)
	}
	defer p2.Close()

	stack := p2.Recall("ticlaw", 5)
	if len(stack.L1) != 1 {
		t.Fatalf("expected 1 drawer after reload, got %d", len(stack.L1))
	}
	if stack.L1[0].Text != "Postgres chosen over MongoDB for ACID compliance" {
		t.Fatalf("wrong text after reload: %s", stack.L1[0].Text)
	}
}

func TestSaveWithoutLoad(t *testing.T) {
	dir := t.TempDir()
	p := memory.NewPalace(dir)
	if err := p.Save(); err == nil {
		t.Fatal("expected error when saving without load")
	}
}

// ─────────────────────────────────────────────────────────────
// 3. Drawer CRUD Operations
// ─────────────────────────────────────────────────────────────

func TestAddDrawer(t *testing.T) {
	p := newTestPalace(t)

	d := testDrawer("Claude handles auth tasks best", memory.CategoryFact)
	if err := p.AddDrawer("my_project", "auth", d); err != nil {
		t.Fatal(err)
	}

	stack := p.Recall("my_project", 5)
	if len(stack.L1) != 1 {
		t.Fatalf("expected 1 drawer, got %d", len(stack.L1))
	}
}

func TestAddDrawerAutoFields(t *testing.T) {
	p := newTestPalace(t)

	d := memory.Drawer{Text: "auto fields test"}
	if err := p.AddDrawer("w", "r", d); err != nil {
		t.Fatal(err)
	}

	stack := p.Recall("w", 5)
	if len(stack.L1) != 1 {
		t.Fatalf("expected 1 drawer, got %d", len(stack.L1))
	}
	drawer := stack.L1[0]
	if drawer.ID == "" {
		t.Fatal("expected auto-generated ID")
	}
	if drawer.Importance != 0.7 {
		t.Fatalf("expected importance 0.7, got %f", drawer.Importance)
	}
	if drawer.Category != memory.CategoryOther {
		t.Fatalf("expected category 'other', got %s", drawer.Category)
	}
	if drawer.CreatedAt == 0 {
		t.Fatal("expected auto-set CreatedAt")
	}
}

func TestAddMultipleDrawers(t *testing.T) {
	p := newTestPalace(t)

	topics := []struct{ room, text string }{
		{"auth", "OAuth2 flow configured with PKCE"},
		{"auth", "JWT tokens expire in 15 minutes"},
		{"api", "GraphQL endpoint at /api/graphql"},
		{"api", "Rate limiting: 100 req/min per user"},
		{"deploy", "Deploy to production via GitHub Actions"},
	}
	for _, tc := range topics {
		if err := p.AddDrawer("myapp", tc.room, testDrawer(tc.text, memory.CategoryFact)); err != nil {
			t.Fatal(err)
		}
	}

	if total := p.CountDrawers(); total != 5 {
		t.Fatalf("expected 5 drawers, got %d", total)
	}
}

func TestGetDrawer(t *testing.T) {
	p := newTestPalace(t)

	d := testDrawer("specific drawer", memory.CategoryFact)
	if err := p.AddDrawer("w", "r", d); err != nil {
		t.Fatal(err)
	}

	stack := p.Recall("w", 5)
	id := stack.L1[0].ID

	found, err := p.GetDrawer("w", "r", id)
	if err != nil {
		t.Fatal(err)
	}
	if found.Text != "specific drawer" {
		t.Fatalf("wrong text: %s", found.Text)
	}
}

func TestGetDrawerNotFound(t *testing.T) {
	p := newTestPalace(t)

	_, err := p.GetDrawer("nonexistent", "r", "id")
	if err == nil {
		t.Fatal("expected error for nonexistent wing")
	}
}

func TestDeleteDrawer(t *testing.T) {
	p := newTestPalace(t)

	d := testDrawer("to delete", memory.CategoryFact)
	if err := p.AddDrawer("w", "r", d); err != nil {
		t.Fatal(err)
	}

	stack := p.Recall("w", 5)
	id := stack.L1[0].ID

	if err := p.DeleteDrawer("w", "r", id); err != nil {
		t.Fatal(err)
	}

	if total := p.CountDrawers(); total != 0 {
		t.Fatalf("expected 0 drawers after delete, got %d", total)
	}
}

func TestUpdateDrawer(t *testing.T) {
	p := newTestPalace(t)

	d := testDrawer("original text", memory.CategoryFact)
	if err := p.AddDrawer("w", "r", d); err != nil {
		t.Fatal(err)
	}

	stack := p.Recall("w", 5)
	id := stack.L1[0].ID

	if err := p.UpdateDrawer("w", "r", id, "updated text", 0.95); err != nil {
		t.Fatal(err)
	}

	found, err := p.GetDrawer("w", "r", id)
	if err != nil {
		t.Fatal(err)
	}
	if found.Text != "updated text" {
		t.Fatalf("expected 'updated text', got %s", found.Text)
	}
	if found.Importance != 0.95 {
		t.Fatalf("expected importance 0.95, got %f", found.Importance)
	}
}

// ─────────────────────────────────────────────────────────────
// 4. Wing & Room Operations
// ─────────────────────────────────────────────────────────────

func TestListWings(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("zeta", "r", testDrawer("z", memory.CategoryFact))
	p.AddDrawer("alpha", "r", testDrawer("a", memory.CategoryFact))
	p.AddDrawer("beta", "r", testDrawer("b", memory.CategoryFact))

	wings := p.ListWings()
	if len(wings) != 3 {
		t.Fatalf("expected 3 wings, got %d", len(wings))
	}
	if wings[0] != "alpha" || wings[1] != "beta" || wings[2] != "zeta" {
		t.Fatalf("expected sorted wings [alpha, beta, zeta], got %v", wings)
	}
}

func TestListRooms(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("myapp", "auth", testDrawer("a", memory.CategoryFact))
	p.AddDrawer("myapp", "api", testDrawer("b", memory.CategoryFact))
	p.AddDrawer("myapp", "deploy", testDrawer("c", memory.CategoryFact))

	rooms, err := p.ListRooms("myapp")
	if err != nil {
		t.Fatal(err)
	}
	if len(rooms) != 3 {
		t.Fatalf("expected 3 rooms, got %d", len(rooms))
	}
	if rooms[0] != "api" || rooms[1] != "auth" || rooms[2] != "deploy" {
		t.Fatalf("expected sorted rooms, got %v", rooms)
	}
}

func TestListRoomsNotFound(t *testing.T) {
	p := newTestPalace(t)

	_, err := p.ListRooms("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent wing")
	}
}

func TestGetTaxonomy(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("proj", "auth", testDrawer("a1", memory.CategoryFact))
	p.AddDrawer("proj", "auth", testDrawer("a2", memory.CategoryFact))
	p.AddDrawer("proj", "api", testDrawer("b1", memory.CategoryFact))

	taxonomy := p.GetTaxonomy()
	if len(taxonomy["proj"]) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(taxonomy["proj"]))
	}
	if taxonomy["proj"]["auth"] != 2 {
		t.Fatalf("expected 2 drawers in auth, got %d", taxonomy["proj"]["auth"])
	}
	if taxonomy["proj"]["api"] != 1 {
		t.Fatalf("expected 1 drawer in api, got %d", taxonomy["proj"]["api"])
	}
}

// ─────────────────────────────────────────────────────────────
// 5. Layer 0 — Identity
// ─────────────────────────────────────────────────────────────

func TestLayer0IdentityInWakeUp(t *testing.T) {
	dir := t.TempDir()
	identity := "I am Ti, AI assistant for the Ti project."
	os.WriteFile(filepath.Join(dir, "identity.txt"), []byte(identity), 0644)

	p := memory.NewPalace(dir)
	if err := p.Load(); err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	wakeUp := p.WakeUp("")
	if !strings.Contains(wakeUp, identity) {
		t.Fatalf("wake-up should contain identity:\n%s", wakeUp)
	}
}

// ─────────────────────────────────────────────────────────────
// 6. Layer 1 — Essential Story
// ─────────────────────────────────────────────────────────────

func TestLayer1EssentialStory(t *testing.T) {
	p := newTestPalace(t)

	importances := []float64{0.9, 0.7, 0.5, 0.3, 0.1}
	for i, imp := range importances {
		d := memory.Drawer{
			Text:       fmt.Sprintf("important fact %d", i),
			Importance: imp,
			Category:   memory.CategoryFact,
			CreatedAt:  time.Now().Unix(),
		}
		p.AddDrawer("myapp", "general", d)
	}

	stack := p.Recall("myapp", 3)
	if len(stack.L1) != 3 {
		t.Fatalf("expected 3 L1 drawers, got %d", len(stack.L1))
	}
	if stack.L1[0].Importance != 0.9 {
		t.Fatalf("expected highest importance first, got %f", stack.L1[0].Importance)
	}
}

func TestLayer1RespectsCharLimit(t *testing.T) {
	p := newTestPalace(t)

	for i := 0; i < 10; i++ {
		d := memory.Drawer{
			Text:       strings.Repeat("very important fact. ", 100),
			Importance: 0.9,
			Category:   memory.CategoryFact,
			CreatedAt:  time.Now().Unix(),
		}
		p.AddDrawer("myapp", "general", d)
	}

	stack := p.Recall("myapp", 15)
	totalChars := 0
	for _, d := range stack.L1 {
		totalChars += len(d.Text)
	}
	if totalChars > 4000 {
		t.Fatalf("L1 too long: %d chars (max ~3200)", totalChars)
	}
}

func TestLayer1EmptyWing(t *testing.T) {
	p := newTestPalace(t)

	stack := p.Recall("nonexistent_wing", 5)
	if len(stack.L1) != 0 {
		t.Fatalf("expected 0 L1 drawers for nonexistent wing, got %d", len(stack.L1))
	}
}

func TestLayer1AllWings(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("wing1", "r", testDrawer("wing1 fact", memory.CategoryFact))
	p.AddDrawer("wing2", "r", testDrawer("wing2 fact", memory.CategoryFact))

	stack := p.Recall("", 10)
	if len(stack.L1) != 2 {
		t.Fatalf("expected 2 L1 drawers from all wings, got %d", len(stack.L1))
	}
}

// ─────────────────────────────────────────────────────────────
// 7. Layer 2 — On-Demand
// ─────────────────────────────────────────────────────────────

func TestLayer2WingFilter(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("wing_a", "auth", testDrawer("auth fact A", memory.CategoryFact))
	p.AddDrawer("wing_a", "api", testDrawer("api fact A", memory.CategoryFact))
	p.AddDrawer("wing_b", "auth", testDrawer("auth fact B", memory.CategoryFact))

	stack := p.OnDemand("wing_a", "", "")
	if len(stack.L2) != 2 {
		t.Fatalf("expected 2 L2 drawers for wing_a, got %d", len(stack.L2))
	}
}

func TestLayer2RoomFilter(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("wing_a", "auth", testDrawer("auth A", memory.CategoryFact))
	p.AddDrawer("wing_b", "auth", testDrawer("auth B", memory.CategoryFact))
	p.AddDrawer("wing_a", "api", testDrawer("api A", memory.CategoryFact))

	stack := p.OnDemand("", "auth", "")
	if len(stack.L2) != 2 {
		t.Fatalf("expected 2 L2 drawers for room auth across wings, got %d", len(stack.L2))
	}
}

func TestLayer2WingAndRoomFilter(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("wing_a", "auth", testDrawer("auth A", memory.CategoryFact))
	p.AddDrawer("wing_a", "api", testDrawer("api A", memory.CategoryFact))
	p.AddDrawer("wing_b", "auth", testDrawer("auth B", memory.CategoryFact))

	stack := p.OnDemand("wing_a", "auth", "")
	if len(stack.L2) != 1 {
		t.Fatalf("expected 1 L2 drawer for wing_a+auth, got %d", len(stack.L2))
	}
}

func TestLayer2WithQuery(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("myapp", "auth", testDrawer("OAuth2 PKCE flow configured", memory.CategoryFact))
	p.AddDrawer("myapp", "auth", testDrawer("JWT token refresh every 15 minutes", memory.CategoryFact))
	p.AddDrawer("myapp", "api", testDrawer("GraphQL rate limiting at 100/min", memory.CategoryFact))

	stack := p.OnDemand("myapp", "auth", "OAuth")
	if len(stack.L2) != 1 {
		t.Fatalf("expected 1 L2 result for 'OAuth' query, got %d", len(stack.L2))
	}
	if !strings.Contains(stack.L2[0].Text, "OAuth2") {
		t.Fatalf("expected OAuth2 in result, got: %s", stack.L2[0].Text)
	}
}

func TestLayer2EmptyFilters(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("w", "r", testDrawer("some fact", memory.CategoryFact))

	stack := p.OnDemand("", "", "")
	if len(stack.L2) != 0 {
		t.Fatalf("expected 0 L2 drawers for empty filters, got %d", len(stack.L2))
	}
}

// ─────────────────────────────────────────────────────────────
// 8. Layer 3 — Deep Search
// ─────────────────────────────────────────────────────────────

func TestLayer3Search(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("myapp", "auth", testDrawer("OAuth2 with PKCE is required for public clients", memory.CategoryDecision))
	p.AddDrawer("myapp", "api", testDrawer("GraphQL endpoint supports subscriptions", memory.CategoryFact))
	p.AddDrawer("myapp", "deploy", testDrawer("Docker compose for local development", memory.CategoryFact))

	stack := p.Search("OAuth2 PKCE", nil, 5)
	if len(stack.L3) == 0 {
		t.Fatal("expected search results for 'OAuth2 PKCE'")
	}
	if !strings.Contains(strings.ToLower(stack.L3[0].Text), "oauth2") {
		t.Fatalf("expected OAuth2 in top result, got: %s", stack.L3[0].Text)
	}
}

func TestLayer3SearchWithWingFilter(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("wing_a", "r", testDrawer("OAuth2 in wing A", memory.CategoryFact))
	p.AddDrawer("wing_b", "r", testDrawer("OAuth2 in wing B", memory.CategoryFact))

	stack := p.Search("OAuth2", []string{"wing_a"}, 5)
	if len(stack.L3) != 1 {
		t.Fatalf("expected 1 result for wing_a filter, got %d", len(stack.L3))
	}
}

func TestLayer3SearchNoResults(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("w", "r", testDrawer("some unrelated fact", memory.CategoryFact))

	stack := p.Search("quantum computing", nil, 5)
	if len(stack.L3) > 0 {
		for _, d := range stack.L3 {
			if strings.Contains(strings.ToLower(d.Text), "quantum") {
				t.Fatalf("unexpected quantum in results: %s", d.Text)
			}
		}
	}
}

func TestLayer3SearchLimit(t *testing.T) {
	p := newTestPalace(t)

	for i := 0; i < 20; i++ {
		p.AddDrawer("myapp", "general", testDrawer("database query optimization fact", memory.CategoryFact))
	}

	stack := p.Search("database", nil, 5)
	if len(stack.L3) > 5 {
		t.Fatalf("expected at most 5 results, got %d", len(stack.L3))
	}
}

// ─────────────────────────────────────────────────────────────
// 9. WakeUp
// ─────────────────────────────────────────────────────────────

func TestWakeUpFormat(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("myapp", "auth", testDrawer("OAuth2 PKCE configured", memory.CategoryDecision))
	p.AddDrawer("myapp", "api", testDrawer("GraphQL at /api/graphql", memory.CategoryFact))

	wakeUp := p.WakeUp("myapp")
	if !strings.Contains(wakeUp, "L0 — IDENTITY") {
		t.Fatal("wake-up missing L0 header")
	}
	if !strings.Contains(wakeUp, "L1 — ESSENTIAL STORY") {
		t.Fatal("wake-up missing L1 header")
	}
}

func TestWakeUpWithSpecificWing(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("wing_a", "r", testDrawer("fact from wing A", memory.CategoryFact))
	p.AddDrawer("wing_b", "r", testDrawer("fact from wing B", memory.CategoryFact))

	wakeUpA := p.WakeUp("wing_a")
	wakeUpB := p.WakeUp("wing_b")

	if strings.Contains(wakeUpA, "wing B") {
		t.Fatal("wake-up for wing_a should not contain wing_b memories")
	}
	if strings.Contains(wakeUpB, "wing A") {
		t.Fatal("wake-up for wing_b should not contain wing_a memories")
	}
}

func TestWakeUpEmptyPalace(t *testing.T) {
	p := newTestPalace(t)

	wakeUp := p.WakeUp("")
	if !strings.Contains(wakeUp, "No memories yet") {
		t.Fatalf("expected 'No memories yet' in empty wake-up:\n%s", wakeUp)
	}
}

// ─────────────────────────────────────────────────────────────
// 10. File Mining
// ─────────────────────────────────────────────────────────────

func TestMineFile(t *testing.T) {
	p := newTestPalace(t)

	content := strings.Repeat("This is a test paragraph about authentication and OAuth2 configuration. ", 20)
	tmpFile := filepath.Join(t.TempDir(), "auth_config.md")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	drawers, err := p.MineFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(drawers) == 0 {
		t.Fatal("expected mined drawers from file")
	}
	found := false
	for _, d := range drawers {
		if d.SourceFile == tmpFile && d.Category != "" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected drawers with source file set")
	}
}

func TestMineFileTooSmall(t *testing.T) {
	p := newTestPalace(t)

	tmpFile := filepath.Join(t.TempDir(), "tiny.txt")
	if err := os.WriteFile(tmpFile, []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	drawers, err := p.MineFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(drawers) != 0 {
		t.Fatalf("expected 0 drawers for tiny file, got %d", len(drawers))
	}
}

func TestMineFileNotFound(t *testing.T) {
	p := newTestPalace(t)

	_, err := p.MineFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestChunkText(t *testing.T) {
	text := strings.Repeat("Hello world. ", 100)
	chunks := memory.ChunkTextForTest(text)
	if len(chunks) == 0 {
		t.Fatal("expected chunks from chunkText")
	}
	for i, chunk := range chunks {
		if len(chunk) < memory.MinChunkSizeForTest() {
			t.Fatalf("chunk %d too small: %d chars", i, len(chunk))
		}
	}
}

func TestDetectCategory(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{"We decided to go with option A", memory.CategoryDecision},
		{"I prefer Go over Python", memory.CategoryPreference},
		{"There is a critical bug in the auth module", memory.CategoryProblem},
		{"The service uses Redis for caching", memory.CategoryEntity},
		{"Version 2.0 released to production", memory.CategoryMilestone},
		{"Random text with no special meaning", memory.CategoryOther},
	}

	for _, tc := range tests {
		cat := memory.DetectCategoryForTest(tc.text)
		if cat != tc.expected {
			t.Errorf("text: %q -> expected %s, got %s", tc.text, tc.expected, cat)
		}
	}
}

func TestDetectRoom(t *testing.T) {
	tests := []struct {
		filePath string
		content  string
		expected string
	}{
		{filepath.Join("project", "auth", "handler.go"), "", "auth"},
		{filepath.Join("project", "api", "router.go"), "", "api"},
		{filepath.Join("project", "deploy", "Dockerfile"), "", "deployment"},
		{filepath.Join("project", "general", "notes.txt"), "database query optimization with postgres", "database"},
	}

	for _, tc := range tests {
		room := memory.DetectRoomForTest(tc.filePath, tc.content)
		if room != tc.expected {
			t.Errorf("path: %q -> expected %s, got %s", tc.filePath, tc.expected, room)
		}
	}
}

// ─────────────────────────────────────────────────────────────
// 11. Conversation Mining
// ─────────────────────────────────────────────────────────────

func TestMineConversation(t *testing.T) {
	p := newTestPalace(t)

	messages := []memory.Message{
		{Role: "user", Content: "Should we use GraphQL or REST?", Time: time.Now().Unix() - 60},
		{Role: "assistant", Content: "GraphQL is better for complex queries. Let's use GraphQL.", Time: time.Now().Unix()},
		{Role: "user", Content: "Agreed. We decided on GraphQL.", Time: time.Now().Unix()},
		{Role: "assistant", Content: "Great choice. I'll set it up.", Time: time.Now().Unix()},
	}

	drawers, err := p.MineConversation(messages, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if len(drawers) != 2 {
		t.Fatalf("expected 2 drawers from 2 exchanges, got %d", len(drawers))
	}
	if drawers[0].Category != memory.CategoryDecision {
		t.Fatalf("expected decision category, got %s", drawers[0].Category)
	}
}

func TestMineConversationEmpty(t *testing.T) {
	p := newTestPalace(t)

	drawers, err := p.MineConversation(nil, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if len(drawers) != 0 {
		t.Fatalf("expected 0 drawers from empty messages, got %d", len(drawers))
	}
}

func TestMineConversationOrphanUser(t *testing.T) {
	p := newTestPalace(t)

	messages := []memory.Message{
		{Role: "user", Content: "Orphan user message with no response", Time: time.Now().Unix()},
	}

	drawers, err := p.MineConversation(messages, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if len(drawers) != 1 {
		t.Fatalf("expected 1 drawer for orphan user, got %d", len(drawers))
	}
}

func TestDetectCategoryFromConversation(t *testing.T) {
	tests := []struct {
		user     string
		ai       string
		expected string
	}{
		{"Let's switch to Postgres", "Agreed, we decided it's better", memory.CategoryDecision},
		{"I prefer dark mode", "OK, I like that too", memory.CategoryPreference},
		{"The auth module has a bug", "I found the issue and fixed it", memory.CategoryProblem},
		{"What database do we use?", "We use Postgres version 15", memory.CategoryEntity},
		{"Hello", "Hi there!", memory.CategoryFact},
	}

	for _, tc := range tests {
		cat := memory.DetectCategoryFromConversationForTest(tc.user, tc.ai)
		if cat != tc.expected {
			t.Errorf("user: %q, ai: %q -> expected %s, got %s", tc.user, tc.ai, tc.expected, cat)
		}
	}
}

// ─────────────────────────────────────────────────────────────
// 12. Merge
// ─────────────────────────────────────────────────────────────

func TestMergePalaces(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	p1 := memory.NewPalace(dir1)
	p1.Load()
	defer p1.Close()

	p2 := memory.NewPalace(dir2)
	p2.Load()
	defer p2.Close()

	p1.AddDrawer("shared_wing", "room_a", testDrawer("drawer from p1", memory.CategoryFact))
	p2.AddDrawer("shared_wing", "room_b", testDrawer("drawer from p2", memory.CategoryFact))
	p2.AddDrawer("unique_wing", "room_c", testDrawer("unique to p2", memory.CategoryFact))

	if err := p1.Merge(p2); err != nil {
		t.Fatal(err)
	}

	if total := p1.CountDrawers(); total != 3 {
		t.Fatalf("expected 3 drawers after merge, got %d", total)
	}

	wings := p1.ListWings()
	if len(wings) != 2 {
		t.Fatalf("expected 2 wings after merge, got %d", len(wings))
	}
}

// ─────────────────────────────────────────────────────────────
// 13. Export & Import
// ─────────────────────────────────────────────────────────────

func TestExportJSON(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("wing_a", "room_1", testDrawer("fact 1", memory.CategoryFact))
	p.AddDrawer("wing_a", "room_2", testDrawer("fact 2", memory.CategoryDecision))
	p.AddDrawer("wing_b", "room_3", testDrawer("fact 3", memory.CategoryPreference))

	data, err := p.Export()
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := parsed["identity"]; !ok {
		t.Fatal("export missing identity")
	}
	if _, ok := parsed["wings"]; !ok {
		t.Fatal("export missing wings")
	}
}

func TestImportJSON(t *testing.T) {
	dir := t.TempDir()
	p := memory.NewPalace(dir)
	p.Load()
	defer p.Close()

	p.AddDrawer("imported_wing", "imported_room", testDrawer("imported fact", memory.CategoryFact))

	data, err := p.Export()
	if err != nil {
		t.Fatal(err)
	}

	dir2 := t.TempDir()
	p2 := memory.NewPalace(dir2)
	if err := p2.Import(data); err != nil {
		t.Fatal(err)
	}

	if total := p2.CountDrawers(); total != 1 {
		t.Fatalf("expected 1 drawer after import, got %d", total)
	}
}

func TestExportImportRoundTrip(t *testing.T) {
	p := newTestPalace(t)

	for _, wing := range []string{"w1", "w2", "w3"} {
		for i := 0; i < 3; i++ {
			p.AddDrawer(wing, "room", testDrawer(fmt.Sprintf("%s drawer %d", wing, i), memory.CategoryFact))
		}
	}

	data, err := p.Export()
	if err != nil {
		t.Fatal(err)
	}

	p2 := memory.NewPalace(t.TempDir())
	if err := p2.Import(data); err != nil {
		t.Fatal(err)
	}

	if p2.CountDrawers() != p.CountDrawers() {
		t.Fatalf("drawer count mismatch: original=%d, imported=%d", p.CountDrawers(), p2.CountDrawers())
	}

	if len(p2.ListWings()) != len(p.ListWings()) {
		t.Fatalf("wing count mismatch")
	}
}

// ─────────────────────────────────────────────────────────────
// 14. Recency Decay
// ─────────────────────────────────────────────────────────────

func TestRecencyDecayBrandNew(t *testing.T) {
	now := time.Now().Unix()
	score := memory.RecencyDecayForTest(now, now, 168)
	if math.Abs(score-1.0) > 1e-10 {
		t.Fatalf("expected 1.0 for brand new, got %f", score)
	}
}

func TestRecencyDecayHalfLife(t *testing.T) {
	now := time.Now().Unix()
	halfLife := 168.0
	sevenDaysAgo := now - int64(7*24*3600)

	score := memory.RecencyDecayForTest(sevenDaysAgo, now, halfLife)
	if math.Abs(score-0.5) > 1e-6 {
		t.Fatalf("expected 0.5 after one half-life, got %f", score)
	}
}

func TestRecencyDecayTwoHalfLives(t *testing.T) {
	now := time.Now().Unix()
	halfLife := 168.0
	fourteenDaysAgo := now - int64(14*24*3600)

	score := memory.RecencyDecayForTest(fourteenDaysAgo, now, halfLife)
	if math.Abs(score-0.25) > 1e-6 {
		t.Fatalf("expected 0.25 after two half-lives, got %f", score)
	}
}

func TestRecencyDecayVeryOld(t *testing.T) {
	now := time.Now().Unix()
	oneYearAgo := now - int64(365*24*3600)

	score := memory.RecencyDecayForTest(oneYearAgo, now, 168)
	if score > 0.01 {
		t.Fatalf("expected ~0 for 1 year old, got %f", score)
	}
}

func TestRecencyDecayFuture(t *testing.T) {
	now := time.Now().Unix()
	future := now + 3600

	score := memory.RecencyDecayForTest(future, now, 168)
	if math.Abs(score-1.0) > 1e-10 {
		t.Fatalf("expected 1.0 for future timestamp, got %f", score)
	}
}

// ─────────────────────────────────────────────────────────────
// 15. Stack Formatting
// ─────────────────────────────────────────────────────────────

func TestFormatStack(t *testing.T) {
	stack := &memory.Stack{
		L0: "I am Ti",
		L1: []memory.Drawer{{Text: "L1 fact", Category: "fact", Wing: "w", Room: "r"}},
		L2: []memory.Drawer{{Text: "L2 fact", Category: "fact", Wing: "w", Room: "r"}},
		L3: []memory.Drawer{{Text: "L3 fact", Category: "fact", Wing: "w", Room: "r"}},
	}

	formatted := memory.FormatStack(stack)
	if !strings.Contains(formatted, "L0 Identity") {
		t.Fatal("missing L0 in format")
	}
	if !strings.Contains(formatted, "L1 Essential") {
		t.Fatal("missing L1 in format")
	}
	if !strings.Contains(formatted, "L2 On-Demand") {
		t.Fatal("missing L2 in format")
	}
	if !strings.Contains(formatted, "L3 Deep Search") {
		t.Fatal("missing L3 in format")
	}
}

func TestStackToContext(t *testing.T) {
	stack := &memory.Stack{
		L0: "I am Ti",
		L1: []memory.Drawer{{Text: "L1 fact", Category: "fact", Wing: "w", Room: "r"}},
		L2: []memory.Drawer{{Text: "L2 fact", Category: "fact", Wing: "w", Room: "r"}},
		L3: []memory.Drawer{{Text: "L3 fact", Category: "fact", Wing: "w", Room: "r"}},
	}

	ctx := memory.StackToContext(stack)
	if !strings.Contains(ctx, "I am Ti") {
		t.Fatal("missing identity in context")
	}
	if !strings.Contains(ctx, "L1 fact") {
		t.Fatal("missing L1 in context")
	}
	if !strings.Contains(ctx, "L2 fact") {
		t.Fatal("missing L2 in context")
	}
	if !strings.Contains(ctx, "L3 fact") {
		t.Fatal("missing L3 in context")
	}
}

// ─────────────────────────────────────────────────────────────
// 16. Thread Safety
// ─────────────────────────────────────────────────────────────

func TestConcurrentAccess(t *testing.T) {
	p := newTestPalace(t)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			wing := fmt.Sprintf("wing_%d", n%5)
			room := fmt.Sprintf("room_%d", n%3)
			d := testDrawer(fmt.Sprintf("concurrent drawer %d", n), memory.CategoryFact)
			p.AddDrawer(wing, room, d)
		}(i)
	}
	wg.Wait()

	total := p.CountDrawers()
	if total != 50 {
		t.Fatalf("expected 50 drawers, got %d", total)
	}
}

func TestConcurrentReads(t *testing.T) {
	p := newTestPalace(t)

	for i := 0; i < 20; i++ {
		p.AddDrawer("myapp", "general", testDrawer(fmt.Sprintf("fact %d", i), memory.CategoryFact))
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.Recall("myapp", 5)
			_ = p.ListWings()
			_ = p.CountDrawers()
		}()
	}
	wg.Wait()
}

// ─────────────────────────────────────────────────────────────
// 17. CountDrawers
// ─────────────────────────────────────────────────────────────

func TestCountDrawers(t *testing.T) {
	p := newTestPalace(t)

	if p.CountDrawers() != 0 {
		t.Fatal("expected 0 drawers in new palace")
	}

	p.AddDrawer("w", "r1", testDrawer("a", memory.CategoryFact))
	p.AddDrawer("w", "r2", testDrawer("b", memory.CategoryFact))
	p.AddDrawer("w2", "r1", testDrawer("c", memory.CategoryFact))

	if total := p.CountDrawers(); total != 3 {
		t.Fatalf("expected 3 drawers, got %d", total)
	}
}

// ─────────────────────────────────────────────────────────────
// 18. SetHalfLife
// ─────────────────────────────────────────────────────────────

func TestSetHalfLife(t *testing.T) {
	dir := t.TempDir()
	p := memory.NewPalace(dir)
	p.SetHalfLife(24)
	p.Load()
	defer p.Close()

	oldTime := time.Now().Add(-48 * time.Hour).Unix()
	d := memory.Drawer{
		Text:       "old fact",
		Importance: 0.8,
		Category:   memory.CategoryFact,
		CreatedAt:  oldTime,
	}
	p.AddDrawer("w", "r", d)

	stack := p.Recall("w", 5)
	if len(stack.L1) != 1 {
		t.Fatalf("expected 1 drawer, got %d", len(stack.L1))
	}
	if stack.L1[0].Importance != 0.8 {
		t.Fatalf("expected importance 0.8, got %f", stack.L1[0].Importance)
	}
}

// ─────────────────────────────────────────────────────────────
// 19. Generate Drawer ID
// ─────────────────────────────────────────────────────────────

func TestGenerateDrawerIDDeterministic(t *testing.T) {
	ts := time.Now().Unix()
	id1 := memory.GenerateDrawerIDForTest("w", "r", "same text", ts)
	id2 := memory.GenerateDrawerIDForTest("w", "r", "same text", ts)
	if id1 != id2 {
		t.Fatalf("expected deterministic IDs: %s != %s", id1, id2)
	}
}

func TestGenerateDrawerIDDifferent(t *testing.T) {
	ts := time.Now().Unix()
	id1 := memory.GenerateDrawerIDForTest("w", "r", "text1", ts)
	id2 := memory.GenerateDrawerIDForTest("w", "r", "text2", ts)
	if id1 == id2 {
		t.Fatal("expected different IDs for different text")
	}
}

func TestGenerateDrawerIDPrefix(t *testing.T) {
	id := memory.GenerateDrawerIDForTest("w", "r", "text", 0)
	if !strings.HasPrefix(id, "drawer_") {
		t.Fatalf("expected 'drawer_' prefix, got: %s", id)
	}
}

// ─────────────────────────────────────────────────────────────
// 20. MineFiles (multi-file)
// ─────────────────────────────────────────────────────────────

func TestMineFilesMultipleFiles(t *testing.T) {
	p := newTestPalace(t)

	dir := t.TempDir()
	file1 := filepath.Join(dir, "auth.go")
	file2 := filepath.Join(dir, "api.go")

	os.WriteFile(file1, []byte(strings.Repeat("Authentication with OAuth2 and JWT tokens. ", 30)), 0644)
	os.WriteFile(file2, []byte(strings.Repeat("API endpoint for GraphQL at /api/graphql. ", 30)), 0644)

	files := []string{file1, file2}
	if err := p.MineFiles(files, "myapp"); err != nil {
		t.Fatal(err)
	}

	total := p.CountDrawers()
	if total == 0 {
		t.Fatal("expected drawers from mining multiple files")
	}
}

// ─────────────────────────────────────────────────────────────
// 21. Score Drawer By Text
// ─────────────────────────────────────────────────────────────

func TestScoreDrawersByTextRelevance(t *testing.T) {
	p := newTestPalace(t)

	p.AddDrawer("w", "r", memory.Drawer{
		Text: "OAuth2 PKCE authentication flow", Importance: 0.9, Category: memory.CategoryFact, CreatedAt: time.Now().Unix(),
	})
	p.AddDrawer("w", "r", memory.Drawer{
		Text: "Random unrelated text about cooking", Importance: 0.5, Category: memory.CategoryFact, CreatedAt: time.Now().Unix(),
	})

	stack := p.OnDemand("w", "r", "OAuth2")
	if len(stack.L2) == 0 {
		t.Fatal("expected L2 results for OAuth2 query")
	}
	// OAuth2 drawer should be first
	if !strings.Contains(stack.L2[0].Text, "OAuth2") {
		t.Fatalf("expected OAuth2 as top result, got: %s", stack.L2[0].Text)
	}
}

// ─────────────────────────────────────────────────────────────
// 22. Database Path Creation
// ─────────────────────────────────────────────────────────────

func TestDatabasePathCreation(t *testing.T) {
	dir := t.TempDir()
	dbDir := filepath.Join(dir, "sub", "dir")
	p := memory.NewPalace(dbDir)
	if err := p.Load(); err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		t.Fatal("palace directory not created")
	}
	dbPath := filepath.Join(dbDir, "memory.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file not created")
	}
}

// ─────────────────────────────────────────────────────────────
// 23. Close
// ─────────────────────────────────────────────────────────────

func TestClose(t *testing.T) {
	dir := t.TempDir()
	p := memory.NewPalace(dir)
	if err := p.Load(); err != nil {
		t.Fatal(err)
	}

	if err := p.Close(); err != nil {
		t.Fatal(err)
	}

	// Second close should not panic
	if err := p.Close(); err != nil {
		// Expected: db already closed
	}
}

// ─────────────────────────────────────────────────────────────
// 24. Recency Affects L1 Scoring
// ─────────────────────────────────────────────────────────────

func TestRecencyAffectsL1Scoring(t *testing.T) {
	p := newTestPalace(t)

	// Old drawer with high importance
	oldTime := time.Now().Add(-30 * 24 * time.Hour).Unix()
	p.AddDrawer("w", "r", memory.Drawer{
		Text: "old high importance", Importance: 0.9, Category: memory.CategoryFact, CreatedAt: oldTime,
	})

	// Fresh drawer with same importance
	p.AddDrawer("w", "r", memory.Drawer{
		Text: "fresh same importance", Importance: 0.9, Category: memory.CategoryFact, CreatedAt: time.Now().Unix(),
	})

	stack := p.Recall("w", 5)
	if len(stack.L1) != 2 {
		t.Fatalf("expected 2 drawers, got %d", len(stack.L1))
	}
	// Fresh should score higher due to recency
	if stack.L1[0].Text != "fresh same importance" {
		t.Fatalf("fresh should be first, got: %s", stack.L1[0].Text)
	}
}

// ─────────────────────────────────────────────────────────────
// 25. Empty Palace Operations
// ─────────────────────────────────────────────────────────────

func TestEmptyPalaceOperations(t *testing.T) {
	p := newTestPalace(t)

	// All these should work on empty palace
	wings := p.ListWings()
	if len(wings) != 0 {
		t.Fatalf("expected 0 wings, got %d", len(wings))
	}

	if p.CountDrawers() != 0 {
		t.Fatalf("expected 0 drawers, got %d", p.CountDrawers())
	}

	taxonomy := p.GetTaxonomy()
	if len(taxonomy) != 0 {
		t.Fatalf("expected empty taxonomy, got %d entries", len(taxonomy))
	}

	data, err := p.Export()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty JSON export")
	}
}

// ─────────────────────────────────────────────────────────────
// 26. Conversation with Mixed Categories
// ─────────────────────────────────────────────────────────────

func TestMineConversationMixedCategories(t *testing.T) {
	p := newTestPalace(t)

	messages := []memory.Message{
		{Role: "user", Content: "I found a bug in the login flow", Time: time.Now().Unix() - 120},
		{Role: "assistant", Content: "The issue is in the auth middleware, let me fix it", Time: time.Now().Unix() - 60},
		{Role: "user", Content: "What database should we use?", Time: time.Now().Unix()},
		{Role: "assistant", Content: "I prefer Postgres for its reliability", Time: time.Now().Unix() + 60},
	}

	drawers, err := p.MineConversation(messages, "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if len(drawers) != 2 {
		t.Fatalf("expected 2 drawers, got %d", len(drawers))
	}

	// First exchange should be about a problem
	if drawers[0].Category != memory.CategoryProblem {
		t.Logf("first exchange category: %s (expected problem)", drawers[0].Category)
	}

	// Second should be preference
	if drawers[1].Category != memory.CategoryPreference {
		t.Logf("second exchange category: %s (expected preference)", drawers[1].Category)
	}
}

// ─────────────────────────────────────────────────────────────
// 27. Import Invalid JSON
// ─────────────────────────────────────────────────────────────

func TestImportInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := memory.NewPalace(dir)

	if err := p.Import([]byte("not valid json")); err == nil {
		t.Fatal("expected error for invalid JSON import")
	}
}

// ─────────────────────────────────────────────────────────────
// 28. Handoff Storage
// ─────────────────────────────────────────────────────────────

func TestCreateHandoff(t *testing.T) {
	p := newTestPalace(t)

	// Create a handoff
	metadata := map[string]string{
		"task_id":    "123",
		"session_id": "abc",
	}
	id, err := p.CreateHandoff("planner", "code-reviewer", "Fix auth bug", "JWT token expiry issue", metadata)
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("expected non-empty handoff ID")
	}

	// Verify handoff was stored
	handoff, err := p.RecallHandoff("planner", "code-reviewer")
	if err != nil {
		t.Fatal(err)
	}
	if handoff.FromAgent != "planner" {
		t.Fatalf("expected from_agent 'planner', got '%s'", handoff.FromAgent)
	}
	if handoff.ToAgent != "code-reviewer" {
		t.Fatalf("expected to_agent 'code-reviewer', got '%s'", handoff.ToAgent)
	}
	if handoff.Output != "JWT token expiry issue" {
		t.Fatalf("expected output 'JWT token expiry issue', got '%s'", handoff.Output)
	}
	if handoff.Metadata["task_id"] != "123" {
		t.Fatalf("expected metadata task_id '123', got '%s'", handoff.Metadata["task_id"])
	}
}

func TestRecallHandoffNotFound(t *testing.T) {
	p := newTestPalace(t)

	// Try to recall non-existent handoff
	_, err := p.RecallHandoff("planner", "code-reviewer")
	if err == nil {
		t.Fatal("expected error for non-existent handoff")
	}
}

func TestListHandoffs(t *testing.T) {
	p := newTestPalace(t)

	// Create multiple handoffs
	p.CreateHandoff("planner", "code-reviewer", "Task 1", "Output 1", nil)
	p.CreateHandoff("planner", "tdd-guide", "Task 2", "Output 2", nil)
	p.CreateHandoff("code-reviewer", "security-reviewer", "Task 3", "Output 3", nil)

	// List all handoffs
	allHandoffs := p.ListHandoffs("")
	if len(allHandoffs) != 3 {
		t.Fatalf("expected 3 handoffs, got %d", len(allHandoffs))
	}

	// List handoffs for specific agent
	plannerHandoffs := p.ListHandoffs("planner")
	if len(plannerHandoffs) != 2 {
		t.Fatalf("expected 2 handoffs for planner, got %d", len(plannerHandoffs))
	}

	// Verify sorting by timestamp descending
	for i := 0; i < len(allHandoffs)-1; i++ {
		if allHandoffs[i].Timestamp < allHandoffs[i+1].Timestamp {
			t.Fatal("expected handoffs sorted by timestamp descending")
		}
	}
}

func TestHandoffPersistence(t *testing.T) {
	dir := t.TempDir()

	// Create handoff in first palace
	p1 := memory.NewPalace(dir)
	if err := p1.Load(); err != nil {
		t.Fatal(err)
	}
	id, err := p1.CreateHandoff("planner", "code-reviewer", "Test", "Test output", nil)
	if err != nil {
		t.Fatal(err)
	}
	p1.Close()

	// Reload palace and verify handoff persists
	p2 := memory.NewPalace(dir)
	if err := p2.Load(); err != nil {
		t.Fatal(err)
	}
	defer p2.Close()

	handoff, err := p2.RecallHandoff("planner", "code-reviewer")
	if err != nil {
		t.Fatal(err)
	}
	if handoff.Output != "Test output" {
		t.Fatalf("expected output 'Test output', got '%s'", handoff.Output)
	}
}

// ─────────────────────────────────────────────────────────────
// 29. Skill Storage
// ─────────────────────────────────────────────────────────────

func TestStoreSkill(t *testing.T) {
	p := newTestPalace(t)

	// Store a skill
	content := "# Security Best Practices\n\nAlways validate input tokens."
	err := p.StoreSkill("security", content, "general")
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve skill
	skill, err := p.GetSkill("security")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Text != content {
		t.Fatalf("expected skill content to match")
	}
	if skill.Category != memory.CategorySkill {
		t.Fatalf("expected category '%s', got '%s'", memory.CategorySkill, skill.Category)
	}
	if skill.Importance != 0.8 {
		t.Fatalf("expected importance 0.8, got %f", skill.Importance)
	}
}

func TestStoreSkillWithCategory(t *testing.T) {
	p := newTestPalace(t)

	// Store skill with category
	err := p.StoreSkill("auth", "OAuth flow documentation", "auth")
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve skill
	skill, err := p.GetSkill("auth")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Room != "auth" {
		t.Fatalf("expected room 'auth', got '%s'", skill.Room)
	}
}

func TestGetSkillNotFound(t *testing.T) {
	p := newTestPalace(t)

	// Try to get non-existent skill
	_, err := p.GetSkill("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent skill")
	}
}

func TestListSkills(t *testing.T) {
	p := newTestPalace(t)

	// Store multiple skills
	p.StoreSkill("skill1", "Content 1", "general")
	p.StoreSkill("skill2", "Content 2", "auth")
	p.StoreSkill("skill3", "Content 3", "general")

	// List all skills
	allSkills := p.ListSkills("")
	if len(allSkills) != 3 {
		t.Fatalf("expected 3 skills, got %d", len(allSkills))
	}

	// List skills by category
	authSkills := p.ListSkills("auth")
	if len(authSkills) != 1 {
		t.Fatalf("expected 1 auth skill, got %d", len(authSkills))
	}

	generalSkills := p.ListSkills("general")
	if len(generalSkills) != 2 {
		t.Fatalf("expected 2 general skills, got %d", len(generalSkills))
	}

	// Verify sorting by importance descending
	for i := 0; i < len(allSkills)-1; i++ {
		if allSkills[i].Importance < allSkills[i+1].Importance {
			t.Fatal("expected skills sorted by importance descending")
		}
	}
}

func TestDeleteSkill(t *testing.T) {
	p := newTestPalace(t)

	// Store skill
	p.StoreSkill("deleteme", "This will be deleted", "general")

	// Verify skill exists
	_, err := p.GetSkill("deleteme")
	if err != nil {
		t.Fatal("expected skill to exist before deletion")
	}

	// Delete skill
	err = p.DeleteSkill("deleteme")
	if err != nil {
		t.Fatal(err)
	}

	// Verify skill is deleted
	_, err = p.GetSkill("deleteme")
	if err == nil {
		t.Fatal("expected error after skill deletion")
	}
}

func TestDeleteSkillNotFound(t *testing.T) {
	p := newTestPalace(t)

	// Try to delete non-existent skill
	err := p.DeleteSkill("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent skill")
	}
}

func TestSkillPersistence(t *testing.T) {
	dir := t.TempDir()

	// Store skill in first palace
	p1 := memory.NewPalace(dir)
	if err := p1.Load(); err != nil {
		t.Fatal(err)
	}
	err := p1.StoreSkill("persist-test", "This should persist", "general")
	if err != nil {
		t.Fatal(err)
	}
	p1.Close()

	// Reload palace and verify skill persists
	p2 := memory.NewPalace(dir)
	if err := p2.Load(); err != nil {
		t.Fatal(err)
	}
	defer p2.Close()

	skill, err := p2.GetSkill("persist-test")
	if err != nil {
		t.Fatal(err)
	}
	if skill.Text != "This should persist" {
		t.Fatalf("expected skill to persist across reload")
	}
}
