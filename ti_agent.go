// Ti Agent Orchestrator
// Central orchestrator that coordinates all agent interactions through TiBrain
// Instead of agents handling processing independently, they all go through this orchestrator
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ti/router/tibrain/internal/memory"
)

// TiAgentOrchestrator is the central orchestrator for all agent interactions
type TiAgentOrchestrator struct {
	hub             *Hub
	retrievalRouter *RetrievalRouter
	ragManager      *RAGSystemManager
	cognitiveMemory *memory.CognitiveMemoryManager
}

// AgentRequest represents a request from any agent to the orchestrator
type AgentRequest struct {
	AgentID     string                 `json:"agent_id"`
	AgentType   string                 `json:"agent_type"`
	SessionID   string                 `json:"session_id"`
	Query       string                 `json:"query"`
	RequestType string                 `json:"request_type"` // "query", "task", "decision", "preference"
	Context     map[string]interface{} `json:"context"`
	Priority    int                    `json:"priority"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AgentResponse represents the response from the orchestrator to an agent
type AgentResponse struct {
	Success    bool                   `json:"success"`
	AgentID    string                 `json:"agent_id"`
	RequestID  string                 `json:"request_id"`
	Response   string                 `json:"response"`
	Data       map[string]interface{} `json:"data"`
	Route      string                 `json:"route"`
	Confidence float64                `json:"confidence"`
	DurationMs int64                  `json:"duration_ms"`
	Timestamp  time.Time              `json:"timestamp"`
	MemoryUsed []string               `json:"memory_used"`
	Error      string                 `json:"error,omitempty"`
}

// NewTiAgentOrchestrator creates a new Ti Agent orchestrator
func NewTiAgentOrchestrator(hub *Hub, retrievalRouter *RetrievalRouter) *TiAgentOrchestrator {
	return &TiAgentOrchestrator{
		hub:             hub,
		retrievalRouter: retrievalRouter,
		ragManager:      NewRAGSystemManager(hub),
	}
}

// ProcessAgentRequest processes a request from any agent through the orchestrator
func (tao *TiAgentOrchestrator) ProcessAgentRequest(ctx context.Context, request AgentRequest) (*AgentResponse, error) {
	startTime := time.Now()
	requestID := generateID()

	response := &AgentResponse{
		AgentID:    request.AgentID,
		RequestID:  requestID,
		Timestamp:  time.Now(),
		Data:       make(map[string]interface{}),
		MemoryUsed: []string{},
	}

	// Log the incoming request
	tao.logAgentRequest(request, requestID)

	// Route the request based on type
	switch request.RequestType {
	case "query":
		err := tao.handleQueryRequest(ctx, request, response)
		if err != nil {
			response.Error = err.Error()
			response.Success = false
		} else {
			response.Success = true
		}

	case "task":
		err := tao.handleTaskRequest(ctx, request, response)
		if err != nil {
			response.Error = err.Error()
			response.Success = false
		} else {
			response.Success = true
		}

	case "decision":
		err := tao.handleDecisionRequest(ctx, request, response)
		if err != nil {
			response.Error = err.Error()
			response.Success = false
		} else {
			response.Success = true
		}

	case "preference":
		err := tao.handlePreferenceRequest(ctx, request, response)
		if err != nil {
			response.Error = err.Error()
			response.Success = false
		} else {
			response.Success = true
		}

	default:
		// Default to query handling
		err := tao.handleQueryRequest(ctx, request, response)
		if err != nil {
			response.Error = err.Error()
			response.Success = false
		} else {
			response.Success = true
		}
	}

	response.DurationMs = time.Since(startTime).Milliseconds()

	// Log agent performance
	tao.logAgentPerformance(request, response)

	// Update agent status
	tao.updateAgentStatus(request.AgentID, response.Success)

	return response, nil
}

// handleQueryRequest handles query-type requests using the retrieval router
func (tao *TiAgentOrchestrator) handleQueryRequest(ctx context.Context, request AgentRequest, response *AgentResponse) error {
	if tao.retrievalRouter == nil {
		return fmt.Errorf("retrieval router not available")
	}

	// Use retrieval router for intelligent routing
	decision := tao.retrievalRouter.RouteQuery(ctx, request.Query)
	ragResponse, err := tao.retrievalRouter.ExecuteRoute(ctx, request.Query, decision, nil, nil)

	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}

	response.Response = ragResponse.Answer
	response.Route = string(ragResponse.Route)
	response.Confidence = ragResponse.Confidence
	response.Data["results"] = ragResponse.Results
	response.Data["sources"] = ragResponse.Sources
	response.Data["rag_metadata"] = ragResponse.Metadata
	response.MemoryUsed = append(response.MemoryUsed, "global_rag", "router_brain")

	return nil
}

// handleTaskRequest handles task-type requests using task memory
func (tao *TiAgentOrchestrator) handleTaskRequest(ctx context.Context, request AgentRequest, response *AgentResponse) error {
	// Create task in task memory
	taskID := generateID()
	timestamp := time.Now().Unix()

	_, err := tao.hub.db.Exec(`
		INSERT INTO task_memory
		(id, task_id, task_type, task_description, task_status, assigned_agent, priority, context, created_at, updated_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		taskID, taskID, request.AgentType, request.Query, "pending",
		request.AgentID, request.Priority, marshalMap(request.Context),
		timestamp, timestamp, marshalMap(request.Context))

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	// Process the task using RAG
	ragQuery := RAGQuery{
		Query:   request.Query,
		Context: fmt.Sprintf("Task from agent %s of type %s", request.AgentID, request.AgentType),
		Metadata: map[string]string{
			"task_id":    taskID,
			"agent_id":   request.AgentID,
			"agent_type": request.AgentType,
		},
		Timestamp: time.Now(),
	}

	result, err := tao.ragManager.ProcessQuery(ragQuery)
	if err != nil {
		return fmt.Errorf("task processing failed: %w", err)
	}

	// Update task with result
	_, err = tao.hub.db.Exec(`
		UPDATE task_memory
		SET task_status = 'completed', result = ?, updated_at = ?, completed_at = ?
		WHERE task_id = ?
	`, result.Response, time.Now().Unix(), time.Now().Unix(), taskID)

	if err != nil {
		logger.Warn("Failed to update task status: %v", err)
	}

	response.Response = result.Response
	response.Data["task_id"] = taskID
	response.Data["task_status"] = "completed"
	response.MemoryUsed = append(response.MemoryUsed, "task_memory")

	return nil
}

// handleDecisionRequest handles decision-type requests using decision memory
func (tao *TiAgentOrchestrator) handleDecisionRequest(ctx context.Context, request AgentRequest, response *AgentResponse) error {
	// Create decision record
	decisionID := generateID()
	timestamp := time.Now().Unix()

	// Query for relevant context
	ragQuery := RAGQuery{
		Query:   request.Query,
		Context: "Decision context for agent " + request.AgentID,
		Metadata: map[string]string{
			"decision_id": decisionID,
			"agent_id":    request.AgentID,
		},
		Timestamp: time.Now(),
	}

	result, err := tao.ragManager.ProcessQuery(ragQuery)
	if err != nil {
		return fmt.Errorf("decision context retrieval failed: %w", err)
	}

	// Store decision in decision memory
	_, err = tao.hub.db.Exec(`
		INSERT INTO decision_memory
		(id, decision_id, decision_type, decision_context, decision_outcome, decision_rationale, confidence, decision_maker, timestamp, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		decisionID, decisionID, request.AgentType, request.Query,
		result.Response, "Based on RAG retrieval", result.Confidence,
		request.AgentID, timestamp, marshalMap(request.Context))

	if err != nil {
		return fmt.Errorf("failed to store decision: %w", err)
	}

	response.Response = result.Response
	response.Confidence = result.Confidence
	response.Data["decision_id"] = decisionID
	response.Data["decision_rationale"] = "Based on RAG retrieval and context"
	response.MemoryUsed = append(response.MemoryUsed, "decision_memory")

	return nil
}

// handlePreferenceRequest handles preference-type requests using user preferences
func (tao *TiAgentOrchestrator) handlePreferenceRequest(ctx context.Context, request AgentRequest, response *AgentResponse) error {
	// Extract preference key and value from context
	preferenceKey, ok := request.Context["key"].(string)
	if !ok {
		return fmt.Errorf("preference key not provided in context")
	}

	preferenceValue, ok := request.Context["value"].(string)
	if !ok {
		return fmt.Errorf("preference value not provided in context")
	}

	preferenceType, ok := request.Context["type"].(string)
	if !ok {
		preferenceType = "string"
	}

	category, _ := request.Context["category"].(string)
	if category == "" {
		category = "general"
	}

	timestamp := time.Now().Unix()

	// Store preference
	_, err := tao.hub.db.Exec(`
		INSERT OR REPLACE INTO user_preferences
		(id, user_id, preference_key, preference_value, preference_type, category, valid_from, created_at, updated_at, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		generateID(), request.AgentID, preferenceKey, preferenceValue,
		preferenceType, category, timestamp, timestamp, timestamp,
		marshalMap(request.Context))

	if err != nil {
		return fmt.Errorf("failed to store preference: %w", err)
	}

	response.Response = fmt.Sprintf("Preference '%s' set to '%s' for agent %s", preferenceKey, preferenceValue, request.AgentID)
	response.Data["preference_key"] = preferenceKey
	response.Data["preference_value"] = preferenceValue
	response.Data["category"] = category
	response.MemoryUsed = append(response.MemoryUsed, "user_preferences")

	return nil
}

// logAgentRequest logs incoming agent requests
func (tao *TiAgentOrchestrator) logAgentRequest(request AgentRequest, requestID string) {
	metadataJSON := marshalMap(request.Context)

	tao.hub.asyncWriter.Enqueue(`
		INSERT INTO orchestration_log
		(id, query, sources, results, orchestration_time_ms, status, timestamp, user_id, session_id, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		requestID, request.Query, request.AgentType, "", 0, "started",
		time.Now().Unix(), request.AgentID, request.SessionID, metadataJSON)
}

// logAgentPerformance logs agent performance metrics
func (tao *TiAgentOrchestrator) logAgentPerformance(request AgentRequest, response *AgentResponse) {
	// Store in enhanced agent performance table asynchronously
	tao.hub.asyncWriter.Enqueue(`
		INSERT INTO agent_performance_enhanced
		(id, agent_id, session_id, performance_metric, metric_value, metric_unit, timestamp, context, performance_rating, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		generateID(), request.AgentID, request.SessionID,
		"request_duration_ms", float64(response.DurationMs), "ms",
		time.Now().Unix(), request.RequestType,
		getPerformanceRating(response.Success, response.DurationMs),
		marshalMap(map[string]interface{}{
			"request_id":  response.RequestID,
			"success":     response.Success,
			"route":       response.Route,
			"memory_used": response.MemoryUsed,
		}))
}

// updateAgentStatus updates the agent status in the registry
func (tao *TiAgentOrchestrator) updateAgentStatus(agentID string, success bool) {
	status := "active"
	if !success {
		status = "error"
	}

	_, err := tao.hub.db.Exec(`
		UPDATE agent_registry
		SET status = ?, last_seen = ?, updated_at = ?
		WHERE id = ?
	`, status, time.Now().Unix(), time.Now().Unix(), agentID)

	if err != nil {
		logger.Warn("Failed to update agent status: %v", err)
	}
}

// GetAgentMemory retrieves memory for a specific agent
func (tao *TiAgentOrchestrator) GetAgentMemory(agentID string, memoryType string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	switch memoryType {
	case "tasks":
		rows, err := tao.hub.db.Query(`
			SELECT task_id, task_type, task_description, task_status, priority, created_at, completed_at
			FROM task_memory
			WHERE assigned_agent = ?
			ORDER BY created_at DESC
			LIMIT 10
		`, agentID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve task memory: %w", err)
		}
		defer rows.Close()

		tasks := []map[string]interface{}{}
		for rows.Next() {
			var taskID, taskType, taskDesc, taskStatus string
			var priority int
			var createdAt int64
			var completedAt sql.NullInt64

			err := rows.Scan(&taskID, &taskType, &taskDesc, &taskStatus, &priority, &createdAt, &completedAt)
			if err != nil {
				continue
			}

			tasks = append(tasks, map[string]interface{}{
				"task_id":      taskID,
				"task_type":    taskType,
				"description":  taskDesc,
				"status":       taskStatus,
				"priority":     priority,
				"created_at":   createdAt,
				"completed_at": nullableInt64(completedAt),
			})
		}
		result["tasks"] = tasks

	case "decisions":
		rows, err := tao.hub.db.Query(`
			SELECT decision_id, decision_type, decision_context, decision_outcome, confidence, timestamp
			FROM decision_memory
			WHERE decision_maker = ?
			ORDER BY timestamp DESC
			LIMIT 10
		`, agentID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve decision memory: %w", err)
		}
		defer rows.Close()

		decisions := []map[string]interface{}{}
		for rows.Next() {
			var decisionID, decisionType, decisionContext, decisionOutcome string
			var confidence float64
			var timestamp int64

			err := rows.Scan(&decisionID, &decisionType, &decisionContext, &decisionOutcome, &confidence, &timestamp)
			if err != nil {
				continue
			}

			decisions = append(decisions, map[string]interface{}{
				"decision_id":   decisionID,
				"decision_type": decisionType,
				"context":       decisionContext,
				"outcome":       decisionOutcome,
				"confidence":    confidence,
				"timestamp":     timestamp,
			})
		}
		result["decisions"] = decisions

	case "preferences":
		rows, err := tao.hub.db.Query(`
			SELECT preference_key, preference_value, preference_type, category, valid_from, valid_until
			FROM user_preferences
			WHERE user_id = ? AND (valid_until IS NULL OR valid_until > ?)
			ORDER BY category, preference_key
		`, agentID, time.Now().Unix())
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve user preferences: %w", err)
		}
		defer rows.Close()

		preferences := []map[string]interface{}{}
		for rows.Next() {
			var key, value, prefType, category string
			var validFrom int64
			var validUntil sql.NullInt64

			err := rows.Scan(&key, &value, &prefType, &category, &validFrom, &validUntil)
			if err != nil {
				continue
			}

			preferences = append(preferences, map[string]interface{}{
				"key":         key,
				"value":       value,
				"type":        prefType,
				"category":    category,
				"valid_from":  validFrom,
				"valid_until": nullableInt64(validUntil),
			})
		}
		result["preferences"] = preferences

	case "performance":
		rows, err := tao.hub.db.Query(`
			SELECT performance_metric, metric_value, metric_unit, timestamp, performance_rating
			FROM agent_performance_enhanced
			WHERE agent_id = ?
			ORDER BY timestamp DESC
			LIMIT 20
		`, agentID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve agent performance: %w", err)
		}
		defer rows.Close()

		metrics := []map[string]interface{}{}
		for rows.Next() {
			var metric, unit, rating string
			var value float64
			var timestamp int64

			err := rows.Scan(&metric, &value, &unit, &timestamp, &rating)
			if err != nil {
				continue
			}

			metrics = append(metrics, map[string]interface{}{
				"metric":    metric,
				"value":     value,
				"unit":      unit,
				"timestamp": timestamp,
				"rating":    rating,
			})
		}
		result["performance"] = metrics

	default:
		return nil, fmt.Errorf("unknown memory type: %s", memoryType)
	}

	return result, nil
}

// Helper functions

func marshalMap(m map[string]interface{}) string {
	if m == nil {
		return ""
	}
	data, _ := json.Marshal(m)
	return string(data)
}

func getPerformanceRating(success bool, durationMs int64) string {
	if !success {
		return "poor"
	}
	if durationMs < 100 {
		return "excellent"
	}
	if durationMs < 500 {
		return "good"
	}
	if durationMs < 1000 {
		return "fair"
	}
	return "slow"
}

func nullableInt64(value sql.NullInt64) interface{} {
	if !value.Valid {
		return nil
	}
	return value.Int64
}
