package opencode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a thin HTTP client for the OpenCode REST API (opencode serve).
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Client targeting the given base URL (e.g. "http://localhost:4096").
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Ping checks if the server is reachable.
func (c *Client) Ping() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}

// Session represents an OpenCode session.
type Session struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	ParentID string `json:"parentID,omitempty"`
}

// SessionList returns all sessions from the server.
func (c *Client) SessionList() ([]Session, error) {
	var result struct {
		Data []Session `json:"data"`
	}
	if err := c.get("/session", &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// SessionCreate creates a new session and returns its ID.
func (c *Client) SessionCreate(title string) (string, error) {
	body := map[string]string{"title": title}
	var result struct {
		Data *Session `json:"data"`
	}
	if err := c.post("/session", body, &result); err != nil {
		return "", err
	}
	if result.Data == nil {
		return "", fmt.Errorf("no session data returned")
	}
	return result.Data.ID, nil
}

// Agent represents an OpenCode agent.
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AppAgents returns the list of available agents.
func (c *Client) AppAgents() ([]Agent, error) {
	var result struct {
		Data []Agent `json:"data"`
	}
	if err := c.get("/app/agents", &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// get performs a GET request and JSON-decodes the response into out.
func (c *Client) get(path string, out any) error {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s: %d %s", path, resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// post performs a POST request with a JSON body and decodes the response into out.
func (c *Client) post(path string, body any, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Post(c.baseURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST %s: %d %s", path, resp.StatusCode, string(b))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
