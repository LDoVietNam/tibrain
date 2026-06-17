package management

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ti/cli/internal/config"
)

func TestServerBootstrapEndpoint(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/management/bootstrap", nil)
	NewHandler(NewService(config.Default())).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	var raw map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["dashboard"]; !ok {
		t.Fatalf("missing dashboard")
	}
}
