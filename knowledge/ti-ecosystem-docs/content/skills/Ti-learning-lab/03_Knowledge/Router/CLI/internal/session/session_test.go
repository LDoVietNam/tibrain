package session_test

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ti/cli/internal/session"
)

func newTestStore(t *testing.T) *session.Store {
	t.Helper()
	dir := t.TempDir()
	s, err := session.NewStore(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreateAndGet(t *testing.T) {
	s := newTestStore(t)
	sess := &session.Session{
		Provider: "anthropic", Model: "claude-sonnet-4",
		Phase: "implement", Dir: "/path",
		Messages: []session.Message{{Role: "user", Content: "hello"}},
	}
	if err := s.Create(sess); err != nil {
		t.Fatal(err)
	}
	if sess.ID == "" {
		t.Fatal("expected auto-generated ID")
	}

	got, err := s.Get(sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected session")
	}
	if got.Provider != "anthropic" || got.Phase != "implement" || len(got.Messages) != 1 {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestGetNotFound(t *testing.T) {
	s := newTestStore(t)
	got, err := s.Get("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil")
	}
}

func TestList(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 5; i++ {
		s.Create(&session.Session{Provider: "anthropic", Model: "test",
			Messages: []session.Message{{Role: "user", Content: "msg"}}})
		time.Sleep(10 * time.Millisecond)
	}
	list, err := s.List(10)
	if err != nil || len(list) != 5 {
		t.Fatalf("list = %d, want 5", len(list))
	}
	for i := 0; i < len(list)-1; i++ {
		if list[i].UpdatedAt < list[i+1].UpdatedAt {
			t.Fatal("not ordered by updated_at DESC")
		}
	}
}

func TestListLimit(t *testing.T) {
	s := newTestStore(t)
	for i := 0; i < 10; i++ {
		s.Create(&session.Session{Model: "test"})
	}
	list, err := s.List(3)
	if err != nil || len(list) != 3 {
		t.Fatalf("limit = %d, want 3", len(list))
	}
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)
	sess := &session.Session{Model: "test"}
	s.Create(sess)
	s.Delete(sess.ID)
	got, _ := s.Get(sess.ID)
	if got != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestDeleteNotFound(t *testing.T) {
	s := newTestStore(t)
	if err := s.Delete("nonexistent"); err != nil {
		t.Fatal("should not error")
	}
}

func TestExport(t *testing.T) {
	s := newTestStore(t)
	sess := &session.Session{Provider: "anthropic", Model: "claude-sonnet-4",
		Messages: []session.Message{{Role: "user", Content: "hello"}}}
	s.Create(sess)
	data, err := s.Export(sess.ID)
	if err != nil || len(data) == 0 {
		t.Fatal("expected non-empty JSON")
	}
}

func TestExportNotFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.Export("nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAppendMessage(t *testing.T) {
	s := newTestStore(t)
	sess := &session.Session{Model: "test"}
	s.Create(sess)
	s.AppendMessage(sess.ID, session.Message{Role: "user", Content: "msg1"}, 100)
	s.AppendMessage(sess.ID, session.Message{Role: "assistant", Content: "reply"}, 200)
	got, _ := s.Get(sess.ID)
	if len(got.Messages) != 2 || got.TokenCount != 300 {
		t.Fatalf("messages=%d tokens=%d, want 2/300", len(got.Messages), got.TokenCount)
	}
}

func TestAppendMessageNotFound(t *testing.T) {
	s := newTestStore(t)
	err := s.AppendMessage("nonexistent", session.Message{Role: "user", Content: "test"}, 100)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCount(t *testing.T) {
	s := newTestStore(t)
	n, _ := s.Count()
	if n != 0 {
		t.Fatalf("initial count = %d", n)
	}
	for i := 0; i < 3; i++ {
		s.Create(&session.Session{Model: "test"})
	}
	n, _ = s.Count()
	if n != 3 {
		t.Fatalf("count = %d", n)
	}
}

func TestCleanup(t *testing.T) {
	s := newTestStore(t)

	// Create old session, then backdate it
	old := &session.Session{Model: "old"}
	s.Create(old)

	// Manually backdate the old session in DB
	s.DB().Exec(`UPDATE sessions SET created_at=?, updated_at=? WHERE id=?`,
		time.Now().Add(-10*24*time.Hour).Unix(),
		time.Now().Add(-10*24*time.Hour).Unix(),
		old.ID)

	s.Create(&session.Session{Model: "recent"})

	s.Cleanup(7 * 24 * time.Hour)
	list, _ := s.List(10)
	if len(list) != 1 || list[0].Model != "recent" {
		t.Fatalf("after cleanup: %+v", list)
	}
}

func TestDefaultPath(t *testing.T) {
	s, err := session.NewStore("")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.Count(); err != nil {
		t.Fatal(err)
	}
}

func TestAutoGenerateID(t *testing.T) {
	s := newTestStore(t)
	sess := &session.Session{Model: "test"}
	s.Create(sess)
	if sess.ID == "" || len(sess.ID) < 10 {
		t.Fatalf("ID = %q", sess.ID)
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s1, _ := session.NewStore(dbPath)
	s1.Create(&session.Session{ID: "persist-test", Model: "test-model"})
	s1.Close()
	s2, _ := session.NewStore(dbPath)
	defer s2.Close()
	got, _ := s2.Get("persist-test")
	if got == nil || got.Model != "test-model" {
		t.Fatalf("persistence failed: %+v", got)
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := newTestStore(t)
	var wg sync.WaitGroup
	errCh := make(chan error, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id := "concurrent-" + string(rune('a'+n%26))
			sess, _ := s.Get(id)
			if sess == nil {
				errCh <- s.Create(&session.Session{ID: id, Model: "test"})
			} else {
				errCh <- nil
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	n, _ := s.Count()
	if n != 26 {
		t.Fatalf("count = %d, want 26", n)
	}
}

func TestMultipleMessages(t *testing.T) {
	s := newTestStore(t)
	sess := &session.Session{Model: "test"}
	s.Create(sess)
	for _, c := range []struct {
		Role, Msg string
		Tokens    int
	}{
		{"user", "hello", 10}, {"assistant", "hi", 20},
		{"user", "how?", 15}, {"assistant", "good!", 25},
	} {
		s.AppendMessage(sess.ID, session.Message{Role: c.Role, Content: c.Msg}, c.Tokens)
	}
	got, _ := s.Get(sess.ID)
	if len(got.Messages) != 4 || got.TokenCount != 70 {
		t.Fatalf("messages=%d tokens=%d", len(got.Messages), got.TokenCount)
	}
}

func TestEmptySession(t *testing.T) {
	s := newTestStore(t)
	sess := &session.Session{Model: "empty-test"}
	s.Create(sess)
	got, _ := s.Get(sess.ID)
	if got == nil || len(got.Messages) != 0 {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestUpdatedAtUpdates(t *testing.T) {
	s := newTestStore(t)
	sess := &session.Session{Model: "test"}
	s.Create(sess)
	createdAt := sess.CreatedAt

	time.Sleep(2 * time.Second) // ensure different Unix timestamp

	s.AppendMessage(sess.ID, session.Message{Role: "user", Content: "test"}, 10)
	got, _ := s.Get(sess.ID)
	if got == nil || got.UpdatedAt <= createdAt {
		t.Fatalf("updated_at not updated: created=%d updated=%d", createdAt, got.UpdatedAt)
	}
}

func TestSessionWithPhase(t *testing.T) {
	s := newTestStore(t)
	for _, phase := range []string{"scan", "plan", "spec", "implement", "review", "summarize"} {
		s.Create(&session.Session{Model: "test", Phase: phase})
	}
	list, _ := s.List(10)
	if len(list) != 6 {
		t.Fatalf("count = %d", len(list))
	}
	phaseSet := make(map[string]bool)
	for _, sess := range list {
		phaseSet[sess.Phase] = true
	}
	if len(phaseSet) != 6 {
		t.Fatalf("unique phases = %d", len(phaseSet))
	}
}
