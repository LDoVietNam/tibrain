package ticonsole

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestProbeHealthyJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"status":"ok","version":"test"}`))
	}))
	defer server.Close()

	backend := NewHTTPBackend(server.Client(), []ServiceConfig{{
		Name:        "TiBrain",
		Role:        "control-plane",
		BaseURL:     server.URL,
		HealthPaths: []string{"/health"},
	}})

	snapshot := backend.Snapshot(context.Background())
	if len(snapshot.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(snapshot.Services))
	}
	if snapshot.Services[0].State != StateHealthy {
		t.Fatalf("expected healthy, got %s (%s)", snapshot.Services[0].State, snapshot.Services[0].Detail)
	}
	if snapshot.Summary.Healthy != 1 {
		t.Fatalf("expected healthy summary count 1, got %+v", snapshot.Summary)
	}
}

func TestProbeFallsBackAfter404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/health" {
			http.NotFound(writer, request)
			return
		}
		_, _ = writer.Write([]byte(`{"healthy":true}`))
	}))
	defer server.Close()

	backend := NewHTTPBackend(server.Client(), []ServiceConfig{{
		Name:        "Service",
		BaseURL:     server.URL,
		HealthPaths: []string{"/api/health", "/health"},
	}})
	snapshot := backend.Snapshot(context.Background())
	service := snapshot.Services[0]
	if service.State != StateHealthy {
		t.Fatalf("expected healthy fallback, got %s", service.State)
	}
	if !strings.HasSuffix(service.Endpoint, "/health") {
		t.Fatalf("expected fallback endpoint, got %s", service.Endpoint)
	}
}

func TestConfiguredEndpointIsNotAssumedActive(t *testing.T) {
	client := &http.Client{Timeout: 50 * time.Millisecond}
	backend := NewHTTPBackend(client, []ServiceConfig{{
		Name:        "Router@1817",
		Role:        "model-router-candidate",
		BaseURL:     "http://127.0.0.1:1",
		HealthPaths: []string{"/health"},
	}})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	snapshot := backend.Snapshot(ctx)
	if snapshot.Services[0].State == StateHealthy {
		t.Fatal("configured router candidate must not be marked healthy without a successful probe")
	}
}

func TestClassifyDegradedPayload(t *testing.T) {
	state := classifyResponse(http.StatusOK, []byte(`{"status":"degraded"}`))
	if state != StateDegraded {
		t.Fatalf("expected degraded, got %s", state)
	}
}
