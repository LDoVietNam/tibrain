package main

import (
	"encoding/json"
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

// GetFeedbackStats returns aggregated feedback metrics
func (r *RAGSystemManager) GetFeedbackStats() map[string]interface{} {
	var total, positive, avgRating float64
	row := r.hub.db.QueryRow(`
		SELECT COUNT(*), COUNT(CASE WHEN rating >= 3 THEN 1 END), AVG(rating)
		FROM rag_feedback
	`)
	row.Scan(&total, &positive, &avgRating)

	return map[string]interface{}{
		"total_feedback": total,
		"positive_count": positive,
		"negative_count": total - positive,
		"average_rating": avgRating,
		"positive_rate":  positive / total * 100,
	}
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
