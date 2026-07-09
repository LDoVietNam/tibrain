// Code graph implementation for TiBrain
// Provides persistent storage and querying of code structure (packages, files, functions, types, etc.)
package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// CodeGraphNode represents a node in the code graph
type CodeGraphNode struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // package, file, func, type, var, const, import, method
	Name        string                 `json:"name"`
	FilePath    string                 `json:"file_path"`
	LineStart   int                    `json:"line_start"`
	LineEnd     int                    `json:"line_end"`
	Docstring   string                 `json:"docstring"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// CodeGraphEdge represents an edge in the code graph
type CodeGraphEdge struct {
	ID          string                 `json:"id"`
	SourceID    string                 `json:"source_id"`
	TargetID    string                 `json:"target_id"`
	EdgeType    string                 `json:"edge_type"` // contains, imports, calls, references, defines_type, implements_interface, etc.
	Weight      float64                `json:"weight,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// CodeGraphService provides an interface for code graph operations
type CodeGraphService interface {
	// BuildOrUpdateGraph walks the codebase and builds/updates the code graph
	BuildOrUpdateGraph() error
	// GetNodeByID returns a node by its ID
	GetNodeByID(id string) (*CodeGraphNode, error)
	// GetNodesByType returns all nodes of a given type
	GetNodesByType(typ string) ([]*CodeGraphNode, error)
	// GetNodesByNamePattern returns nodes matching a name pattern (case-insensitive substring)
	GetNodesByNamePattern(pattern string) ([]*CodeGraphNode, error)
	// GetIncomingEdges returns edges pointing to the given node
	GetIncomingEdges(nodeID string) ([]*CodeGraphEdge, error)
	// GetOutgoingEdges returns edges originating from the given node
	GetOutgoingEdges(nodeID string) ([]*CodeGraphEdge, error)
	// FindCallers returns functions that call the given function
	FindCallers(functionID string) ([]*CodeGraphNode, error)
	// FindCallees returns functions called by the given function
	FindCallees(functionID string) ([]*CodeGraphNode, error)
	// GetFileSymbols returns all symbols defined in a given file
	GetFileSymbols(filePath string) ([]*CodeGraphNode, error)
	// GetStats returns statistics about the code graph
	GetStats() (*CodeGraphStats, error)
	// Close closes any resources
	Close() error
}

// CodeGraphStats holds statistics about the code graph
type CodeGraphStats struct {
	NodeCount int
	EdgeCount int
}

// codeGraphImpl implements CodeGraphService using the TiBrain hub database
type codeGraphImpl struct {
	hub *Hub
	mu  sync.RWMutex
}

// NewCodeGraphService creates a new code graph service
func NewCodeGraphService(hub *Hub) CodeGraphService {
	return &codeGraphImpl{
		hub: hub,
	}
}

// BuildOrUpdateGraph walks the codebase and builds/updates the code graph
func (c *codeGraphImpl) BuildOrUpdateGraph() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Define the root directory to scan (the tibrain source root)
	roots := []string{
		getTiBrainDir(),
	}

	// We'll collect all Go files
	var goFiles []string
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// Skip certain directories
				switch d.Name() {
				case ".git", "node_modules", "dist", "build", "vendor", "venv", ".venv", "__pycache__":
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
				goFiles = append(goFiles, path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	logger.Info("Building code graph from root: %s, found %d Go files", roots[0], len(goFiles))

	// Process each file
	for _, filePath := range goFiles {
		if err := c.processGoFile(filePath); err != nil {
			logger.Warn("Failed to process Go file %s: %v", filePath, err)
			continue
		}
	}

	logger.Info("Code graph built successfully")
	return nil
}

// processGoFile parses a single Go file and updates the code graph
func (c *codeGraphImpl) processGoFile(filePath string) error {
	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// Get package name
	pkgName := node.Name.String()

	// Create file node
	fileID := c.generateFileNodeID(filePath)
	fileNode := &CodeGraphNode{
		ID:       fileID,
		Type:     "file",
		Name:     filepath.Base(filePath),
		FilePath: filePath,
		LineStart: 1,
		LineEnd:  fset.Position(node.End()).Line,
	}

	// Upsert file node
	if err := c.upsertNode(fileNode); err != nil {
		return err
	}

	// Create package node (if not exists)
	pkgID := c.generatePackageNodeID(pkgName)
	pkgNode := &CodeGraphNode{
		ID:   pkgID,
		Type: "package",
		Name: pkgName,
	}

	// Upsert package node
	if err := c.upsertNode(pkgNode); err != nil {
		return err
	}

	// Create contains edge: package -> file
	containsID := c.generateEdgeID(pkgID, fileID, "contains")
	containsEdge := &CodeGraphEdge{
		ID:       containsID,
		SourceID: pkgID,
		TargetID: fileID,
		EdgeType: "contains",
		Weight:   1.0,
	}

	// Upsert contains edge
	if err := c.upsertEdge(containsEdge); err != nil {
		return err
	}

	// Process imports
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		if importPath == "" {
			continue
		}

		// Create import node (using a simple ID based on the import path)
		importID := c.generateImportNodeID(importPath)
		importNode := &CodeGraphNode{
			ID:   importID,
			Type: "import",
			Name: importPath,
		}

		// Upsert import node
		if err := c.upsertNode(importNode); err != nil {
			return err
		}

		// Create imports edge: file -> import
		importsID := c.generateEdgeID(fileID, importID, "imports")
		importsEdge := &CodeGraphEdge{
			ID:       importsID,
			SourceID: fileID,
			TargetID: importID,
			EdgeType: "imports",
			Weight:   1.0,
		}

		// Upsert imports edge
		if err := c.upsertEdge(importsEdge); err != nil {
			return err
		}
	}

	// Process declarations
	ast.Inspect(node, func(n ast.Node) bool {
		switch d := n.(type) {
		case *ast.FuncDecl:
			// Function declaration
			if err := c.processFuncDecl(fset, fileID, d); err != nil {
				// Continue processing other declarations
			}
			return true
		case *ast.GenDecl:
			// General declaration (const, type, var)
			if err := c.processGenDecl(fset, fileID, d); err != nil {
				// Continue processing other declarations
			}
			return true
		}
		return true
	})

	return nil
}

// processFuncDecl processes a function declaration
func (c *codeGraphImpl) processFuncDecl(fset *token.FileSet, fileID string, decl *ast.FuncDecl) error {
	// Get function name
	funcName := decl.Name.String()
	if funcName == "" {
		return nil
	}

	// Create function node
	funcID := c.generateFuncNodeID(fileID, funcName)
	funcNode := &CodeGraphNode{
		ID:       funcID,
		Type:     "func",
		Name:     funcName,
		FilePath: c.getFilePathFromID(fileID),
		LineStart: fset.Position(decl.Pos()).Line,
		LineEnd:   fset.Position(decl.End()).Line,
		Docstring: c.extractDocstring(decl.Doc),
	}

	// Upsert function node
	if err := c.upsertNode(funcNode); err != nil {
		return err
	}

	// Create defines edge: file -> function
	definesID := c.generateEdgeID(fileID, funcID, "defines")
	definesEdge := &CodeGraphEdge{
		ID:       definesID,
		SourceID: fileID,
		TargetID: funcID,
		EdgeType: "defines",
		Weight:   1.0,
	}

	// Upsert defines edge
	if err := c.upsertEdge(definesEdge); err != nil {
		return err
	}

	// TODO: Process function body to find calls to other functions
	// For now, we'll skip this to keep the initial implementation simple

	return nil
}

// processGenDecl processes a general declaration (const, type, var)
func (c *codeGraphImpl) processGenDecl(fset *token.FileSet, fileID string, decl *ast.GenDecl) error {
	switch decl.Tok {
	case token.TYPE:
		// Type declarations
		for _, spec := range decl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				typeName := typeSpec.Name.String()
				if typeName == "" {
					continue
				}

				// Create type node
				typeID := c.generateTypeNodeID(fileID, typeName)
				typeNode := &CodeGraphNode{
					ID:       typeID,
					Type:     "type",
					Name:     typeName,
					FilePath: c.getFilePathFromID(fileID),
					LineStart: fset.Position(decl.Pos()).Line,
					LineEnd:   fset.Position(decl.End()).Line,
					Docstring: c.extractDocstring(decl.Doc),
				}

				// Upsert type node
				if err := c.upsertNode(typeNode); err != nil {
					return err
				}

				// Create defines edge: file -> type
				definesID := c.generateEdgeID(fileID, typeID, "defines")
				definesEdge := &CodeGraphEdge{
					ID:       definesID,
					SourceID: fileID,
					TargetID: typeID,
					EdgeType: "defines",
					Weight:   1.0,
				}

				// Upsert defines edge
				if err := c.upsertEdge(definesEdge); err != nil {
					return err
				}
			}
		}
	case token.CONST:
		// Constant declarations
		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range valueSpec.Names {
					constName := name.String()
					if constName == "" {
						continue
					}

					// Create constant node
					constID := c.generateConstNodeID(fileID, constName)
					constNode := &CodeGraphNode{
						ID:       constID,
						Type:     "const",
						Name:     constName,
						FilePath: c.getFilePathFromID(fileID),
						LineStart: fset.Position(decl.Pos()).Line,
						LineEnd:   fset.Position(decl.End()).Line,
						Docstring: c.extractDocstring(decl.Doc),
					}

					// Upsert constant node
					if err := c.upsertNode(constNode); err != nil {
						return err
					}

					// Create defines edge: file -> constant
					definesID := c.generateEdgeID(fileID, constID, "defines")
					definesEdge := &CodeGraphEdge{
						ID:       definesID,
						SourceID: fileID,
						TargetID: constID,
						EdgeType: "defines",
						Weight:   1.0,
					}

					// Upsert defines edge
					if err := c.upsertEdge(definesEdge); err != nil {
						return err
					}
				}
			}
		}
	case token.VAR:
		// Variable declarations
		for _, spec := range decl.Specs {
			if valueSpec, ok := spec.(*ast.ValueSpec); ok {
				for _, name := range valueSpec.Names {
					varName := name.String()
					if varName == "" {
						continue
					}

					// Create variable node
					varID := c.generateVarNodeID(fileID, varName)
					varNode := &CodeGraphNode{
						ID:       varID,
						Type:     "var",
						Name:     varName,
						FilePath: c.getFilePathFromID(fileID),
						LineStart: fset.Position(decl.Pos()).Line,
						LineEnd:   fset.Position(decl.End()).Line,
						Docstring: c.extractDocstring(decl.Doc),
					}

					// Upsert variable node
					if err := c.upsertNode(varNode); err != nil {
						return err
					}

					// Create defines edge: file -> variable
					definesID := c.generateEdgeID(fileID, varID, "defines")
					definesEdge := &CodeGraphEdge{
						ID:       definesID,
						SourceID: fileID,
						TargetID: varID,
						EdgeType: "defines",
						Weight:   1.0,
					}

					// Upsert defines edge
					if err := c.upsertEdge(definesEdge); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// extractDocstring extracts the docstring from a comment group
func (c *codeGraphImpl) extractDocstring(comment *ast.CommentGroup) string {
	if comment == nil {
		return ""
	}
	return strings.TrimSpace(comment.Text())
}

// Helper functions to generate IDs

func (c *codeGraphImpl) generateFileNodeID(filePath string) string {
	// Simple hash of the file path
	return "file-" + hashString(filePath)
}

func (c *codeGraphImpl) generatePackageNodeID(pkgName string) string {
	return "pkg-" + hashString(pkgName)
}

func (c *codeGraphImpl) generateImportNodeID(importPath string) string {
	return "import-" + hashString(importPath)
}

func (c *codeGraphImpl) generateFuncNodeID(fileID, funcName string) string {
	return "func-" + hashString(fileID+"."+funcName)
}

func (c *codeGraphImpl) generateTypeNodeID(fileID, typeName string) string {
	return "type-" + hashString(fileID+"."+typeName)
}

func (c *codeGraphImpl) generateConstNodeID(fileID, constName string) string {
	return "const-" + hashString(fileID+"."+constName)
}

func (c *codeGraphImpl) generateVarNodeID(fileID, varName string) string {
	return "var-" + hashString(fileID+"."+varName)
}

func (c *codeGraphImpl) generateEdgeID(sourceID, targetID, edgeType string) string {
	return "edge-" + hashString(sourceID+"->"+targetID+":"+edgeType)
}

// getFilePathFromID extracts the file path from a file node ID
// This is a simplified implementation - in practice, we'd store this mapping
func (c *codeGraphImpl) GetStats() (*CodeGraphStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.hub == nil || c.hub.db == nil {
		return &CodeGraphStats{}, nil
	}

	var stats CodeGraphStats
	err := c.hub.db.QueryRow("SELECT COUNT(*) FROM code_graph_nodes").Scan(&stats.NodeCount)
	if err != nil {
		stats.NodeCount = 0
	}
	err = c.hub.db.QueryRow("SELECT COUNT(*) FROM code_graph_edges").Scan(&stats.EdgeCount)
	if err != nil {
		stats.EdgeCount = 0
	}
	return &stats, nil
}

func (c *codeGraphImpl) getFilePathFromID(nodeID string) string {
	// For now, we'll return empty string and improve this later
	// In a full implementation, we'd query the database or maintain a map
	return ""
}

// codeGraphHashString creates a simple hash of a string
func codeGraphHashString(s string) string {
	// Simple hash function - in practice, we might use something better
	hash := 0
	for i := 0; i < len(s); i++ {
		hash = 31*hash + int(s[i])
	}
	if hash < 0 {
		hash = -hash
	}
	return strconv.Itoa(hash)
}

// upsertNode inserts or updates a node in the database
func (c *codeGraphImpl) upsertNode(node *CodeGraphNode) error {
	// Convert properties to JSON
	propertiesJSON := "{}"
	if node.Properties != nil {
		// In a real implementation, we'd marshal to JSON
		// For now, we'll use an empty object
		propertiesJSON = "{}"
	}

	_, err := c.hub.db.Exec(`
		INSERT OR REPLACE INTO code_graph_nodes
		(id, type, name, file_path, line_start, line_end, docstring, properties)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, node.ID, node.Type, node.Name, node.FilePath, node.LineStart, node.LineEnd, node.Docstring, propertiesJSON)
	return err
}

// upsertEdge inserts or updates an edge in the database
func (c *codeGraphImpl) upsertEdge(edge *CodeGraphEdge) error {
	// Convert properties to JSON
	propertiesJSON := "{}"
	if edge.Properties != nil {
		// In a real implementation, we'd marshal to JSON
		// For now, we'll use an empty object
		propertiesJSON = "{}"
	}

	_, err := c.hub.db.Exec(`
		INSERT OR REPLACE INTO code_graph_edges
		(id, source_id, target_id, edge_type, weight, properties)
		VALUES (?, ?, ?, ?, ?, ?)
	`, edge.ID, edge.SourceID, edge.TargetID, edge.EdgeType, edge.Weight, propertiesJSON)
	return err
}

// GetNodeByID returns a node by its ID
func (c *codeGraphImpl) GetNodeByID(id string) (*CodeGraphNode, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var node CodeGraphNode
	var properties string
	err := c.hub.db.QueryRow(`
		SELECT id, type, name, file_path, line_start, line_end, docstring, properties
		FROM code_graph_nodes WHERE id = ?
	`, id).Scan(&node.ID, &node.Type, &node.Name, &node.FilePath, &node.LineStart, &node.LineEnd, &node.Docstring, &properties)
	if err != nil {
		return nil, err
	}

	// Parse properties JSON (simplified)
	node.Properties = make(map[string]interface{})
	// In a real implementation, we'd unmarshal the JSON

	return &node, nil
}

// GetNodesByType returns all nodes of a given type
func (c *codeGraphImpl) GetNodesByType(typ string) ([]*CodeGraphNode, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.hub.db.Query(`
		SELECT id, type, name, file_path, line_start, line_end, docstring, properties
		FROM code_graph_nodes WHERE type = ?
	`, typ)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*CodeGraphNode
	for rows.Next() {
		var node CodeGraphNode
		var properties string
		if err := rows.Scan(&node.ID, &node.Type, &node.Name, &node.FilePath, &node.LineStart, &node.LineEnd, &node.Docstring, &properties); err != nil {
			return nil, err
		}
		// Parse properties JSON (simplified)
		node.Properties = make(map[string]interface{})
		nodes = append(nodes, &node)
	}
	return nodes, nil
}

// GetNodesByNamePattern returns nodes matching a name pattern (case-insensitive substring)
func (c *codeGraphImpl) GetNodesByNamePattern(pattern string) ([]*CodeGraphNode, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.hub.db.Query(`
		SELECT id, type, name, file_path, line_start, line_end, docstring, properties
		FROM code_graph_nodes WHERE LOWER(name) LIKE LOWER(?)
	`, "%"+pattern+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*CodeGraphNode
	for rows.Next() {
		var node CodeGraphNode
		var properties string
		if err := rows.Scan(&node.ID, &node.Type, &node.Name, &node.FilePath, &node.LineStart, &node.LineEnd, &node.Docstring, &properties); err != nil {
			return nil, err
		}
		// Parse properties JSON (simplified)
		node.Properties = make(map[string]interface{})
		nodes = append(nodes, &node)
	}
	return nodes, nil
}

// GetIncomingEdges returns edges pointing to the given node
func (c *codeGraphImpl) GetIncomingEdges(nodeID string) ([]*CodeGraphEdge, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.hub.db.Query(`
		SELECT id, source_id, target_id, edge_type, weight, properties
		FROM code_graph_edges WHERE target_id = ?
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*CodeGraphEdge
	for rows.Next() {
		var edge CodeGraphEdge
		var properties string
		if err := rows.Scan(&edge.ID, &edge.SourceID, &edge.TargetID, &edge.EdgeType, &edge.Weight, &properties); err != nil {
			return nil, err
		}
		// Parse properties JSON (simplified)
		edge.Properties = make(map[string]interface{})
		edges = append(edges, &edge)
	}
	return edges, nil
}

// GetOutgoingEdges returns edges originating from the given node
func (c *codeGraphImpl) GetOutgoingEdges(nodeID string) ([]*CodeGraphEdge, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.hub.db.Query(`
		SELECT id, source_id, target_id, edge_type, weight, properties
		FROM code_graph_edges WHERE source_id = ?
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*CodeGraphEdge
	for rows.Next() {
		var edge CodeGraphEdge
		var properties string
		if err := rows.Scan(&edge.ID, &edge.SourceID, &edge.TargetID, &edge.EdgeType, &edge.Weight, &properties); err != nil {
			return nil, err
		}
		// Parse properties JSON (simplified)
		edge.Properties = make(map[string]interface{})
		edges = append(edges, &edge)
	}
	return edges, nil
}

// FindCallers returns functions that call the given function
func (c *codeGraphImpl) FindCallers(functionID string) ([]*CodeGraphNode, error) {
	// TODO: Implement call graph traversal
	// For now, return empty slice
	return []*CodeGraphNode{}, nil
}

// FindCallees returns functions called by the given function
func (c *codeGraphImpl) FindCallees(functionID string) ([]*CodeGraphNode, error) {
	// TODO: Implement call graph traversal
	// For now, return empty slice
	return []*CodeGraphNode{}, nil
}

// GetFileSymbols returns all symbols defined in a given file
func (c *codeGraphImpl) GetFileSymbols(filePath string) ([]*CodeGraphNode, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rows, err := c.hub.db.Query(`
		SELECT id, type, name, file_path, line_start, line_end, docstring, properties
		FROM code_graph_nodes WHERE file_path = ? AND type IN ('func', 'type', 'var', 'const')
	`, filePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*CodeGraphNode
	for rows.Next() {
		var node CodeGraphNode
		var properties string
		if err := rows.Scan(&node.ID, &node.Type, &node.Name, &node.FilePath, &node.LineStart, &node.LineEnd, &node.Docstring, &properties); err != nil {
			return nil, err
		}
		// Parse properties JSON (simplified)
		node.Properties = make(map[string]interface{})
		nodes = append(nodes, &node)
	}
	return nodes, nil
}

// Close closes any resources
func (c *codeGraphImpl) Close() error {
	return nil
}