package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// RAGFeedback represents user feedback on a RAG response
type RAGFeedback struct {
	QueryID    string `json:"query_id"`
	Rating     int    `json:"rating"` // 1-5
	Correction string `json:"correction,omitempty"`
	Comment    string `json:"comment,omitempty"`
	Timestamp  int64  `json:"timestamp"`
	UserID     string `json:"user_id,omitempty"`
}

// SubmitFeedback stores feedback and triggers learning updates
func (r *RAGSystemManager) SubmitFeedback(fb RAGFeedback) error {
	timestamp := time.Now().Unix()
	if fb.Timestamp == 0 {
		fb.Timestamp = timestamp
	}

	_, err := r.hub.db.Exec(`
		INSERT INTO rag_feedback
		(id, query_id, rating, feedback, improvement_suggestions, timestamp, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, generateID(), fb.QueryID, fb.Rating, fb.Comment, fb.Correction, fb.Timestamp, fb.UserID)

	if err != nil {
		return fmt.Errorf("store feedback: %w", err)
	}

	// Update routing decision log success status if applicable
	r.updateRoutingFeedback(fb.QueryID, fb.Rating)

	logger.Info("RAG feedback submitted: query=%s rating=%d", fb.QueryID, fb.Rating)
	return nil
}

func (r *RAGSystemManager) updateRoutingFeedback(queryID string, rating int) {
	_, err := r.hub.db.Exec(`
		UPDATE routing_decision_log
		SET success = ?
		WHERE query LIKE ?
	`, rating >= 3, "%"+queryID+"%")
	if err != nil {
		logger.Warn("Failed to update routing feedback: %v", err)
	}
}

// GetFeedbackStats returns aggregated feedback metrics.
// Returns an error if the DB scan fails (was previously silently swallowed,
// returning a map of zeros that looked like real data).
func (r *RAGSystemManager) GetFeedbackStats() (map[string]interface{}, error) {
	var total, positive float64
	var avgRating sql.NullFloat64
	err := r.hub.db.QueryRow(`
		SELECT COUNT(*), COUNT(CASE WHEN rating >= 3 THEN 1 END), AVG(rating)
		FROM rag_feedback
	`).Scan(&total, &positive, &avgRating)
	if err != nil {
		return nil, fmt.Errorf("scan feedback stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_feedback": int(total),
		"positive_count": int(positive),
		"negative_count": int(total - positive),
	}
	if avgRating.Valid {
		stats["average_rating"] = avgRating.Float64
	} else {
		stats["average_rating"] = 0.0
	}
	// Guard against division-by-zero: 0 feedbacks → 0% rate (not NaN).
	if total > 0 {
		stats["positive_rate"] = positive / total * 100
	} else {
		stats["positive_rate"] = 0.0
	}
	return stats, nil
}

// feedbackHandler handles POST /api/rag/feedback
func (s *APIServer) feedbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var fb RAGFeedback
	if err := json.NewDecoder(r.Body).Decode(&fb); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	ragManager := s.getRAGManager()
	if err := ragManager.SubmitFeedback(fb); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "submitted"})
}

// ensure errors import is used (keep goimports happy if file evolves)
var _ = errors.New
