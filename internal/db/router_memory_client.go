// Router Memory Client - Integration with server-memory
package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RouterMemoryClient handles communication with server-memory
type RouterMemoryClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Enabled    bool
}

// NewRouterMemoryClient creates a new router memory client
func NewRouterMemoryClient(baseURL string, enabled bool) *RouterMemoryClient {
	return &RouterMemoryClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Enabled: enabled,
	}
}

// CreateEntity creates a new entity in knowledge graph
func (c *RouterMemoryClient) CreateEntity(name string, entityType string, observations []string) error {
	if !c.Enabled {
		return nil
	}

	url := fmt.Sprintf("%s/tools/call", c.BaseURL)

	request := map[string]interface{}{
		"name": "create_entities",
		"arguments": map[string]interface{}{
			"entities": []map[string]interface{}{
				{
					"name":         name,
					"entityType":   entityType,
					"observations": observations,
				},
			},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// CreateRelation creates a relation between entities
func (c *RouterMemoryClient) CreateRelation(from string, to string, relationType string) error {
	if !c.Enabled {
		return nil
	}

	url := fmt.Sprintf("%s/tools/call", c.BaseURL)

	request := map[string]interface{}{
		"name": "create_relations",
		"arguments": map[string]interface{}{
			"relations": []map[string]interface{}{
				{
					"from":         from,
					"to":           to,
					"relationType": relationType,
				},
			},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SearchEntities searches for entities by name
func (c *RouterMemoryClient) SearchEntities(query string) ([]map[string]interface{}, error) {
	if !c.Enabled {
		return []map[string]interface{}{}, nil
	}

	url := fmt.Sprintf("%s/tools/call", c.BaseURL)

	request := map[string]interface{}{
		"name": "search_nodes",
		"arguments": map[string]interface{}{
			"query": query,
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Extract results from MCP response
	if results, ok := result["result"].([]interface{}); ok {
		var entities []map[string]interface{}
		for _, r := range results {
			if entity, ok := r.(map[string]interface{}); ok {
				entities = append(entities, entity)
			}
		}
		return entities, nil
	}

	return nil, fmt.Errorf("invalid response format")
}

// AddObservations adds observations to an entity
func (c *RouterMemoryClient) AddObservations(entityName string, observations []string) error {
	if !c.Enabled {
		return nil
	}

	url := fmt.Sprintf("%s/tools/call", c.BaseURL)

	request := map[string]interface{}{
		"name": "add_observations",
		"arguments": map[string]interface{}{
			"observations": []map[string]interface{}{
				{
					"entityName": entityName,
					"contents":   observations,
				},
			},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// OpenNodes retrieves detailed information about entities
func (c *RouterMemoryClient) OpenNodes(entityNames []string) ([]map[string]interface{}, error) {
	if !c.Enabled {
		return []map[string]interface{}{}, nil
	}

	url := fmt.Sprintf("%s/tools/call", c.BaseURL)

	request := map[string]interface{}{
		"name": "open_nodes",
		"arguments": map[string]interface{}{
			"names": entityNames,
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Extract results from MCP response
	if results, ok := result["result"].([]interface{}); ok {
		var entities []map[string]interface{}
		for _, r := range results {
			if entity, ok := r.(map[string]interface{}); ok {
				entities = append(entities, entity)
			}
		}
		return entities, nil
	}

	return nil, fmt.Errorf("invalid response format")
}

// HealthCheck checks if server-memory is accessible
func (c *RouterMemoryClient) HealthCheck() error {
	if !c.Enabled {
		return nil
	}

	// Try to search for a simple query to verify connection
	_, err := c.SearchEntities("test")
	return err
}

// ===== Pattern-specific helper functions =====

// SaveRoutingDecision saves a routing decision to memory
func (c *RouterMemoryClient) SaveRoutingDecision(decisionID string, model string, provider string, reason string, requestID string) error {
	entityName := fmt.Sprintf("routing_decision_%s", decisionID)

	observations := []string{
		fmt.Sprintf("Model: %s", model),
		fmt.Sprintf("Provider: %s", provider),
		fmt.Sprintf("Reason: %s", reason),
		fmt.Sprintf("Timestamp: %s", time.Now().Format(time.RFC3339)),
	}

	if err := c.CreateEntity(entityName, "routing_decision", observations); err != nil {
		return err
	}

	if requestID != "" {
		return c.CreateRelation(entityName, requestID, "for")
	}

	return c.CreateRelation(entityName, model, "uses")
}

// TrackRoutingOutcome tracks the outcome of a routing decision
func (c *RouterMemoryClient) TrackRoutingOutcome(decisionID string, status string, latency string, cost string, errorMessage string) error {
	entityName := fmt.Sprintf("routing_decision_%s", decisionID)

	observations := []string{
		fmt.Sprintf("Status: %s", status),
		fmt.Sprintf("Latency: %s", latency),
		fmt.Sprintf("Cost: %s", cost),
	}

	if errorMessage != "" {
		observations = append(observations, fmt.Sprintf("Error: %s", errorMessage))
	}

	if err := c.AddObservations(entityName, observations); err != nil {
		return err
	}

	// Create outcome entity
	outcomeName := fmt.Sprintf("outcome_%s_%s", decisionID, status)
	outcomeObs := []string{
		fmt.Sprintf("Decision: %s", decisionID),
		fmt.Sprintf("Status: %s", status),
		fmt.Sprintf("Timestamp: %s", time.Now().Format(time.RFC3339)),
	}

	if err := c.CreateEntity(outcomeName, "outcome", outcomeObs); err != nil {
		return err
	}

	return c.CreateRelation(entityName, outcomeName, "resulted_in")
}

// TrackModelPerformance tracks model performance metrics
func (c *RouterMemoryClient) TrackModelPerformance(modelName string, executionTime string, success bool, cost string) error {
	entityName := fmt.Sprintf("model_perf_%s", modelName)

	// Check if entity exists
	entities, err := c.SearchEntities(entityName)
	if err != nil {
		return err
	}

	if len(entities) == 0 {
		// Create new entity
		observations := []string{
			fmt.Sprintf("Model: %s", modelName),
			fmt.Sprintf("Total calls: 1"),
			fmt.Sprintf("Success rate: %.0f%%", map[bool]float64{true: 100, false: 0}[success]),
			fmt.Sprintf("Avg latency: %s", executionTime),
			fmt.Sprintf("Total cost: %s", cost),
			fmt.Sprintf("Last updated: %s", time.Now().Format(time.RFC3339)),
		}

		if err := c.CreateEntity(entityName, "model_performance", observations); err != nil {
			return err
		}

		return c.CreateRelation(entityName, modelName, "of")
	}

	// Update existing entity
	observations := []string{
		fmt.Sprintf("Last execution: %s", time.Now().Format(time.RFC3339)),
		fmt.Sprintf("Last execution time: %s", executionTime),
		fmt.Sprintf("Last success: %t", success),
		fmt.Sprintf("Last cost: %s", cost),
	}

	return c.AddObservations(entityName, observations)
}

// TrackToolUsage tracks tool execution metrics
func (c *RouterMemoryClient) TrackToolUsage(toolName string, executionTime string, success bool, model string) error {
	entityName := fmt.Sprintf("tool_usage_%s", toolName)

	// Check if entity exists
	entities, err := c.SearchEntities(entityName)
	if err != nil {
		return err
	}

	if len(entities) == 0 {
		// Create new entity
		observations := []string{
			fmt.Sprintf("Tool: %s", toolName),
			fmt.Sprintf("Total executions: 1"),
			fmt.Sprintf("Success rate: %.0f%%", map[bool]float64{true: 100, false: 0}[success]),
			fmt.Sprintf("Avg latency: %s", executionTime),
			fmt.Sprintf("Last updated: %s", time.Now().Format(time.RFC3339)),
		}

		if err := c.CreateEntity(entityName, "tool_usage", observations); err != nil {
			return err
		}

		return c.CreateRelation(entityName, toolName, "of")
	}

	// Update existing entity
	observations := []string{
		fmt.Sprintf("Last execution: %s", time.Now().Format(time.RFC3339)),
		fmt.Sprintf("Last execution time: %s", executionTime),
		fmt.Sprintf("Last success: %t", success),
		fmt.Sprintf("Last executed by: %s", model),
	}

	if err := c.AddObservations(entityName, observations); err != nil {
		return err
	}

	return c.CreateRelation(entityName, model, "executed_by")
}

// TrackSystemHealth tracks system health metrics
func (c *RouterMemoryClient) TrackSystemHealth(component string, status string, cpu string, memory string, errorRate string) error {
	entityName := fmt.Sprintf("system_health_%s_%s", component, time.Now().Format("20060102"))

	observations := []string{
		fmt.Sprintf("Component: %s", component),
		fmt.Sprintf("Status: %s", status),
		fmt.Sprintf("CPU: %s", cpu),
		fmt.Sprintf("Memory: %s", memory),
		fmt.Sprintf("Error rate: %s", errorRate),
		fmt.Sprintf("Timestamp: %s", time.Now().Format(time.RFC3339)),
	}

	if err := c.CreateEntity(entityName, "system_health", observations); err != nil {
		return err
	}

	return c.CreateRelation(entityName, component, "tracked_in")
}

// LearnUserPreference learns user preferences
func (c *RouterMemoryClient) LearnUserPreference(username string, preferenceType string, preferenceValue string, context string) error {
	entityName := fmt.Sprintf("user_%s_preference", username)

	observations := []string{
		fmt.Sprintf("Preference type: %s", preferenceType),
		fmt.Sprintf("Preference value: %s", preferenceValue),
		fmt.Sprintf("Context: %s", context),
		fmt.Sprintf("Learned at: %s", time.Now().Format(time.RFC3339)),
	}

	// Check if entity exists
	entities, err := c.SearchEntities(entityName)
	if err != nil {
		return err
	}

	if len(entities) == 0 {
		if err := c.CreateEntity(entityName, "user_preference", observations); err != nil {
			return err
		}
		return c.CreateRelation(entityName, username, "from")
	}

	return c.AddObservations(entityName, observations)
}

// GetUserPreferences retrieves user preferences
func (c *RouterMemoryClient) GetUserPreferences(username string) ([]string, error) {
	entityName := fmt.Sprintf("user_%s_preference", username)

	entities, err := c.OpenNodes([]string{entityName})
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return []string{}, nil
	}

	var preferences []string
	if obs, ok := entities[0]["observations"].([]interface{}); ok {
		for _, o := range obs {
			if obsStr, ok := o.(string); ok {
				preferences = append(preferences, obsStr)
			}
		}
	}

	return preferences, nil
}

// TrackThinkingProcess tracks thinking process
func (c *RouterMemoryClient) TrackThinkingProcess(thinkingID string, taskType string, complexity string, status string, latency string) error {
	entityName := fmt.Sprintf("thinking_process_%s", thinkingID)

	observations := []string{
		fmt.Sprintf("Task type: %s", taskType),
		fmt.Sprintf("Complexity: %s", complexity),
		fmt.Sprintf("Status: %s", status),
		fmt.Sprintf("Latency: %s", latency),
		fmt.Sprintf("Timestamp: %s", time.Now().Format(time.RFC3339)),
	}

	if err := c.CreateEntity(entityName, "thinking_process", observations); err != nil {
		return err
	}

	return c.CreateRelation(entityName, taskType, "for")
}
