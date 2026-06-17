// TiBrain Retrieval Router
// Routes queries within the unified Ti Brain knowledge base (Router Brain merged).
package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// RetrievalRoute determines where to send a query
type RetrievalRoute string

const (
	RouteGlobalRAG   RetrievalRoute = "global_rag"
	RouteRouterBrain RetrievalRoute = "router_brain" // Merged: uses same local RAG
	RouteHybrid      RetrievalRoute = "hybrid"
	RouteBoth        RetrievalRoute = "both"
	RouteGraph       RetrievalRoute = "graph" // Neo4j graph-based queries
)

// RetrievalRouter handles intelligent query routing within unified Ti Brain
type RetrievalRouter struct {
	hub             *Hub
	ragManager      *RAGSystemManager
	graphStore      *Neo4jGraphStore
	routerRules     []RetrievalRouterRoutingRule
	learningEnabled bool
}

// RoutingRule defines a rule for query routing
// (Defined in ecosystem_integration.go - removed duplicate)
type RetrievalRouterRoutingRule struct {
	Name        string
	Pattern     string
	Priority    int
	Route       RetrievalRoute
	Description string
}

// RoutingDecision contains the routing decision and metadata
type RoutingDecision struct {
	Route          RetrievalRoute `json:"route"`
	Confidence     float64        `json:"confidence"`
	Reasoning      string         `json:"reasoning"`
	RuleMatched    string         `json:"rule_matched"`
	ShouldFallback bool           `json:"should_fallback"`
}

// StandardizedRAGResponse is the unified response format
type StandardizedRAGResponse struct {
	Success    bool                   `json:"success"`
	Query      string                 `json:"query"`
	Answer     string                 `json:"answer,omitempty"`
	Results    []StandardizedResult   `json:"results"`
	Sources    []Source               `json:"sources,omitempty"`
	Confidence float64                `json:"confidence"`
	DurationMs int64                  `json:"duration_ms"`
	Route      RetrievalRoute         `json:"route"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
	Error      string                 `json:"error,omitempty"`
}

// StandardizedResult is a unified result format
type StandardizedResult struct {
	Content  string                 `json:"content"`
	Score    float64                `json:"score"`
	Source   string                 `json:"source"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Source represents a document source
type Source struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Path     string `json:"path"`
	Category string `json:"category"`
}

// NewRetrievalRouter creates a new retrieval router (Router Brain merged)
func NewRetrievalRouter(hub *Hub) *RetrievalRouter {
	router := &RetrievalRouter{
		hub:             hub,
		ragManager:      NewRAGSystemManager(hub),
		learningEnabled: true,
	}

	router.initializeRoutingRules()
	return router
}

// NewRetrievalRouterWithGraph creates a retrieval router with Neo4j graph support
func NewRetrievalRouterWithGraph(hub *Hub, graphStore *Neo4jGraphStore) *RetrievalRouter {
	router := &RetrievalRouter{
		hub:        hub,
		ragManager: NewRAGSystemManager(hub),
		graphStore: graphStore,
		learningEnabled: true,
	}

	router.initializeRoutingRules()
	return router
}

// initializeRoutingRules sets up the default routing rules
func (rr *RetrievalRouter) initializeRoutingRules() {
	rr.routerRules = []RetrievalRouterRoutingRule{
		{
			Name:        "Graph relationship queries",
			Pattern:     "(?i)(relationship|entity|connected|linked|related|graph|neo4j)",
			Priority:    1,
			Route:       RouteGraph,
			Description: "Queries about entity relationships and graph traversal",
		},
		{
			Name:        "Router-specific queries",
			Pattern:     "(?i)(router|routing|agent.*brain|multi.*cli|handoff|coordination)",
			Priority:    2,
			Route:       RouteRouterBrain,
			Description: "Queries about router system, agent brain, or multi-CLI coordination",
		},
		{
			Name:        "Ti ecosystem architecture",
			Pattern:     "(?i)(architecture|structure|design|system.*overview|ti.*ecosystem)",
			Priority:    3,
			Route:       RouteGlobalRAG,
			Description: "Queries about Ti ecosystem architecture and design",
		},
		{
			Name:        "Documentation and guides",
			Pattern:     "(?i)(documentation|guide|tutorial|how.*to|example)",
			Priority:    4,
			Route:       RouteGlobalRAG,
			Description: "Queries seeking documentation, guides, or tutorials",
		},
		{
			Name:        "Real-time operations",
			Pattern:     "(?i)(status|active|running|current|now|live)",
			Priority:    5,
			Route:       RouteRouterBrain,
			Description: "Queries requiring real-time operational data",
		},
		{
			Name:        "Historical knowledge",
			Pattern:     "(?i)(history|past|previous|archive|log)",
			Priority:    6,
			Route:       RouteGlobalRAG,
			Description: "Queries about historical data and archives",
		},
		{
			Name:        "Complex multi-domain queries",
			Pattern:     "(?i).{50,}", // Long queries
			Priority:    7,
			Route:       RouteHybrid,
			Description: "Complex queries that may benefit from multiple sources",
		},
		{
			Name:        "Default fallback",
			Pattern:     "(?i).*",
			Priority:    10,
			Route:       RouteGlobalRAG,
			Description: "Default: query Ti Brain local knowledge base",
		},
	}
}

// RouteQuery determinIntes the best route for a query
func (rr *RetrievalRouter) RouteQuery(ctx context.Context, query string) *RoutingDecision {
	decision := &RoutingDecision{
		Route:          RouteGlobalRAG,
		Confidence:     0.5,
		Reasoning:      "No specific rule matched, using default global RAG route",
		ShouldFallback: false,
	}

	// Check routing rules in priority order
	for _, rule := range rr.routerRules {
		matched, err := regexp.MatchString(rule.Pattern, query)
		if err != nil {
			logger.Warn("Failed to match routing rule %s: %v", rule.Name, err)
			continue
		}

		if matched {
			decision.Route = rule.Route
			decision.RuleMatched = rule.Name
			decision.Reasoning = fmt.Sprintf("Matched rule: %s - %s", rule.Name, rule.Description)
			decision.Confidence = 0.8
			break
		}
	}

	// Apply learning-based adjustments if enabled
	if rr.learningEnabled {
		rr.applyLearningAdjustments(ctx, query, decision)
	}

	logger.Info("Routing decision: query='%s' route=%s confidence=%.2f reason=%s",
		truncateQuery(query, 50), decision.Route, decision.Confidence, decision.Reasoning)

	return decision
}

// applyLearningAdjustments adjusts routing based on historical performance
func (rr *RetrievalRouter) applyLearningAdjustments(ctx context.Context, query string, decision *RoutingDecision) {
	// Check historical performance for similar queries
	// This would query the agent_performance table to see which routes performed well
	// For now, this is a placeholder for future learning implementation
}

// ExecuteRoute executes the query using the determinInted route
func (rr *RetrievalRouter) ExecuteRoute(ctx context.Context, query string, decision *RoutingDecision, scopes []string, contextSrc *AgentContextSource) (*StandardizedRAGResponse, error) {
	return rr.ExecuteRouteWithFilters(ctx, query, decision, scopes, contextSrc, nil)
}

func (rr *RetrievalRouter) ExecuteRouteWithFilters(ctx context.Context, query string, decision *RoutingDecision, scopes []string, contextSrc *AgentContextSource, metadataFilters map[string]string) (*StandardizedRAGResponse, error) {
	if decision == nil {
		decision = &RoutingDecision{
			Route:          RouteGlobalRAG,
			Confidence:     0.5,
			Reasoning:      "No routing decision provided, using local RAG",
			ShouldFallback: true,
		}
	}

	startTime := time.Now()
	response := &StandardizedRAGResponse{
		Query:     query,
		Route:     decision.Route,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	switch decision.Route {
	case RouteGraph:
		result, err := rr.executeGraph(ctx, query)
		if err != nil {
			// Fallback to global RAG if graph fails
			fallback, fallbackErr := rr.executeGlobalRAG(ctx, query, scopes, contextSrc, metadataFilters)
			if fallbackErr != nil {
				response.Error = fmt.Sprintf("graph search failed: %v; global RAG fallback failed: %v", err, fallbackErr)
				response.Success = false
				return response, err
			}
			response.Route = RouteGlobalRAG
			response.Metadata["fallback_from"] = string(RouteGraph)
			response.Metadata["fallback_error"] = err.Error()
			response.mergeFromGlobalRAG(fallback)
			break
		}
		response.mergeFromGraph(result)

	case RouteGlobalRAG:
		result, err := rr.executeGlobalRAG(ctx, query, scopes, contextSrc, metadataFilters)
		if err != nil {
			response.Error = err.Error()
			response.Success = false
			return response, err
		}
		response.mergeFromGlobalRAG(result)

	case RouteRouterBrain:
		result, err := rr.executeRouterBrain(ctx, query, scopes, contextSrc, metadataFilters)
		if err != nil {
			fallback, fallbackErr := rr.executeGlobalRAG(ctx, query, scopes, contextSrc, metadataFilters)
			if fallbackErr != nil {
				response.Error = fmt.Sprintf("router brain failed: %v; global RAG fallback failed: %v", err, fallbackErr)
				response.Success = false
				return response, err
			}
			response.Route = RouteGlobalRAG
			response.Metadata["fallback_from"] = string(RouteRouterBrain)
			response.Metadata["fallback_error"] = err.Error()
			response.mergeFromGlobalRAG(fallback)
			break
		}
		response.mergeFromRouterBrain(result)

	case RouteHybrid:
		// Try both and merge
		globalResult, globalErr := rr.executeGlobalRAG(ctx, query, scopes, contextSrc, metadataFilters)
		routerResult, routerErr := rr.executeRouterBrain(ctx, query, scopes, contextSrc, metadataFilters)

		if globalErr != nil && routerErr != nil {
			response.Error = fmt.Sprintf("Both routes failed: global=%v, router=%v", globalErr, routerErr)
			response.Success = false
			return response, nil
		}

		response.mergeHybrid(globalResult, routerResult, globalErr, routerErr)

	case RouteBoth:
		// Query both and return all results
		globalResult, globalErr := rr.executeGlobalRAG(ctx, query, scopes, contextSrc, metadataFilters)
		routerResult, routerErr := rr.executeRouterBrain(ctx, query, scopes, contextSrc, metadataFilters)

		response.mergeBoth(globalResult, routerResult, globalErr, routerErr)

	default:
		response.Error = fmt.Sprintf("unsupported retrieval route: %s", decision.Route)
		response.Success = false
	}

	response.DurationMs = time.Since(startTime).Milliseconds()
	response.Success = response.Error == ""
	if response.Confidence == 0 {
		response.Confidence = decision.Confidence
	}

	// Log the routing decision for learning
	rr.logRoutingDecision(query, decision, response)

	return response, nil
}

// executeGlobalRAG executes query against global RAG
func (rr *RetrievalRouter) executeGlobalRAG(ctx context.Context, query string, scopes []string, contextSrc *AgentContextSource, metadataFilters map[string]string) (*StandardizedRAGResponse, error) {
	ragQuery := RAGQuery{
		Query:      query,
		Metadata:   make(map[string]string),
		Timestamp:  time.Now(),
		Scopes:     scopes,
		ContextSrc: contextSrc,
	}
	for key, value := range metadataFilters {
		ragQuery.Metadata[key] = value
	}

	result, err := rr.ragManager.ProcessQueryWithContext(ctx, ragQuery)
	if err != nil {
		return nil, fmt.Errorf("global RAG error: %w", err)
	}

	response := &StandardizedRAGResponse{
		Query:      query,
		Answer:     result.Response,
		Confidence: result.Confidence,
		Route:      RouteGlobalRAG,
		Timestamp:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}
	if result.Runtime != nil {
		response.Metadata["runtime"] = result.Runtime
	}

	// Convert documents to standardized format
	for _, docID := range result.Documents {
		doc, err := rr.ragManager.GetDocument(docID)
		if err != nil {
			logger.Warn("Failed to retrieve document %s: %v", docID, err)
			continue
		}

		response.Results = append(response.Results, StandardizedResult{
			Content: doc.Content,
			Score:   result.Confidence,
			Source:  doc.Path,
			Metadata: map[string]interface{}{
				"id":       doc.ID,
				"title":    doc.Title,
				"category": doc.Category,
				"tags":     doc.Tags,
			},
		})

		response.Sources = append(response.Sources, Source{
			ID:       doc.ID,
			Title:    doc.Title,
			Path:     doc.Path,
			Category: doc.Category,
		})
	}

	return response, nil
}

// executeRouterBrain executes query against local RAG (Router Brain merged)
func (rr *RetrievalRouter) executeRouterBrain(ctx context.Context, query string, scopes []string, contextSrc *AgentContextSource, metadataFilters map[string]string) (*StandardizedRAGResponse, error) {
	// Router Brain is merged - use local RAG with router-specific prefix
	routerQuery := fmt.Sprintf("router: %s", query)
	return rr.executeGlobalRAG(ctx, routerQuery, scopes, contextSrc, metadataFilters)
}

// executeGraph executes query against Neo4j graph store
func (rr *RetrievalRouter) executeGraph(ctx context.Context, query string) (*StandardizedRAGResponse, error) {
	if rr.graphStore == nil {
		return nil, fmt.Errorf("graph store not initialized")
	}

	graphResults, err := rr.graphStore.SearchGraph(ctx, query, 2)
	if err != nil {
		return nil, fmt.Errorf("graph search failed: %w", err)
	}

	response := &StandardizedRAGResponse{
		Query:    query,
		Route:    RouteGraph,
		Timestamp: time.Now(),
		Metadata: make(map[string]interface{}),
	}

	// Convert graph results to standardized format
	for _, result := range graphResults {
		for _, node := range result.Nodes {
			response.Results = append(response.Results, StandardizedResult{
				Content: fmt.Sprintf("%v", node.Properties),
				Score:   result.Score,
				Source:  string(RouteGraph),
				Metadata: map[string]interface{}{
					"id":     node.ID,
					"labels": node.Labels,
					"type":   "node",
				},
			})
		}
		for _, rel := range result.Relationships {
			response.Results = append(response.Results, StandardizedResult{
				Content: fmt.Sprintf("%s -> %s (%s)", rel.StartNode, rel.EndNode, rel.Type),
				Score:   result.Score,
				Source:  string(RouteGraph),
				Metadata: map[string]interface{}{
					"id":        rel.ID,
					"type":      "relationship",
					"rel_type":  rel.Type,
				},
			})
		}
	}

	if len(response.Results) > 0 {
		response.Success = true
		response.Confidence = 0.8
	} else {
		response.Success = true
		response.Confidence = 0.1
	}

	return response, nil
}

// mergeFromGlobalRAG merges global RAG results
func (r *StandardizedRAGResponse) mergeFromGlobalRAG(other *StandardizedRAGResponse) {
	r.Results = other.Results
	r.Sources = other.Sources
	r.Answer = other.Answer
	r.Confidence = other.Confidence
}

// mergeFromRouterBrain merges Router Brain results (merged - same format)
func (r *StandardizedRAGResponse) mergeFromRouterBrain(other *StandardizedRAGResponse) {
	r.Results = other.Results
	if other.Answer != "" {
		r.Answer = other.Answer
	}
	r.Confidence = other.Confidence
	if other.Metadata != nil {
		r.Metadata["router_brain_info"] = other.Metadata
	}
}

// mergeFromGraph merges Neo4j graph results
func (r *StandardizedRAGResponse) mergeFromGraph(other *StandardizedRAGResponse) {
	r.Results = other.Results
	r.Confidence = other.Confidence
	r.Success = other.Success
	if other.Metadata != nil {
		r.Metadata["graph_info"] = other.Metadata
	}
}

// mergeHybrid merges results from both sources with hybrid logic
func (r *StandardizedRAGResponse) mergeHybrid(global, router *StandardizedRAGResponse, globalErr, routerErr error) {
	// Prefer Router Brain for real-time data, Global RAG for documentation
	resultsMap := make(map[string]StandardizedResult)

	if globalErr == nil && global != nil {
		for _, result := range global.Results {
			key := result.Source + result.Content[:minInt(50, len(result.Content))]
			resultsMap[key] = result
		}
	}

	if routerErr == nil && router != nil {
		for _, result := range router.Results {
			key := result.Source + result.Content[:minInt(50, len(result.Content))]
			// Router brain results get priority in hybrid mode
			resultsMap[key] = result
		}
	}

	for _, result := range resultsMap {
		r.Results = append(r.Results, result)
	}

	// Merge sources
	if globalErr == nil && global != nil {
		r.Sources = append(r.Sources, global.Sources...)
	}
	if routerErr == nil && router != nil {
		r.Sources = append(r.Sources, router.Sources...)
	}

	// Use Router Brain answer if available, otherwise Global RAG
	if routerErr == nil && router != nil && router.Answer != "" {
		r.Answer = router.Answer
		r.Confidence = router.Confidence
	} else if globalErr == nil && global != nil {
		r.Answer = global.Answer
		r.Confidence = global.Confidence
	}

	r.Metadata["hybrid_sources"] = map[string]interface{}{
		"global_rag":   globalErr == nil,
		"router_brain": routerErr == nil,
	}
}

// mergeBoth merges all results from both sources
func (r *StandardizedRAGResponse) mergeBoth(global, router *StandardizedRAGResponse, globalErr, routerErr error) {
	resultsMap := make(map[string]StandardizedResult)
	successfulSources := 0
	confidenceTotal := 0.0

	if globalErr == nil && global != nil {
		for _, result := range global.Results {
			key := resultKey("global", result)
			resultsMap[key] = result
		}
		r.Sources = append(r.Sources, global.Sources...)
		successfulSources++
		confidenceTotal += global.Confidence
	}

	if routerErr == nil && router != nil {
		for _, result := range router.Results {
			key := resultKey("router", result)
			resultsMap[key] = result
		}
		r.Sources = append(r.Sources, router.Sources...)
		successfulSources++
		confidenceTotal += router.Confidence
	}

	for _, result := range resultsMap {
		r.Results = append(r.Results, result)
	}

	// Combine answers
	answers := []string{}
	if globalErr == nil && global != nil && global.Answer != "" {
		answers = append(answers, fmt.Sprintf("Global RAG: %s", global.Answer))
	}
	if routerErr == nil && router != nil && router.Answer != "" {
		answers = append(answers, fmt.Sprintf("Router Brain: %s", router.Answer))
	}

	if len(answers) > 0 {
		r.Answer = strings.Join(answers, "\n\n")
	}
	if successfulSources > 0 {
		r.Confidence = confidenceTotal / float64(successfulSources)
	}

	r.Metadata["both_sources"] = map[string]interface{}{
		"global_rag":   globalErr == nil,
		"router_brain": routerErr == nil,
		"global_error": errorString(globalErr),
		"router_error": errorString(routerErr),
	}
}

// logRoutingDecision logs routing decisions for learning
func (rr *RetrievalRouter) logRoutingDecision(query string, decision *RoutingDecision, response *StandardizedRAGResponse) {
	if rr == nil || rr.hub == nil || rr.hub.db == nil || decision == nil || response == nil {
		return
	}

	// Store in routing_decision_log table for learning asynchronously
	rr.hub.asyncWriter.Enqueue(`
		INSERT INTO routing_decision_log
		(id, query, route, confidence, reasoning, rule_matched, success, duration_ms, result_count, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, generateID(), query, string(decision.Route), decision.Confidence,
		decision.Reasoning, decision.RuleMatched, response.Success,
		response.DurationMs, len(response.Results), time.Now().Unix())
}

func resultKey(prefix string, result StandardizedResult) string {
	return prefix + ":" + result.Source + ":" + result.Content[:minInt(50, len(result.Content))]
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// truncateQuery truncates query for logging
func truncateQuery(query string, maxLen int) string {
	if len(query) <= maxLen {
		return query
	}
	return query[:maxLen] + "..."
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
