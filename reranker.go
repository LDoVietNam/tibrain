package main

import (
	"math"
	"sort"
	"strings"
)

// SearchResult holds a document with multiple scores
type SearchResult struct {
	Doc          RAGDocument
	VectorScore  float64
	KeywordScore float64
	FinalScore   float64
}

// HybridReranker combines vector and keyword scores with optional cross-attention weighting
type HybridReranker struct {
	vectorWeight  float64
	keywordWeight float64
	threshold     float64
}

func NewHybridReranker() *HybridReranker {
	return &HybridReranker{
		vectorWeight:  0.6,
		keywordWeight: 0.4,
		threshold:     0.3,
	}
}

// Rerank merges vector and keyword results into a unified ranked list
func (hr *HybridReranker) Rerank(vectorResults, keywordResults []RAGDocument, query string) []SearchResult {
	merged := make(map[string]*SearchResult)

	for _, doc := range vectorResults {
		merged[doc.ID] = &SearchResult{
			Doc:          doc,
			VectorScore:  0.8, // default high score from vector search
			KeywordScore: 0.0,
		}
	}

	for _, doc := range keywordResults {
		if existing, ok := merged[doc.ID]; ok {
			existing.KeywordScore = hr.keywordOverlapScore(doc, query)
		} else {
			merged[doc.ID] = &SearchResult{
				Doc:          doc,
				VectorScore:  0.0,
				KeywordScore: hr.keywordOverlapScore(doc, query),
			}
		}
	}

	results := make([]SearchResult, 0, len(merged))
	for _, sr := range merged {
		sr.FinalScore = hr.vectorWeight*sr.VectorScore + hr.keywordWeight*sr.KeywordScore
		if sr.FinalScore >= hr.threshold {
			results = append(results, *sr)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].FinalScore > results[j].FinalScore
	})

	return results
}

func (hr *HybridReranker) keywordOverlapScore(doc RAGDocument, query string) float64 {
	queryTokens := strings.Fields(strings.ToLower(query))
	contentTokens := strings.Fields(strings.ToLower(doc.Content + " " + doc.Title + " " + doc.Tags))
	if len(queryTokens) == 0 {
		return 0
	}

	matchCount := 0
	for _, qt := range queryTokens {
		for _, ct := range contentTokens {
			if strings.Contains(ct, qt) || strings.Contains(qt, ct) {
				matchCount++
				break
			}
		}
	}

	return math.Min(1.0, float64(matchCount)/float64(len(queryTokens)))
}
