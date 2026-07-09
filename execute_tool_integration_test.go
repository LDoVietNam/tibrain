package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExecuteToolIntegrationHonorsAllowedRoots(t *testing.T) {
	hub, err := NewHub(t.TempDir())
	if err != nil {
		t.Fatalf("NewHub() error = %v", err)
	}
	defer hub.Close()

	allowedRoot := t.TempDir()
	denyRoot := t.TempDir()
	t.Cleanup(func() {
		setExecutionPlaneAllowedRoots(nil)
	})
	setExecutionPlaneAllowedRoots([]string{allowedRoot})

	server := NewServer(hub, &Config{AllowedRoots: []string{allowedRoot}})

	dirPath := filepath.Join(allowedRoot, "smoke-roundtrip")
	filePath := filepath.Join(dirPath, "roundtrip.txt")

	createDir := invokeExecuteTool(t, server, "create_directory", map[string]interface{}{
		"dir_path": dirPath,
	})
	assertExecuteToolSuccess(t, createDir, "create_directory")

	write := invokeExecuteTool(t, server, "write_file", map[string]interface{}{
		"file_path": filePath,
		"content":   "alpha",
	})
	assertExecuteToolSuccess(t, write, "write_file")

	read := invokeExecuteTool(t, server, "read_file", map[string]interface{}{
		"file_path": filePath,
	})
	assertExecuteToolSuccess(t, read, "read_file")
	if got := read["result"].(map[string]interface{})["content"]; got != "alpha" {
		t.Fatalf("read_file content = %v, want alpha", got)
	}

	edit := invokeExecuteTool(t, server, "edit_file", map[string]interface{}{
		"file_path":  filePath,
		"old_string": "alpha",
		"new_string": "beta",
	})
	assertExecuteToolSuccess(t, edit, "edit_file")

	readBack := invokeExecuteTool(t, server, "read_file", map[string]interface{}{
		"file_path": filePath,
	})
	assertExecuteToolSuccess(t, readBack, "read_file after edit")
	if got := readBack["result"].(map[string]interface{})["content"]; got != "beta" {
		t.Fatalf("edited content = %v, want beta", got)
	}

	if runtime.GOOS == "windows" {
		execResp := invokeExecuteTool(t, server, "exec", map[string]interface{}{
			"command":    "echo smoke-ok",
			"wait":       true,
			"timeout_ms": 5000,
		})
		assertExecuteToolSuccess(t, execResp, "exec")
		result := execResp["result"].(map[string]interface{})
		if got := int(result["exitCode"].(float64)); got != 0 {
			t.Fatalf("exec exitCode = %d, want 0", got)
		}
		if got := strings.TrimSpace(result["stdout"].(string)); got != "smoke-ok" {
			t.Fatalf("exec stdout = %q, want smoke-ok", got)
		}
	} else {
		t.Log("skipping exec assertion on non-Windows platform")
	}

	deleteResp := invokeExecuteTool(t, server, "delete_file", map[string]interface{}{
		"file_path": filePath,
	})
	assertExecuteToolSuccess(t, deleteResp, "delete_file")

	existsResp := invokeExecuteTool(t, server, "file_exists", map[string]interface{}{
		"file_path": filePath,
	})
	assertExecuteToolSuccess(t, existsResp, "file_exists")
	if got := existsResp["result"].(map[string]interface{})["exists"]; got != false {
		t.Fatalf("file_exists = %v, want false", got)
	}

	denyResp := invokeExecuteTool(t, server, "write_file", map[string]interface{}{
		"file_path": filepath.Join(denyRoot, "outside.txt"),
		"content":   "blocked",
	})
	if got := denyResp["success"]; got != false {
		t.Fatalf("expected outside-root write to fail, got %#v", denyResp)
	}
	if errMsg, _ := denyResp["error"].(string); !strings.Contains(strings.ToLower(errMsg), "outside allowed roots") {
		t.Fatalf("unexpected deny error: %q", errMsg)
	}
}

func invokeExecuteTool(t *testing.T, server *Server, name string, params map[string]interface{}) map[string]interface{} {
	t.Helper()

	body := map[string]interface{}{
		"name":     name,
		"params":   params,
		"agent_id": "execute-tool-integration-test",
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/execute-tool", strings.NewReader(string(raw)))
	rec := httptest.NewRecorder()

	server.handleExecuteTool(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("execute-tool status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func assertExecuteToolSuccess(t *testing.T, resp map[string]interface{}, label string) {
	t.Helper()
	if got := resp["success"]; got != true {
		t.Fatalf("%s success = %v, want true; resp=%#v", label, got, resp)
	}
	if got := resp["source"]; got != "local" {
		t.Fatalf("%s source = %v, want local", label, got)
	}
	if got := resp["executor_id"]; got != "ti-local-cli" {
		t.Fatalf("%s executor_id = %v, want ti-local-cli", label, got)
	}
}
