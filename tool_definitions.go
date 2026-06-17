// Tool Definitions - Default tools for TiBrain Tool Registry
package main

import (
	"encoding/json"
	"fmt"
)

// ToolDefinition represents a tool definition for registration
type ToolDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Category    string                 `json:"category"`
	Permissions []string               `json:"permissions"`
}

// DefaultToolDefinitions returns the default tool definitions
func DefaultToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		// File Operations
		{
			ID:          "tool-read-file",
			Name:        "read_file",
			Description: "Read file content from local filesystem",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute path to file",
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "Line offset (optional)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Line limit (optional)",
					},
				},
				"required": []string{"file_path"},
			},
			Category:    "file",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-write-file",
			Name:        "write_file",
			Description: "Write/create file content",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute path to file",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "File content to write",
					},
				},
				"required": []string{"file_path", "content"},
			},
			Category:    "file",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-edit-file",
			Name:        "edit_file",
			Description: "Edit file with search/replace",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute path to file",
					},
					"old_string": map[string]interface{}{
						"type":        "string",
						"description": "String to replace",
					},
					"new_string": map[string]interface{}{
						"type":        "string",
						"description": "Replacement string",
					},
				},
				"required": []string{"file_path", "old_string", "new_string"},
			},
			Category:    "file",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-delete-file",
			Name:        "delete_file",
			Description: "Delete file",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute path to file",
					},
				},
				"required": []string{"file_path"},
			},
			Category:    "file",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "tool-list-files",
			Name:        "list_files",
			Description: "List directory contents",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dir_path": map[string]interface{}{
						"type":        "string",
						"description": "Directory path",
					},
				},
				"required": []string{"dir_path"},
			},
			Category:    "file",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-file-info",
			Name:        "file_info",
			Description: "Get file metadata",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute path to file",
					},
				},
				"required": []string{"file_path"},
			},
			Category:    "file",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-copy-file",
			Name:        "copy_file",
			Description: "Copy file",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"src": map[string]interface{}{
						"type":        "string",
						"description": "Source file path",
					},
					"dst": map[string]interface{}{
						"type":        "string",
						"description": "Destination file path",
					},
				},
				"required": []string{"src", "dst"},
			},
			Category:    "file",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "tool-move-file",
			Name:        "move_file",
			Description: "Move/rename file",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"src": map[string]interface{}{
						"type":        "string",
						"description": "Source file path",
					},
					"dst": map[string]interface{}{
						"type":        "string",
						"description": "Destination file path",
					},
				},
				"required": []string{"src", "dst"},
			},
			Category:    "file",
			Permissions: []string{"devin", "claude"},
		},

		// Network Operations
		{
			ID:          "tool-http-request",
			Name:        "http_request",
			Description: "HTTP GET/POST/PUT/DELETE request",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "Request URL",
					},
					"method": map[string]interface{}{
						"type":        "string",
						"description": "HTTP method (GET, POST, PUT, DELETE)",
					},
					"headers": map[string]interface{}{
						"type":        "object",
						"description": "Request headers",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Request body",
					},
				},
				"required": []string{"url"},
			},
			Category:    "network",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "tool-http-download",
			Name:        "http_download",
			Description: "Download file from URL",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "Download URL",
					},
					"dst": map[string]interface{}{
						"type":        "string",
						"description": "Destination file path",
					},
				},
				"required": []string{"url", "dst"},
			},
			Category:    "network",
			Permissions: []string{"devin", "claude"},
		},

		// System Operations
		{
			ID:          "tool-run-command",
			Name:        "run_command",
			Description: "Execute shell command",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "Command to execute",
					},
				},
				"required": []string{"command"},
			},
			Category:    "system",
			Permissions: []string{"devin"},
		},
		{
			ID:          "tool-process-list",
			Name:        "process_list",
			Description: "List running processes",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"filter": map[string]interface{}{
						"type":        "string",
						"description": "Process name filter (optional)",
					},
				},
			},
			Category:    "system",
			Permissions: []string{"devin"},
		},
		{
			ID:          "tool-process-kill",
			Name:        "process_kill",
			Description: "Kill process by PID",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pid": map[string]interface{}{
						"type":        "integer",
						"description": "Process ID",
					},
				},
				"required": []string{"pid"},
			},
			Category:    "system",
			Permissions: []string{"devin"},
		},

		// Git Operations
		{
			ID:          "tool-git-status",
			Name:        "git_status",
			Description: "Git status",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Repository directory",
					},
				},
				"required": []string{"dir"},
			},
			Category:    "git",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-git-clone",
			Name:        "git_clone",
			Description: "Clone git repository",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "Repository URL",
					},
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Destination directory",
					},
				},
				"required": []string{"url", "dir"},
			},
			Category:    "git",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "tool-git-commit",
			Name:        "git_commit",
			Description: "Create git commit",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Repository directory",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Commit message",
					},
				},
				"required": []string{"dir", "message"},
			},
			Category:    "git",
			Permissions: []string{"devin", "claude"},
		},

		// Search Operations
		{
			ID:          "tool-grep-search",
			Name:        "grep_search",
			Description: "Search text in files using ripgrep",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Search pattern",
					},
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Directory to search",
					},
				},
				"required": []string{"pattern", "dir"},
			},
			Category:    "search",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-find-files",
			Name:        "find_files",
			Description: "Find files by pattern",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "File pattern (e.g., *.go)",
					},
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Directory to search",
					},
				},
				"required": []string{"pattern", "dir"},
			},
			Category:    "search",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Build Operations
		{
			ID:          "tool-go-build",
			Name:        "go_build",
			Description: "Build Go project",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Project directory",
					},
				},
				"required": []string{"dir"},
			},
			Category:    "build",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-go-test",
			Name:        "go_test",
			Description: "Run Go tests",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Project directory",
					},
				},
				"required": []string{"dir"},
			},
			Category:    "build",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-npm-install",
			Name:        "npm_install",
			Description: "Install npm dependencies",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Project directory",
					},
				},
				"required": []string{"dir"},
			},
			Category:    "build",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-npm-test",
			Name:        "npm_test",
			Description: "Run npm tests",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dir": map[string]interface{}{
						"type":        "string",
						"description": "Project directory",
					},
				},
				"required": []string{"dir"},
			},
			Category:    "build",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Docker Operations
		{
			ID:          "tool-docker-build",
			Name:        "docker_build",
			Description: "Build Docker image",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dockerfile": map[string]interface{}{
						"type":        "string",
						"description": "Dockerfile path",
					},
					"tag": map[string]interface{}{
						"type":        "string",
						"description": "Image tag",
					},
				},
				"required": []string{"dockerfile", "tag"},
			},
			Category:    "docker",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "tool-docker-run",
			Name:        "docker_run",
			Description: "Run Docker container",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Image name",
					},
					"args": map[string]interface{}{
						"type":        "array",
						"description": "Additional arguments",
					},
				},
				"required": []string{"image"},
			},
			Category:    "docker",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Database Operations
		{
			ID:          "tool-sqlite-query",
			Name:        "sqlite_query",
			Description: "Execute SQLite query",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"db_path": map[string]interface{}{
						"type":        "string",
						"description": "Database file path",
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL query",
					},
				},
				"required": []string{"db_path", "query"},
			},
			Category:    "database",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Best Source - General Patterns
		{
			ID:          "best-api-design",
			Name:        "api_design",
			Description: "REST API design patterns including resource naming, status codes, pagination, filtering, error responses, versioning, and rate limiting for production APIs",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "API design task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-backend-patterns",
			Name:        "backend_patterns",
			Description: "Backend architecture patterns, API design, database optimization, and server-side best practices for Node.js, Express, and Next.js API routes",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Backend pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-coding-standards",
			Name:        "coding_standards",
			Description: "Universal coding standards, best practices, and patterns for TypeScript, JavaScript, React, and Node.js development",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Coding standards task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-docker-patterns",
			Name:        "docker_patterns",
			Description: "Docker and Docker Compose patterns for local development, container security, networking, volume strategies, and multi-service orchestration",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Docker pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-frontend-patterns",
			Name:        "frontend_patterns",
			Description: "Frontend development patterns for React, Next.js, state management, performance optimization, and UI best practices",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Frontend pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-git-workflow",
			Name:        "git_workflow",
			Description: "Git workflow patterns including branching strategies, commit conventions, merge vs rebase, conflict resolution, and collaborative development best practices",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Git workflow task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-golang-patterns",
			Name:        "golang_patterns",
			Description: "Idiomatic Go patterns, best practices, and conventions for building robust, efficient, and maintainable Go applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Go pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-python-patterns",
			Name:        "python_patterns",
			Description: "Pythonic idioms, PEP 8 standards, type hints, and best practices for building robust, efficient, and maintainable Python applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Python pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-rust-patterns",
			Name:        "rust_patterns",
			Description: "Idiomatic Rust patterns, ownership, error handling, traits, concurrency, and best practices for building safe, performant applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Rust pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-tdd-workflow",
			Name:        "tdd_workflow",
			Description: "Test-driven development workflow with 80%+ coverage including unit, integration, and E2E tests",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "TDD workflow task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-verification-loop",
			Name:        "verification_loop",
			Description: "A comprehensive verification system for Claude Code sessions including build, static analysis, tests with coverage, security scans, and diff review",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Verification loop task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-security-review",
			Name:        "security_review",
			Description: "Use this skill when adding authentication, handling user input, working with secrets, creating API endpoints, or implementing payment/sensitive features",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Security review task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-design-system",
			Name:        "design_system",
			Description: "Use this skill to generate or audit design systems, check visual consistency, and review PRs that touch styling",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Design system task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-search-first",
			Name:        "search_first",
			Description: "Research-before-coding workflow. Search for existing tools, libraries, and patterns before writing custom code",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Search first task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-blueprint",
			Name:        "blueprint",
			Description: "Turn a one-line objective into a step-by-step construction plan for multi-session, multi-agent engineering projects",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Blueprint task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-plan",
			Name:        "plan",
			Description: "Restate requirements, assess risks, and create step-by-step implementation plan",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Plan task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-deployment-patterns",
			Name:        "deployment_patterns",
			Description: "Deployment workflows, CI/CD pipeline patterns, Docker containerization, health checks, rollback strategies, and production readiness checklists for web applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Deployment pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-database-migrations",
			Name:        "database_migrations",
			Description: "Database migration best practices for schema changes, data migrations, rollbacks, and zero-downtime deployments across PostgreSQL, MySQL, and common ORMs",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Database migration task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-postgres-patterns",
			Name:        "postgres_patterns",
			Description: "PostgreSQL database patterns for query optimization, schema design, indexing, and security based on Supabase best practices",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "PostgreSQL pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-jpa-patterns",
			Name:        "jpa_patterns",
			Description: "JPA/Hibernate patterns for entity design, relationships, query optimization, transactions, auditing, indexing, pagination, and pooling in Spring Boot",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "JPA pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-springboot-patterns",
			Name:        "springboot_patterns",
			Description: "Spring Boot architecture patterns, REST API design, layered services, data access, caching, async processing, and logging",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Spring Boot pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-django-patterns",
			Name:        "django_patterns",
			Description: "Django architecture patterns, REST API design with DRF, ORM best practices, caching, signals, middleware, and production-grade Django apps",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Django pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-kotlin-patterns",
			Name:        "kotlin_patterns",
			Description: "Idiomatic Kotlin patterns, best practices, and conventions for building robust, efficient, and maintainable Kotlin applications with coroutines, null safety, and DSL builders",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Kotlin pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-swiftui-patterns",
			Name:        "swiftui_patterns",
			Description: "SwiftUI architecture patterns, state management with @Observable, view composition, navigation, performance optimization, and modern iOS/macOS UI best practices",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "SwiftUI pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-perl-patterns",
			Name:        "perl_patterns",
			Description: "Modern Perl 5.36+ idioms, best practices, and conventions for building robust, maintainable Perl applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Perl pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-e2e-testing",
			Name:        "e2e_testing",
			Description: "Playwright E2E testing patterns, Page Object Model, configuration, CI/CD integration, artifact management, and flaky test strategies",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "E2E testing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-python-testing",
			Name:        "python_testing",
			Description: "Python testing strategies using pytest, TDD methodology, fixtures, mocking, parametrization, and coverage requirements",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Python testing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-golang-testing",
			Name:        "golang_testing",
			Description: "Go testing patterns including table-driven tests, subtests, benchmarks, fuzzing, and test coverage",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Go testing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-rust-testing",
			Name:        "rust_testing",
			Description: "Rust testing patterns including unit tests, integration tests, async testing, property-based testing, mocking, and coverage",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Rust testing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-kotlin-testing",
			Name:        "kotlin_testing",
			Description: "Kotlin testing patterns with Kotest, MockK, coroutine testing, property-based testing, and Kover coverage",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Kotlin testing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "pattern",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Laravel
		{
			ID:          "best-laravel-patterns",
			Name:        "laravel_patterns",
			Description: "Laravel architecture patterns, routing/controllers, Eloquent ORM, service layers, queues, events, caching, and API resources for production apps",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Laravel patterns task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-php",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-laravel-plugin-discovery",
			Name:        "laravel_plugin_discovery",
			Description: "Discover and evaluate Laravel packages via LaraPlugins.io MCP. Use when the user wants to find plugins, check package health, or assess Laravel/PHP compatibility",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Laravel plugin discovery task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-php",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "best-laravel-security",
			Name:        "laravel_security",
			Description: "Laravel security best practices for authn/authz, validation, CSRF, mass assignment, file uploads, secrets, rate limiting, and secure deployment",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Laravel security task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-php",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "best-laravel-tdd",
			Name:        "laravel_tdd",
			Description: "Test-driven development for Laravel with PHPUnit and Pest, factories, database testing, fakes, and coverage targets",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Laravel TDD task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-php",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-laravel-verification",
			Name:        "laravel_verification",
			Description: "Verification loop for Laravel projects: env checks, linting, static analysis, tests with coverage, security scans, and deployment readiness",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Laravel verification task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-php",
			Permissions: []string{"devin", "claude"},
		},

		// Language-Specific Skills - Python
		{
			ID:          "best-django-security",
			Name:        "django_security",
			Description: "Django security best practices for authn/authz, CSRF, middleware, secrets, rate limiting, and secure deployment",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Django security task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-python",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "best-django-tdd",
			Name:        "django_tdd",
			Description: "Test-driven development for Django with pytest, factories, database testing, fakes, and coverage targets",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Django TDD task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-python",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-django-verification",
			Name:        "django_verification",
			Description: "Verification loop for Django projects: env checks, linting, static analysis, tests with coverage, security scans, and deployment readiness",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Django verification task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-python",
			Permissions: []string{"devin", "claude"},
		},

		// Language-Specific Skills - Java
		{
			ID:          "best-java-coding-standards",
			Name:        "java_coding_standards",
			Description: "Java coding standards, best practices, and conventions for building robust, efficient, and maintainable Java applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Java coding standards task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-java",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-springboot-security",
			Name:        "springboot_security",
			Description: "Spring Boot security best practices for authn/authz, validation, CSRF, JWT, OAuth2, secrets, and secure deployment",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Spring Boot security task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-java",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "best-springboot-tdd",
			Name:        "springboot_tdd",
			Description: "Test-driven development for Spring Boot with JUnit, Mockito, WebMvcTest, and integration testing",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Spring Boot TDD task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-java",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-springboot-verification",
			Name:        "springboot_verification",
			Description: "Verification loop for Spring Boot projects: env checks, linting, static analysis, tests with coverage, security scans, and deployment readiness",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Spring Boot verification task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-java",
			Permissions: []string{"devin", "claude"},
		},

		// Language-Specific Skills - Kotlin
		{
			ID:          "best-kotlin-coroutines-flows",
			Name:        "kotlin_coroutines_flows",
			Description: "Kotlin coroutines and flows patterns for asynchronous programming, structured concurrency, and reactive streams",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Kotlin coroutines/flows task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-kotlin",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-kotlin-exposed-patterns",
			Name:        "kotlin_exposed_patterns",
			Description: "Kotlin Exposed ORM patterns for database operations, transactions, migrations, and query optimization",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Kotlin Exposed pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-kotlin",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-kotlin-ktor-patterns",
			Name:        "kotlin_ktor_patterns",
			Description: "Kotlin Ktor framework patterns for building web applications, routing, plugins, and websockets",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Kotlin Ktor pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-kotlin",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Language-Specific Skills - Swift
		{
			ID:          "best-swift-actor-persistence",
			Name:        "swift_actor_persistence",
			Description: "Swift Actor patterns for thread-safe data access and persistence in modern Swift applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Swift Actor persistence task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-swift",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-swift-concurrency-6-2",
			Name:        "swift_concurrency_6_2",
			Description: "Swift 6.2 concurrency patterns including async/await, structured concurrency, actors, and task groups",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Swift concurrency task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-swift",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-swift-protocol-di-testing",
			Name:        "swift_protocol_di_testing",
			Description: "Swift protocol-oriented design patterns and dependency injection testing strategies",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Swift protocol DI testing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-swift",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Language-Specific Skills - Perl
		{
			ID:          "best-perl-security",
			Name:        "perl_security",
			Description: "Perl security best practices for input validation, taint mode, secure file handling, and safe system calls",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Perl security task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-perl",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "best-perl-testing",
			Name:        "perl_testing",
			Description: "Perl testing patterns with Test::More, Test::Deep, Test::Exception, and coverage analysis",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Perl testing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-perl",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Language-Specific Skills - Flutter
		{
			ID:          "best-flutter-dart-code-review",
			Name:        "flutter_dart_code_review",
			Description: "Flutter and Dart code review patterns for widget composition, state management, performance, and testing",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Flutter/Dart code review task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-flutter",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Language-Specific Skills - Android
		{
			ID:          "best-android-clean-architecture",
			Name:        "android_clean_architecture",
			Description: "Android clean architecture patterns with MVVM, repositories, use cases, and dependency injection",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Android clean architecture task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-android",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Language-Specific Skills - Compose
		{
			ID:          "best-compose-multiplatform-patterns",
			Name:        "compose_multiplatform_patterns",
			Description: "Jetpack Compose Multiplatform patterns for cross-platform UI development with shared UI code",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Compose Multiplatform pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-compose",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Language-Specific Skills - C++
		{
			ID:          "best-cpp-coding-standards",
			Name:        "cpp_coding_standards",
			Description: "C++ coding standards, best practices, and conventions for building robust, efficient, and maintainable C++ applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "C++ coding standards task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-cpp",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-cpp-testing",
			Name:        "cpp_testing",
			Description: "C++ testing patterns with Google Test, Google Mock, Catch2, and coverage analysis",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "C++ testing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "language-cpp",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Database Skills
		{
			ID:          "best-clickhouse-io",
			Name:        "clickhouse_io",
			Description: "ClickHouse database patterns for analytics, time-series data, columnar storage, and query optimization",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "ClickHouse I/O task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "database",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// AI/ML Skills
		{
			ID:          "best-agentic-engineering",
			Name:        "agentic_engineering",
			Description: "Agentic engineering patterns for building autonomous AI agents, tool use, planning, and multi-agent systems",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Agentic engineering task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-ai-first-engineering",
			Name:        "ai_first_engineering",
			Description: "AI-first engineering methodology for building products with LLMs as core components",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "AI-first engineering task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-ai-regression-testing",
			Name:        "ai_regression_testing",
			Description: "AI regression testing patterns for LLM applications, evaluation harnesses, and automated quality checks",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "AI regression testing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-claude-api",
			Name:        "claude_api",
			Description: "Claude API integration patterns, tool use, streaming, and best practices for Claude-powered applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Claude API task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-cost-aware-llm-pipeline",
			Name:        "cost_aware_llm_pipeline",
			Description: "Cost-aware LLM pipeline patterns for optimizing token usage, caching, model selection, and budget management",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Cost-aware LLM pipeline task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-foundation-models-on-device",
			Name:        "foundation_models_on_device",
			Description: "Foundation models on-device patterns for running LLMs locally, privacy-preserving AI, and edge computing",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Foundation models on-device task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-pytorch-patterns",
			Name:        "pytorch_patterns",
			Description: "PyTorch patterns for deep learning, model training, optimization, and deployment",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "PyTorch pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-agent-eval",
			Name:        "agent_eval",
			Description: "Agent evaluation patterns for measuring AI agent performance, benchmarking, and quality metrics",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Agent evaluation task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-agent-harness-construction",
			Name:        "agent_harness_construction",
			Description: "Agent harness construction patterns for building test harnesses, evaluation frameworks, and agent orchestration",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Agent harness construction task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-continuous-agent-loop",
			Name:        "continuous_agent_loop",
			Description: "Continuous agent loop patterns for autonomous AI agents, self-improvement, and iterative task execution",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Continuous agent loop task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-continuous-learning",
			Name:        "continuous_learning",
			Description: "Continuous learning patterns for AI agents, knowledge accumulation, and experience replay",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Continuous learning task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-continuous-learning-v2",
			Name:        "continuous_learning_v2",
			Description: "Continuous learning v2 patterns for AI agents with improved knowledge management and experience tracking",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Continuous learning v2 task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-eval-harness",
			Name:        "eval_harness",
			Description: "Evaluation harness patterns for testing AI systems, benchmarking, and quality assurance",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Evaluation harness task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-agent-payment-x402",
			Name:        "agent_payment_x402",
			Description: "Agent payment patterns for monetizing AI agents, billing, and payment processing",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Agent payment task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-nanoclaw-repl",
			Name:        "nanoclaw_repl",
			Description: "Nanoclaw REPL patterns for interactive AI agent development and testing",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Nanoclaw REPL task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "ai-ml",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// MCP Skills
		{
			ID:          "best-mcp-server-patterns",
			Name:        "mcp_server_patterns",
			Description: "MCP server patterns for building Model Context Protocol servers, tool implementation, and client integration",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "MCP server pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "mcp",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-configure-ecc",
			Name:        "configure_ecc",
			Description: "Configure ECC patterns for error correction, validation, and configuration management",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Configure ECC task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "mcp",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Web & Frontend Skills
		{
			ID:          "best-frontend-slides",
			Name:        "frontend_slides",
			Description: "Frontend slides patterns for presentation development, slide decks, and interactive presentations",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Frontend slides task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "web-frontend",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-nextjs-turbopack",
			Name:        "nextjs_turbopack",
			Description: "Next.js Turbopack patterns for fast builds, hot module replacement, and development optimization",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Next.js Turbopack task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "web-frontend",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-nuxt4-patterns",
			Name:        "nuxt4_patterns",
			Description: "Nuxt 4 patterns for Vue.js development, SSR, and modern web applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Nuxt 4 pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "web-frontend",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-react-agent-pattern",
			Name:        "react_agent_pattern",
			Description: "React agent patterns for building AI-powered React components and integrations",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "React agent pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "web-frontend",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-browser-qa",
			Name:        "browser_qa",
			Description: "Browser QA patterns for cross-browser testing, compatibility checks, and quality assurance",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Browser QA task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "web-frontend",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-browserbase-browser-automation",
			Name:        "browserbase_browser_automation",
			Description: "Browserbase browser automation patterns for headless browser testing, scraping, and automation",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Browserbase browser automation task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "web-frontend",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// DevOps & Infrastructure Skills
		{
			ID:          "best-benchmark",
			Name:        "benchmark",
			Description: "Benchmark patterns for performance testing, load testing, and optimization measurement",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Benchmark task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "devops",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-canary-watch",
			Name:        "canary_watch",
			Description: "Canary deployment patterns for gradual rollouts, monitoring, and safe production releases",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Canary watch task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "devops",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-enterprise-agent-ops",
			Name:        "enterprise_agent_ops",
			Description: "Enterprise agent operations patterns for scaling, monitoring, and managing AI agent fleets",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Enterprise agent ops task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "devops",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Security Skills
		{
			ID:          "best-security-scan",
			Name:        "security_scan",
			Description: "Security scan patterns for vulnerability assessment, dependency checking, and security auditing",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Security scan task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "security",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "best-safety-guard",
			Name:        "safety_guard",
			Description: "Safety guard patterns for AI safety, content filtering, and risk mitigation",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Safety guard task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "security",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "best-healthcare-phi-compliance",
			Name:        "healthcare_phi_compliance",
			Description: "Healthcare PHI compliance patterns for HIPAA, data privacy, and secure healthcare applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Healthcare PHI compliance task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "security",
			Permissions: []string{"devin", "claude"},
		},

		// Domain-Specific Skills - Healthcare
		{
			ID:          "best-healthcare-cdss-patterns",
			Name:        "healthcare_cdss_patterns",
			Description: "Clinical Decision Support System patterns for healthcare applications, medical reasoning, and patient care",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Healthcare CDSS pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-healthcare",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-healthcare-emr-patterns",
			Name:        "healthcare_emr_patterns",
			Description: "Electronic Medical Record patterns for healthcare systems, patient data management, and clinical workflows",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Healthcare EMR pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-healthcare",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-healthcare-eval-harness",
			Name:        "healthcare_eval_harness",
			Description: "Healthcare evaluation harness patterns for medical AI testing, clinical validation, and quality assurance",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Healthcare eval harness task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-healthcare",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Domain-Specific Skills - Logistics
		{
			ID:          "best-logistics-exception-management",
			Name:        "logistics_exception_management",
			Description: "Logistics exception management patterns for handling shipping issues, delays, and resolution workflows",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Logistics exception management task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-logistics",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Domain-Specific Skills - Energy
		{
			ID:          "best-energy-procurement",
			Name:        "energy_procurement",
			Description: "Energy procurement patterns for utility management, cost optimization, and energy trading",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Energy procurement task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-energy",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Domain-Specific Skills - Inventory
		{
			ID:          "best-inventory-demand-planning",
			Name:        "inventory_demand_planning",
			Description: "Inventory demand planning patterns for stock management, forecasting, and supply chain optimization",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Inventory demand planning task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-inventory",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Domain-Specific Skills - Production
		{
			ID:          "best-production-scheduling",
			Name:        "production_scheduling",
			Description: "Production scheduling patterns for manufacturing, resource allocation, and workflow optimization",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Production scheduling task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-production",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Domain-Specific Skills - Quality
		{
			ID:          "best-quality-nonconformance",
			Name:        "quality_nonconformance",
			Description: "Quality nonconformance patterns for defect tracking, root cause analysis, and quality improvement",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Quality nonconformance task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-quality",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Domain-Specific Skills - Customs
		{
			ID:          "best-customs-trade-compliance",
			Name:        "customs_trade_compliance",
			Description: "Customs trade compliance patterns for international shipping, regulatory compliance, and documentation",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Customs trade compliance task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-customs",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Domain-Specific Skills - Returns
		{
			ID:          "best-returns-reverse-logistics",
			Name:        "returns_reverse_logistics",
			Description: "Returns reverse logistics patterns for handling product returns, refunds, and reverse supply chain",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Returns reverse logistics task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-returns",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Domain-Specific Skills - Carrier
		{
			ID:          "best-carrier-relationship-management",
			Name:        "carrier_relationship_management",
			Description: "Carrier relationship management patterns for shipping partner management, rate negotiation, and performance tracking",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Carrier relationship management task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "domain-carrier",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Content & Media Skills
		{
			ID:          "best-article-writing",
			Name:        "article_writing",
			Description: "Article writing patterns for content creation, SEO optimization, and structured writing",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Article writing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "content-media",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-content-engine",
			Name:        "content_engine",
			Description: "Content engine patterns for automated content generation, management, and distribution",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Content engine task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "content-media",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-crosspost",
			Name:        "crosspost",
			Description: "Crosspost patterns for publishing content across multiple platforms and channels",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Crosspost task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "content-media",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-video-editing",
			Name:        "video_editing",
			Description: "Video editing patterns for content creation, post-production, and media optimization",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Video editing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "content-media",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-videodb",
			Name:        "videodb",
			Description: "VideoDB patterns for video storage, management, and retrieval",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "VideoDB task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "content-media",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Search & Research Skills
		{
			ID:          "best-deep-research",
			Name:        "deep_research",
			Description: "Deep research patterns for comprehensive information gathering, analysis, and synthesis",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Deep research task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "search-research",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-exa-search",
			Name:        "exa_search",
			Description: "Exa search patterns for semantic search, web indexing, and intelligent information retrieval",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Exa search task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "search-research",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-repo-scan",
			Name:        "repo_scan",
			Description: "Repository scan patterns for code analysis, dependency checking, and security auditing",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Repo scan task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "search-research",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-documentation-lookup",
			Name:        "documentation_lookup",
			Description: "Documentation lookup patterns for finding and retrieving technical documentation and API references",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Documentation lookup task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "search-research",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Agent & Workflow Skills
		{
			ID:          "best-agent-workflow-compliance",
			Name:        "agent_workflow_compliance",
			Description: "Agent workflow compliance patterns for ensuring AI agents follow regulations and best practices",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Agent workflow compliance task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "agent-workflow",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-agent-training-prompt",
			Name:        "agent_training_prompt",
			Description: "Agent training prompt patterns for designing effective prompts and training AI agents",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Agent training prompt task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "agent-workflow",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-autonomous-loops",
			Name:        "autonomous_loops",
			Description: "Autonomous loop patterns for self-running AI agents, continuous improvement, and autonomous task execution",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Autonomous loops task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "agent-workflow",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-claude-devfleet",
			Name:        "claude_devfleet",
			Description: "Claude DevFleet patterns for managing fleets of Claude-powered agents and coordinated workflows",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Claude DevFleet task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "agent-workflow",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-dmux-workflows",
			Name:        "dmux_workflows",
			Description: "Dmux workflow patterns for distributed task execution and workflow orchestration",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Dmux workflows task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "agent-workflow",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-ralphinho-rfc-pipeline",
			Name:        "ralphinho_rfc_pipeline",
			Description: "Ralphinho RFC pipeline patterns for request for comment workflows and collaborative decision making",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Ralphinho RFC pipeline task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "agent-workflow",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-santa-method",
			Name:        "santa_method",
			Description: "Santa method patterns for gift-giving workflows, resource allocation, and distribution optimization",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Santa method task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "agent-workflow",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-team-builder",
			Name:        "team_builder",
			Description: "Team builder patterns for assembling AI agent teams, role assignment, and collaboration workflows",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Team builder task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "agent-workflow",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Tools & Utilities Skills
		{
			ID:          "best-ck",
			Name:        "ck",
			Description: "CK patterns for command-line toolkit and utility operations",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "CK task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-context-budget",
			Name:        "context_budget",
			Description: "Context budget patterns for managing token usage, context windows, and optimization",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Context budget task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-strategic-compact",
			Name:        "strategic_compact",
			Description: "Strategic compact patterns for information compression, summarization, and efficient communication",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Strategic compact task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-compaction-gate",
			Name:        "compaction_gate",
			Description: "Compaction gate patterns for filtering, validation, and quality control",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Compaction gate task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-rules-distill",
			Name:        "rules_distill",
			Description: "Rules distill patterns for extracting and condensing rules, guidelines, and best practices",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Rules distill task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-skill-comply",
			Name:        "skill_comply",
			Description: "Skill comply patterns for ensuring skills follow guidelines and compliance requirements",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Skill comply task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-skill-stocktake",
			Name:        "skill_stocktake",
			Description: "Skill stocktake patterns for inventorying, auditing, and managing skill collections",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Skill stocktake task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-token-budget-advisor",
			Name:        "token_budget_advisor",
			Description: "Token budget advisor patterns for optimizing token usage, cost management, and efficiency",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Token budget advisor task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-bun-runtime",
			Name:        "bun_runtime",
			Description: "Bun runtime patterns for JavaScript/TypeScript execution, testing, and development",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Bun runtime task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-content-hash-cache-pattern",
			Name:        "content_hash_cache_pattern",
			Description: "Content hash cache patterns for efficient caching, deduplication, and content integrity",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Content hash cache pattern task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-regex-vs-llm-structured-text",
			Name:        "regex_vs_llm_structured_text",
			Description: "Regex vs LLM structured text patterns for parsing, extraction, and text processing",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Regex vs LLM structured text task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-iterative-retrieval",
			Name:        "iterative_retrieval",
			Description: "Iterative retrieval patterns for information gathering, refinement, and multi-step search",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Iterative retrieval task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-openclaw-persona-forge",
			Name:        "openclaw_persona_forge",
			Description: "Openclaw persona forge patterns for creating and managing AI personas and character profiles",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Openclaw persona forge task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-liquid-glass-design",
			Name:        "liquid_glass_design",
			Description: "Liquid glass design patterns for UI/UX, transparency, and modern interface design",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Liquid glass design task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-fal-ai-media",
			Name:        "fal_ai_media",
			Description: "Fal AI media patterns for AI-powered media generation, processing, and optimization",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Fal AI media task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Document Processing Skills
		{
			ID:          "best-nutrient-document-processing",
			Name:        "nutrient_document_processing",
			Description: "Nutrient document processing patterns for nutrition data extraction, analysis, and management",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Nutrient document processing task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "document-processing",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "best-visa-doc-translate",
			Name:        "visa_doc_translate",
			Description: "Visa document translation patterns for visa application documents, translation, and localization",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Visa doc translate task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "document-processing",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// API Integration Skills
		{
			ID:          "best-x-api",
			Name:        "x_api",
			Description: "X API patterns for Twitter/X API integration, social media automation, and data processing",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "X API task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "api-integration",
			Permissions: []string{"devin", "claude", "cursor"},
		},

		// Router-Specific Skills
		{
			ID:          "best-router-admin-ops",
			Name:        "router_admin_ops",
			Description: "Router admin operations patterns for managing router configuration, monitoring, and maintenance",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Router admin ops task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "best-router-decision-engine",
			Name:        "router_decision_engine",
			Description: "Router decision engine patterns for model selection, routing logic, and provider orchestration",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Router decision engine task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "best-router-system-prompt-injection",
			Name:        "router_system_prompt_injection",
			Description: "Router system prompt injection patterns for customizing router behavior and system-level prompts",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Router system prompt injection task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "tools-utilities",
			Permissions: []string{"devin", "claude"},
		},

		// Commands
		{
			ID:          "cmd-build-fix",
			Name:        "build_fix",
			Description: "Fix build errors across multiple languages and frameworks",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Build fix task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-cpp-build",
			Name:        "cpp_build",
			Description: "Build C++ projects with CMake, Make, or other build systems",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "C++ build task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-go-build",
			Name:        "go_build_cmd",
			Description: "Build Go projects with go build",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Go build task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-gradle-build",
			Name:        "gradle_build",
			Description: "Build Java/Kotlin projects with Gradle",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Gradle build task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-kotlin-build",
			Name:        "kotlin_build",
			Description: "Build Kotlin projects with Gradle or Maven",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Kotlin build task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-rust-build",
			Name:        "rust_build",
			Description: "Build Rust projects with cargo",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Rust build task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-setup-pm2",
			Name:        "setup_pm2",
			Description: "Setup PM2 process manager for Node.js applications",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Setup PM2 task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-pm2",
			Name:        "pm2",
			Description: "PM2 process manager commands for managing Node.js processes",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "PM2 task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-code-review",
			Name:        "code_review",
			Description: "Perform code review across multiple languages and frameworks",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Code review task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-cpp-review",
			Name:        "cpp_review",
			Description: "Review C++ code for best practices, performance, and security",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "C++ review task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-go-review",
			Name:        "go_review",
			Description: "Review Go code for best practices, performance, and security",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Go review task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-kotlin-review",
			Name:        "kotlin_review",
			Description: "Review Kotlin code for best practices, performance, and security",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Kotlin review task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-python-review",
			Name:        "python_review",
			Description: "Review Python code for best practices, performance, and security",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Python review task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-rust-review",
			Name:        "rust_review",
			Description: "Review Rust code for best practices, performance, and security",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Rust review task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-harness-audit",
			Name:        "harness_audit",
			Description: "Audit evaluation harnesses for AI systems and testing frameworks",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Harness audit task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-cpp-test",
			Name:        "cpp_test",
			Description: "Run C++ tests with Google Test, Catch2, or other frameworks",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "C++ test task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-go-test",
			Name:        "go_test",
			Description: "Run Go tests with go test",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Go test task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-kotlin-test",
			Name:        "kotlin_test",
			Description: "Run Kotlin tests with Gradle or Maven",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Kotlin test task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-rust-test",
			Name:        "rust_test",
			Description: "Run Rust tests with cargo test",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Rust test task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-tdd",
			Name:        "tdd",
			Description: "Test-driven development workflow with test-first approach",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "TDD task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-test-coverage",
			Name:        "test_coverage",
			Description: "Generate and analyze test coverage reports",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Test coverage task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-save-session",
			Name:        "save_session",
			Description: "Save current AI agent session state for later resumption",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Save session task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-resume-session",
			Name:        "resume_session",
			Description: "Resume a previously saved AI agent session",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Resume session task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-sessions",
			Name:        "sessions",
			Description: "List and manage saved AI agent sessions",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Sessions task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-checkpoint",
			Name:        "checkpoint",
			Description: "Create a checkpoint of current work state",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Checkpoint task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-learn",
			Name:        "learn",
			Description: "Learn from previous sessions and improve agent capabilities",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Learn task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-plan",
			Name:        "plan_cmd",
			Description: "Create implementation plans for complex tasks",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Plan task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-multi-plan",
			Name:        "multi_plan",
			Description: "Create multiple alternative plans for a task",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Multi-plan task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-evolve",
			Name:        "evolve",
			Description: "Evolve and improve existing code or architecture",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Evolve task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-blueprint",
			Name:        "blueprint_cmd",
			Description: "Create detailed blueprints for multi-session projects",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Blueprint task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-research",
			Name:        "research_cmd",
			Description: "Conduct research on specific topics or technologies",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Research task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-aside",
			Name:        "aside",
			Description: "Set aside context for later reference",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Aside task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-multi-backend",
			Name:        "multi_backend",
			Description: "Execute backend tasks across multiple technologies",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Multi-backend task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-multi-frontend",
			Name:        "multi_frontend",
			Description: "Execute frontend tasks across multiple frameworks",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Multi-frontend task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-multi-execute",
			Name:        "multi_execute",
			Description: "Execute tasks across multiple environments or contexts",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Multi-execute task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-multi-workflow",
			Name:        "multi_workflow",
			Description: "Execute workflows across multiple systems or platforms",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Multi-workflow task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-orchestrate",
			Name:        "orchestrate",
			Description: "Orchestrate complex multi-step workflows",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Orchestrate task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-devfleet",
			Name:        "devfleet",
			Description: "Manage and coordinate fleet of AI agents",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Devfleet task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-claw",
			Name:        "claw",
			Description: "Execute claw operations for agent management",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Claw task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-verify",
			Name:        "verify",
			Description: "Verify code quality, security, and compliance",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Verify task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-quality-gate",
			Name:        "quality_gate",
			Description: "Run quality gate checks before deployment",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Quality gate task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-audit",
			Name:        "audit",
			Description: "Audit code, security, and compliance",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Audit task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude"},
		},
		{
			ID:          "cmd-eval",
			Name:        "eval",
			Description: "Evaluate AI systems, models, or agents",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Eval task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-docs",
			Name:        "docs",
			Description: "Generate or update documentation",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Docs task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-update-docs",
			Name:        "update_docs",
			Description: "Update existing documentation",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Update docs task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-update-codemaps",
			Name:        "update_codemaps",
			Description: "Update code maps and architecture documentation",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Update codemaps task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-learn-eval",
			Name:        "learn_eval",
			Description: "Evaluate learning progress and agent improvement",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Learn eval task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-instinct-export",
			Name:        "instinct_export",
			Description: "Export agent instincts and learned patterns",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Instinct export task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-instinct-import",
			Name:        "instinct_import",
			Description: "Import agent instincts and learned patterns",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Instinct import task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-instinct-status",
			Name:        "instinct_status",
			Description: "Check status of agent instincts and learned patterns",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Instinct status task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-prune",
			Name:        "prune",
			Description: "Prune and optimize learned patterns and instincts",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Prune task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-skill-create",
			Name:        "skill_create",
			Description: "Create new skills for AI agents",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Skill create task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-skill-health",
			Name:        "skill_health",
			Description: "Check health and status of agent skills",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Skill health task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-context-budget",
			Name:        "context_budget_cmd",
			Description: "Manage context budget for AI agents",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Context budget task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-strategic-compact",
			Name:        "strategic_compact_cmd",
			Description: "Compact information strategically for efficiency",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Strategic compact task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-loop-start",
			Name:        "loop_start",
			Description: "Start a continuous agent loop",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Loop start task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-loop-status",
			Name:        "loop_status",
			Description: "Check status of running agent loops",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Loop status task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-model-route",
			Name:        "model_route",
			Description: "Route requests to appropriate AI models",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Model route task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-projects",
			Name:        "projects",
			Description: "List and manage projects",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Projects task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-promote",
			Name:        "promote",
			Description: "Promote code or changes to next environment",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Promote task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-e2e",
			Name:        "e2e",
			Description: "Run end-to-end tests",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "E2E task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
		{
			ID:          "cmd-refactor-clean",
			Name:        "refactor_clean",
			Description: "Refactor and clean code",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task": map[string]interface{}{
						"type":        "string",
						"description": "Refactor clean task",
					},
				},
				"required": []string{"task"},
			},
			Category:    "command",
			Permissions: []string{"devin", "claude", "cursor"},
		},
	}
}

// RegisterDefaultTools registers default tools in TiBrain
func RegisterDefaultTools(hub *Hub) error {
	definitions := DefaultToolDefinitions()

	for _, def := range definitions {
		parametersJSON, err := json.Marshal(def.Parameters)
		if err != nil {
			return fmt.Errorf("marshal parameters for %s: %w", def.ID, err)
		}

		permissionsJSON, err := json.Marshal(def.Permissions)
		if err != nil {
			return fmt.Errorf("marshal permissions for %s: %w", def.ID, err)
		}

		toolReg := Tool{
			ID:                   def.ID,
			Name:                 def.Name,
			Description:          def.Description,
			Parameters:           string(parametersJSON),
			Handler:              "router",
			Category:             def.Category,
			Permissions:          string(permissionsJSON),
			Enabled:              true,
			QualityScore:         80,
			SecurityScore:        80,
			BestPracticesScore:   80,
			Source:               "best-source",
			Family:               "",
			Tags:                 "",
			Version:              "",
			SkillLevel:           "l2",
			QualityTier:          "platinum",
			SecurityTier:         "hardened",
			SecurityStatus:       "passed",
			ValidationStatus:     "passed",
			VariantID:            "",
			VariantLabel:         "",
			SourceType:           "community",
			RootPath:             "",
		}

		err = hub.RegisterTool(toolReg)
		if err != nil {
			return fmt.Errorf("register tool %s: %w", def.ID, err)
		}

		logger.Info("Tool registered: %s (%s)", def.ID, def.Name)
	}

	return nil
}
