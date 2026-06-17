package routeagent

import (
	"path/filepath"
	"testing"
)

func TestDecisionPrefersSharedChatWhenSessionExists(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "route.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	engine := NewEngine(store)
	d, err := engine.Decide("", "chat", "", true, "sharedchat")
	if err != nil {
		t.Fatal(err)
	}
	if d.Primary != RouteSharedChat || d.Fallback != RouteCLIProxyAPI {
		t.Fatalf("unexpected decision: %+v", d)
	}
}

func TestDecisionFallsBackToProxyWithoutSession(t *testing.T) {
	store, err := NewStore(filepath.Join(t.TempDir(), "route.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	engine := NewEngine(store)
	d, err := engine.Decide("", "chat", "", false, "")
	if err != nil {
		t.Fatal(err)
	}
	if d.Primary != RouteCLIProxyAPI {
		t.Fatalf("unexpected decision: %+v", d)
	}
}
