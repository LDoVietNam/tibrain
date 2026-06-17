package grcs

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// GRCS Generalized Reference Contrast Scoring
// Ported from Python GRCS implementation
type GRCS struct {
	PositiveCentroids [][]float64 `json:"positive_centroids"`
	NegativeCentroids [][]float64 `json:"negative_centroids"`
	AnchorSample      string      `json:"anchor_sample"`
	Alpha             float64     `json:"alpha"`
	K                 int         `json:"k"`
}

type Candidate struct {
	Content   string    `json:"content"`
	Embedding []float64 `json:"embedding,omitempty"`
	Score     float64   `json:"score"`
}

func NewGRCS() *GRCS {
	return &GRCS{
		Alpha: 0.1,
		K:     3,
	}
}

// CosineSimilarity calculates cosine similarity between two vectors
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ScoreCandidate scores a single candidate against GRCS model
func (g *GRCS) ScoreCandidate(candidate *Candidate) float64 {
	maxPos := 0.0
	for _, centroid := range g.PositiveCentroids {
		sim := CosineSimilarity(candidate.Embedding, centroid)
		if sim > maxPos {
			maxPos = sim
		}
	}

	maxNeg := 0.0
	for _, centroid := range g.NegativeCentroids {
		sim := CosineSimilarity(candidate.Embedding, centroid)
		if sim > maxNeg {
			maxNeg = sim
		}
	}

	return maxPos - (g.Alpha * maxNeg)
}

// RankCandidates scores and ranks multiple candidates
func (g *GRCS) RankCandidates(candidates []*Candidate) []*Candidate {
	for i := range candidates {
		candidates[i].Score = g.ScoreCandidate(candidates[i])
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	return candidates
}

// SelectBest returns highest scoring candidate
func (g *GRCS) SelectBest(candidates []*Candidate) (*Candidate, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidates provided")
	}

	ranked := g.RankCandidates(candidates)
	return ranked[0], nil
}

// GeneratePrimingPrompt creates system prompt with anchor reference
func (g *GRCS) GeneratePrimingPrompt(taskType string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are an expert %s. ", taskType))
	sb.WriteString("The following high-quality reference is the industry standard. ")
	sb.WriteString("Match this quality level in your output:\n\n")
	sb.WriteString("```\n")
	sb.WriteString(g.AnchorSample)
	sb.WriteString("\n```\n\n")
	sb.WriteString("Generate exactly 3 distinct, high-quality completions for the request.")

	return sb.String()
}

// Marshal serializes GRCS model to JSON
func (g *GRCS) Marshal() ([]byte, error) {
	return json.MarshalIndent(g, "", "  ")
}

// Unmarshal loads GRCS model from JSON
func Unmarshal(data []byte) (*GRCS, error) {
	var grcs GRCS
	if err := json.Unmarshal(data, &grcs); err != nil {
		return nil, err
	}
	return &grcs, nil
}
