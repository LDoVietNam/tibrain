package main

import (
	"context"
	"time"
)

type RouteType string

const (
	RouteGlobalRAG RouteType = "global_rag"
)

type RoutingDecision struct {
	Route RouteType `json:"route"`
}

type RAGResponse struct {
	Results      []*RAGResult `json:"results"`
	Query        string       `json:"query"`
	ResponseTime time.Duration `json:"response_time,omitempty"`
	Confidence   float64      `json:"confidence"`
	Tier         string       `json:"tier"`
	Metadata     *QueryMetadata `json:"metadata,omitempty"`
}

type RAGResult struct {
	DocumentID string   `json:"document_id"`
	FilePath   string   `json:"file_path"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Score      float64  `json:"score"`
	Relevance  float64  `json:"relevance"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags,omitempty"`
	Source     string   `json:"source"`
}

type QueryMetadata struct {
	TotalResults int           `json:"total_results"`
	SearchTime   time.Duration `json:"search_time"`
}

type StandardizedRAGResponse struct {
	Answer     string      `json:"answer,omitempty"`
	Results    []*RAGResult  `json:"results,omitempty"`
	Confidence float64       `json:"confidence"`
	Route      RouteType     `json:"route"`
}

type RetrievalRouter struct {
	hub *Hub
}

func NewRetrievalRouter(hub *Hub) *RetrievalRouter {
	return &RetrievalRouter{hub: hub}
}

func (r *RetrievalRouter) RouteQuery(ctx context.Context, query string) *RoutingDecision {
	return &RoutingDecision{Route: RouteGlobalRAG}
}

func (r *RetrievalRouter) ExecuteRoute(ctx context.Context, query string, decision *RoutingDecision, _ interface{}, _ interface{}) (*StandardizedRAGResponse, error) {
	resp, err := r.hub.Query(ctx, query, "default", 10, 0.5)
	if err != nil {
		return nil, err
	}
	return &StandardizedRAGResponse{
		Results:    resp.Results,
		Confidence: resp.Confidence,
		Route:      decision.Route,
	}, nil
}

func (r *RetrievalRouter) Ping() error {
	return nil
}