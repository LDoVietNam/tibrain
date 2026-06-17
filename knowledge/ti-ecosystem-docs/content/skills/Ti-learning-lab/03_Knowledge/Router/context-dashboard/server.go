package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var contextDir = "Z:\\10_WORKPLACE\\Ti\\Ti-learning-lab\\03_Knowledge\\Router\\"

type ContextFile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

var contextFiles = []ContextFile{
	{
		ID:          "router",
		Name:        "Ti Router",
		Description: "AI routing gateway with OpenAI-compatible API",
		Path:        contextDir + "router-context.json",
	},
	{
		ID:          "cli",
		Name:        "Ti CLI",
		Description: "Microkernel + Plugin architecture for AI agent ecosystem",
		Path:        contextDir + "cli-context.json",
	},
	{
		ID:          "automation",
		Name:        "Ti Automation",
		Description: "Plugin registry system for automation tasks",
		Path:        contextDir + "automation-context.json",
	},
	{
		ID:          "dashboard",
		Name:        "Ti Dashboard",
		Description: "Web dashboard with embedded templates and static files",
		Path:        contextDir + "dashboard-context.json",
	},
	{
		ID:          "donutbrowser",
		Name:        "Donut Browser",
		Description: "Open Source Anti-Detect Browser (Next.js + Tauri/Rust)",
		Path:        contextDir + "donutbrowser-context.json",
	},
	{
		ID:          "mcp",
		Name:        "MCP Hub",
		Description: "Central hub for all MCP servers and tools",
		Path:        contextDir + "mcp-context.json",
	},
	{
		ID:          "providers",
		Name:        "Providers",
		Description: "API providers and adapters collection",
		Path:        contextDir + "providers-context.json",
	},
	{
		ID:          "tui",
		Name:        "Ti TUI",
		Description: "Terminal UI component for Ti Platform",
		Path:        contextDir + "tui-context.json",
	},
}

func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	// API endpoint to get context files list
	http.HandleFunc("/api/contexts", handleContextsList)

	// API endpoint to get specific context file
	http.HandleFunc("/api/context/", handleContextFile)

	port := 8080
	fmt.Printf("🚀 Ti Context Dashboard running at http://localhost:%d\n", port)
	fmt.Printf("📁 Serving from: %s\n", contextDir)
	fmt.Printf("⏹️  Press Ctrl+C to stop\n")
	
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func handleContextsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	searchTerm := strings.ToLower(r.URL.Query().Get("search"))
	
	var filtered []ContextFile
	for _, cf := range contextFiles {
		if searchTerm == "" ||
			strings.Contains(strings.ToLower(cf.Name), searchTerm) ||
			strings.Contains(strings.ToLower(cf.Description), searchTerm) {
			filtered = append(filtered, cf)
		}
	}
	
	json.NewEncoder(w).Encode(filtered)
}

func handleContextFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/context/")
	if path == "" {
		http.Error(w, "Context ID required", http.StatusBadRequest)
		return
	}
	
	// Find context file
	var cf *ContextFile
	for i := range contextFiles {
		if contextFiles[i].ID == path {
			cf = &contextFiles[i]
			break
		}
	}
	
	if cf == nil {
		http.Error(w, "Context not found", http.StatusNotFound)
		return
	}
	
	// Read JSON file
	data, err := os.ReadFile(cf.Path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Return JSON data
	w.Write(data)
}
