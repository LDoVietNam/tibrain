package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type RAGRuntimeTrace struct {
	Mode                string           `json:"mode"`
	CacheHit            bool             `json:"cache_hit"`
	Verified            bool             `json:"verified"`
	UsedSlowPath        bool             `json:"used_slow_path"`
	FastCandidates      int              `json:"fast_candidates"`
	SlowCandidates      int              `json:"slow_candidates"`
	SelectedDocuments   int              `json:"selected_documents"`
	ConfidenceScore     float64          `json:"confidence_score"`
	GroundednessScore   float64          `json:"groundedness_score"`
	CoverageScore       float64          `json:"coverage_score"`
	QualityScore        float64          `json:"quality_score"`
	DistilledCandidateID string          `json:"distilled_candidate_id,omitempty"`
	StageDurationsMs    map[string]int64 `json:"stage_durations_ms,omitempty"`
}

type BrainPatternCandidate struct {
	ID         string     `json:"id"`
	QueryID    string     `json:"query_id"`
	Title      string     `json:"title"`
	Summary    string     `json:"summary"`
	PatternBody string    `json:"pattern_body"`
	Evidence   string     `json:"evidence"`
	Score      float64    `json:"score"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	PromotedAt *time.Time `json:"promoted_at,omitempty"`
}

type RuntimeTraceRecord struct {
	ID                  string           `json:"id"`
	QueryID             string           `json:"query_id"`
	QueryText           string           `json:"query_text"`
	Mode                string           `json:"mode"`
	CacheHit            bool             `json:"cache_hit"`
	Verified            bool             `json:"verified"`
	UsedSlowPath        bool             `json:"used_slow_path"`
	ConfidenceScore     float64          `json:"confidence_score"`
	GroundednessScore   float64          `json:"groundedness_score"`
	CoverageScore       float64          `json:"coverage_score"`
	QualityScore        float64          `json:"quality_score"`
	FastCandidates      int              `json:"fast_candidates"`
	SlowCandidates      int              `json:"slow_candidates"`
	SelectedDocuments   int              `json:"selected_documents"`
	DistilledCandidateID string          `json:"distilled_candidate_id,omitempty"`
	StageDurationsMs    map[string]int64 `json:"stage_durations_ms,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
}

type retrievalCandidate struct {
	Doc           RAGDocument
	VectorScore   float64
	LexicalScore  float64
	FreshnessScore float64
	FinalScore    float64
}

type AdaptiveRetrievalRuntime struct {
	manager                   *RAGSystemManager
	slowPathThreshold         float64
	generativeAnswerThreshold float64
	answerTimeout             time.Duration
}

func NewAdaptiveRetrievalRuntime(manager *RAGSystemManager) *AdaptiveRetrievalRuntime {
	return &AdaptiveRetrievalRuntime{
		manager:                   manager,
		slowPathThreshold:         0.62,
		generativeAnswerThreshold: 0.45,
		answerTimeout:             3500 * time.Millisecond,
	}
}

func (r *RAGSystemManager) ProcessQueryWithContext(ctx context.Context, query RAGQuery) (*RAGQuery, error) {
	runtime := NewAdaptiveRetrievalRuntime(r)
	return runtime.Execute(ctx, query)
}

func (rt *AdaptiveRetrievalRuntime) Execute(ctx context.Context, query RAGQuery) (*RAGQuery, error) {
	startTime := time.Now()
	if query.ID == "" {
		query.ID = generateID()
	}

	trace := &RAGRuntimeTrace{
		Mode:             "fast_path",
		StageDurationsMs: make(map[string]int64),
	}

	if cached, ok := rt.manager.cache.GetQuery(query.Query); ok {
		cloned := cloneRAGQuery(cached)
		if cloned.Runtime == nil {
			cloned.Runtime = &RAGRuntimeTrace{}
		}
		cloned.Runtime.CacheHit = true
		cloned.Runtime.Mode = "cache_hit"
		cloned.Timestamp = time.Now()
		return cloned, nil
	}

	scopes := query.Scopes
	if len(scopes) == 0 && query.ContextSrc != nil {
		tags, _ := rt.manager.tagEngine.Process(*query.ContextSrc)
		scopes = tags
	}
	sqlFilter := BuildScopeQuery(scopes)

	fastStart := time.Now()
	fastCandidates := rt.fastPath(query.Query, sqlFilter)
	trace.StageDurationsMs["fast_path"] = time.Since(fastStart).Milliseconds()
	trace.FastCandidates = len(fastCandidates)

	selected := rt.selectTopDocuments(fastCandidates, rt.manager.config.TopK)
	confidence := rt.estimateConfidence(query.Query, selected)

	if rt.shouldUseSlowPath(query.Query, selected, confidence) {
		trace.Mode = "adaptive_slow_path"
		trace.UsedSlowPath = true
		slowStart := time.Now()
		slowCandidates := rt.slowPath(query.Query, sqlFilter)
		trace.StageDurationsMs["slow_path"] = time.Since(slowStart).Milliseconds()
		trace.SlowCandidates = len(slowCandidates)
		selected = rt.selectTopDocuments(rt.mergeCandidates(fastCandidates, slowCandidates), rt.manager.config.TopK)
		confidence = rt.estimateConfidence(query.Query, selected)
	}

	documents := make([]RAGDocument, 0, len(selected))
	docIDs := make([]string, 0, len(selected))
	for _, candidate := range selected {
		documents = append(documents, candidate.Doc)
		docIDs = append(docIDs, candidate.Doc.ID)
	}

	assemblyStart := time.Now()
	assembled := rt.manager.assembler.Assemble(documents, query.Query)
	trace.StageDurationsMs["assemble"] = time.Since(assemblyStart).Milliseconds()

	answerStart := time.Now()
	response := generateSimpleResponse(query.Query, documents)
	if rt.shouldUseGenerativeAnswer(confidence, documents, assembled.Text) {
		answerCtx, cancel := context.WithTimeout(ctx, rt.answerTimeout)
		defer cancel()
		if llmResp, err := rt.manager.llm.GenerateAnswer(answerCtx, query.Query, assembled.Text); err == nil && strings.TrimSpace(llmResp) != "" {
			if rt.calculateGroundedness(llmResp, documents) >= 0.18 {
				response = llmResp
			}
		}
	}
	trace.StageDurationsMs["answer"] = time.Since(answerStart).Milliseconds()

	trace.SelectedDocuments = len(documents)
	trace.ConfidenceScore = confidence
	trace.GroundednessScore = rt.calculateGroundedness(response, documents)
	trace.CoverageScore = rt.calculateCoverage(query.Query, documents)
	trace.QualityScore = rt.calculateQuality(trace.ConfidenceScore, trace.GroundednessScore, trace.CoverageScore)
	trace.Verified = len(documents) > 0 &&
		trace.CoverageScore >= 0.30 &&
		trace.GroundednessScore >= 0.18 &&
		trace.QualityScore >= 0.28

	result := &RAGQuery{
		ID:         query.ID,
		Query:      query.Query,
		Context:    query.Context,
		Documents:  docIDs,
		Response:   response,
		Confidence: confidence,
		Timestamp:  time.Now(),
		UserID:     query.UserID,
		SessionID:  query.SessionID,
		Metadata:   cloneStringMap(query.Metadata),
		Runtime:    trace,
		Scopes:     query.Scopes,
		ContextSrc: query.ContextSrc,
	}

	traceLogStart := time.Now()
	if trace.Verified {
		if candidateID, err := rt.manager.createPatternCandidate(query.ID, query.Query, response, documents, trace.QualityScore); err == nil {
			trace.DistilledCandidateID = candidateID
		}
	}
	rt.manager.persistQueryHistory(result, int(time.Since(startTime).Milliseconds()))
	rt.manager.logRuntimeTrace(query.Query, result)
	rt.manager.cache.SetQuery(query.Query, result)
	rt.manager.logAnalytics("adaptive_retrieval", trace.Mode, trace.QualityScore, map[string]string{
		"confidence":   fmt.Sprintf("%.3f", trace.ConfidenceScore),
		"groundedness": fmt.Sprintf("%.3f", trace.GroundednessScore),
		"coverage":     fmt.Sprintf("%.3f", trace.CoverageScore),
		"documents":    fmt.Sprintf("%d", len(documents)),
	})
	trace.StageDurationsMs["persist"] = time.Since(traceLogStart).Milliseconds()
	return result, nil
}

func (rt *AdaptiveRetrievalRuntime) fastPath(query, sqlFilter string) []retrievalCandidate {
	candidates := rt.manager.searchFTSScored(query, rt.manager.config.TopK*3, "", sqlFilter)
	if len(candidates) == 0 {
		candidates = rt.manager.searchKeywordScored(query, rt.manager.config.TopK*3, "", sqlFilter)
	}
	vectorCandidates := rt.manager.searchVectorScored(query, rt.manager.config.TopK*2, "", sqlFilter)
	return rt.mergeCandidates(candidates, vectorCandidates)
}

func (rt *AdaptiveRetrievalRuntime) slowPath(query, sqlFilter string) []retrievalCandidate {
	variants := []string{
		query,
		strings.Join(queryTokens(query), " "),
		strings.Join(queryTokens(query), " OR "),
	}

	merged := []retrievalCandidate{}
	for _, variant := range variants {
		if strings.TrimSpace(variant) == "" {
			continue
		}
		merged = rt.mergeCandidates(merged, rt.manager.searchKeywordScored(variant, rt.manager.config.TopK*4, "", sqlFilter))
	}
	return merged
}

func (rt *AdaptiveRetrievalRuntime) shouldUseSlowPath(query string, selected []retrievalCandidate, confidence float64) bool {
	if len(selected) == 0 {
		return true
	}
	if confidence < rt.slowPathThreshold {
		return true
	}
	return len(queryTokens(query)) >= 5 && confidence < 0.75
}

func (rt *AdaptiveRetrievalRuntime) shouldUseGenerativeAnswer(confidence float64, docs []RAGDocument, assembledText string) bool {
	return rt.manager != nil &&
		rt.manager.llm != nil &&
		len(docs) > 0 &&
		strings.TrimSpace(assembledText) != "" &&
		confidence >= rt.generativeAnswerThreshold
}

func (rt *AdaptiveRetrievalRuntime) mergeCandidates(base, extra []retrievalCandidate) []retrievalCandidate {
	if len(base) == 0 && len(extra) == 0 {
		return nil
	}

	merged := make(map[string]retrievalCandidate)
	for _, candidate := range append(base, extra...) {
		existing, ok := merged[candidate.Doc.ID]
		if !ok {
			candidate.FinalScore = rt.finalScore(candidate)
			merged[candidate.Doc.ID] = candidate
			continue
		}
		if candidate.VectorScore > existing.VectorScore {
			existing.VectorScore = candidate.VectorScore
		}
		if candidate.LexicalScore > existing.LexicalScore {
			existing.LexicalScore = candidate.LexicalScore
		}
		if candidate.FreshnessScore > existing.FreshnessScore {
			existing.FreshnessScore = candidate.FreshnessScore
		}
		existing.FinalScore = rt.finalScore(existing)
		merged[candidate.Doc.ID] = existing
	}

	results := make([]retrievalCandidate, 0, len(merged))
	for _, candidate := range merged {
		candidate.FinalScore = rt.finalScore(candidate)
		results = append(results, candidate)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].FinalScore == results[j].FinalScore {
			return results[i].Doc.UpdatedAt.After(results[j].Doc.UpdatedAt)
		}
		return results[i].FinalScore > results[j].FinalScore
	})
	return results
}

func (rt *AdaptiveRetrievalRuntime) selectTopDocuments(candidates []retrievalCandidate, limit int) []retrievalCandidate {
	if len(candidates) <= limit {
		return candidates
	}
	return candidates[:limit]
}

func (rt *AdaptiveRetrievalRuntime) finalScore(candidate retrievalCandidate) float64 {
	return math.Min(1.0, candidate.VectorScore*0.55+candidate.LexicalScore*0.35+candidate.FreshnessScore*0.10)
}

func (rt *AdaptiveRetrievalRuntime) estimateConfidence(query string, candidates []retrievalCandidate) float64 {
	if len(candidates) == 0 {
		return 0
	}
	top := candidates[0].FinalScore
	topLexical := candidates[0].LexicalScore
	gap := 0.0
	if len(candidates) > 1 {
		gap = math.Max(0, candidates[0].FinalScore-candidates[1].FinalScore)
	}
	tokenBoost := 0.0
	if len(queryTokens(query)) <= 4 {
		tokenBoost = 0.05
	}
	docBoost := math.Min(0.12, float64(len(candidates))*0.04)
	return math.Min(1.0, top*0.45+topLexical*0.35+gap*0.10+tokenBoost+docBoost)
}

func (rt *AdaptiveRetrievalRuntime) calculateGroundedness(answer string, docs []RAGDocument) float64 {
	if len(docs) == 0 || strings.TrimSpace(answer) == "" {
		return 0
	}
	answerTokens := uniqueNormalizedTokens(answer)
	if len(answerTokens) == 0 {
		return 0
	}
	sourceTokens := make(map[string]bool)
	for _, doc := range docs {
		for _, token := range uniqueNormalizedTokens(doc.Title + " " + doc.Content + " " + doc.Tags) {
			sourceTokens[token] = true
		}
	}
	matches := 0
	for _, token := range answerTokens {
		if sourceTokens[token] {
			matches++
		}
	}
	tokenOverlap := float64(matches) / float64(len(answerTokens))

	answerLower := strings.ToLower(answer)
	extractiveBoost := 0.0
	for i, doc := range docs {
		if i >= 3 {
			break
		}
		if title := strings.TrimSpace(strings.ToLower(doc.Title)); title != "" && strings.Contains(answerLower, title) {
			extractiveBoost += 0.12
		}
		snippet := strings.TrimSpace(strings.ToLower(truncateContent(doc.Content, 96)))
		if len(snippet) >= 24 && strings.Contains(answerLower, snippet[:24]) {
			extractiveBoost += 0.20
		}
	}

	return math.Min(1.0, tokenOverlap*0.75+extractiveBoost)
}

func (rt *AdaptiveRetrievalRuntime) calculateCoverage(query string, docs []RAGDocument) float64 {
	queryTokens := uniqueNormalizedTokens(query)
	if len(queryTokens) == 0 || len(docs) == 0 {
		return 0
	}
	docTokens := make(map[string]bool)
	for _, doc := range docs {
		for _, token := range uniqueNormalizedTokens(doc.Title + " " + doc.Content + " " + doc.Tags) {
			docTokens[token] = true
		}
	}
	matches := 0
	for _, token := range queryTokens {
		if docTokens[token] {
			matches++
		}
	}
	return math.Min(1.0, float64(matches)/float64(len(queryTokens)))
}

func (rt *AdaptiveRetrievalRuntime) calculateQuality(confidence, groundedness, coverage float64) float64 {
	return math.Min(1.0, confidence*0.25+groundedness*0.45+coverage*0.30)
}

func (r *RAGSystemManager) persistQueryHistory(result *RAGQuery, responseTime int) {
	if result == nil {
		return
	}
	timestamp := time.Now().Unix()
	r.hub.asyncWriter.Enqueue(`
		INSERT OR REPLACE INTO rag_query_history
		(id, query, context, documents, response, confidence, timestamp, user_id, session_id, metadata, query_hash, response_time, model_used, tokens_used)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		result.ID, result.Query, result.Context,
		strings.Join(result.Documents, ","), result.Response, result.Confidence,
		timestamp, result.UserID, result.SessionID,
		marshalMetadata(result.Metadata), generateQueryHash(result.Query),
		responseTime, "tibrain-llm", len(result.Response))
}

func (r *RAGSystemManager) logRuntimeTrace(queryText string, result *RAGQuery) {
	if result == nil || result.Runtime == nil {
		return
	}
	metadata, _ := json.Marshal(map[string]interface{}{
		"stage_durations_ms":    result.Runtime.StageDurationsMs,
		"distilled_candidate_id": result.Runtime.DistilledCandidateID,
	})
	r.hub.asyncWriter.Enqueue(`
		INSERT OR REPLACE INTO rag_runtime_traces
		(id, query_id, query_text, mode, cache_hit, verified, used_slow_path, confidence_score, groundedness_score, coverage_score, quality_score, fast_candidates, slow_candidates, selected_documents, distilled_candidate_id, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		generateID(), result.ID, queryText, result.Runtime.Mode, boolToInt(result.Runtime.CacheHit), boolToInt(result.Runtime.Verified),
		boolToInt(result.Runtime.UsedSlowPath), result.Runtime.ConfidenceScore, result.Runtime.GroundednessScore, result.Runtime.CoverageScore,
		result.Runtime.QualityScore, result.Runtime.FastCandidates, result.Runtime.SlowCandidates, result.Runtime.SelectedDocuments,
		result.Runtime.DistilledCandidateID, string(metadata), time.Now().Unix(),
	)
}

func (r *RAGSystemManager) createPatternCandidate(queryID, queryText, answer string, docs []RAGDocument, score float64) (string, error) {
	if len(docs) == 0 {
		return "", fmt.Errorf("no evidence documents")
	}
	titles := make([]string, 0, len(docs))
	evidenceIDs := make([]string, 0, len(docs))
	for i, doc := range docs {
		if i >= 3 {
			break
		}
		titles = append(titles, doc.Title)
		evidenceIDs = append(evidenceIDs, doc.ID)
	}

	title := fmt.Sprintf("Pattern from query: %s", truncateContent(queryText, 72))
	summary := fmt.Sprintf("Distilled from verified retrieval using sources: %s", strings.Join(titles, ", "))
	patternBody := strings.TrimSpace(answer)
	if patternBody == "" {
		patternBody = summary
	}

	id := generateID()
	now := time.Now().Unix()
	r.hub.asyncWriter.Enqueue(`
		INSERT OR REPLACE INTO brain_pattern_candidates
		(id, query_id, title, summary, pattern_body, evidence, score, status, created_at, updated_at, promoted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'candidate', ?, ?, NULL)
	`,
		id, queryID, title, summary, patternBody, strings.Join(evidenceIDs, ","), score, now, now,
	)
	return id, nil
}

func (r *RAGSystemManager) ListPatternCandidates(status string, limit int) ([]BrainPatternCandidate, error) {
	if limit <= 0 {
		limit = 20
	}
	query := `
		SELECT id, query_id, title, summary, pattern_body, evidence, score, status, created_at, updated_at, promoted_at
		FROM brain_pattern_candidates`
	args := []interface{}{}
	if status != "" {
		query += ` WHERE status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY updated_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.hub.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list pattern candidates: %w", err)
	}
	defer rows.Close()

	var candidates []BrainPatternCandidate
	for rows.Next() {
		var item BrainPatternCandidate
		var createdAt, updatedAt int64
		var promotedAt sql.NullInt64
		if err := rows.Scan(&item.ID, &item.QueryID, &item.Title, &item.Summary, &item.PatternBody, &item.Evidence, &item.Score, &item.Status, &createdAt, &updatedAt, &promotedAt); err != nil {
			return nil, fmt.Errorf("scan pattern candidate: %w", err)
		}
		item.CreatedAt = time.Unix(createdAt, 0)
		item.UpdatedAt = time.Unix(updatedAt, 0)
		if promotedAt.Valid {
			tm := time.Unix(promotedAt.Int64, 0)
			item.PromotedAt = &tm
		}
		candidates = append(candidates, item)
	}
	return candidates, nil
}

func (r *RAGSystemManager) PromotePatternCandidate(candidateID string) error {
	res, err := r.hub.db.Exec(`
		UPDATE brain_pattern_candidates
		SET status = 'promoted', updated_at = ?, promoted_at = ?
		WHERE id = ? AND status = 'candidate'
	`, time.Now().Unix(), time.Now().Unix(), candidateID)
	if err != nil {
		return fmt.Errorf("promote pattern candidate: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("pattern candidate not found or already promoted")
	}
	return nil
}

func (r *RAGSystemManager) searchFTSScored(query string, limit int, category string, sqlFilter string) []retrievalCandidate {
	matchQuery := buildFTSMatchQuery(query)
	if strings.TrimSpace(matchQuery) == "" {
		return nil
	}

	sqlQuery := `
		SELECT d.id, d.title, d.content, d.path, d.category, d.tags, d.created_at, d.updated_at, d.status, d.vector_id, d.metadata
		FROM rag_documents_fts
		JOIN rag_documents d ON d.rowid = rag_documents_fts.rowid
		WHERE rag_documents_fts MATCH ? AND d.status = 'active' ` + sqlFilter

	args := []interface{}{matchQuery}
	if category != "" {
		sqlQuery += ` AND d.category = ?`
		args = append(args, category)
	}
	sqlQuery += ` ORDER BY bm25(rag_documents_fts) LIMIT ?`
	args = append(args, limit)

	rows, err := r.hub.db.Query(sqlQuery, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var candidates []retrievalCandidate
	for rows.Next() {
		doc, err := scanRAGDocumentRow(rows)
		if err != nil {
			continue
		}
		candidates = append(candidates, retrievalCandidate{
			Doc:            doc,
			LexicalScore:   keywordOverlapScore(doc, query),
			FreshnessScore: freshnessScore(doc),
		})
	}
	return candidates
}

func (r *RAGSystemManager) searchKeywordScored(query string, limit int, category string, sqlFilter string) []retrievalCandidate {
	docs, err := r.searchKeyword(query, limit, category, sqlFilter)
	if err != nil {
		return nil
	}
	candidates := make([]retrievalCandidate, 0, len(docs))
	for _, doc := range docs {
		candidates = append(candidates, retrievalCandidate{
			Doc:            doc,
			LexicalScore:   keywordOverlapScore(doc, query),
			FreshnessScore: freshnessScore(doc),
		})
	}
	return candidates
}

func (r *RAGSystemManager) searchVectorScored(query string, limit int, category string, sqlFilter string) []retrievalCandidate {
	if !r.config.UseVectorSearch || r.vectorStore == nil || r.embedder == nil {
		return nil
	}
	if !r.hasIndexedVectors() {
		return nil
	}
	queryVector, err := r.embedQuery(query)
	if err != nil {
		return nil
	}
	results, err := r.vectorStore.SearchSimilar(queryVector, limit, r.config.Threshold, sqlFilter)
	if err != nil {
		return nil
	}
	candidates := make([]retrievalCandidate, 0, len(results))
	seen := make(map[string]bool)
	for _, result := range results {
		if seen[result.DocumentID] {
			continue
		}
		doc, err := r.GetDocument(result.DocumentID)
		if err != nil {
			continue
		}
		if category != "" && doc.Category != category {
			continue
		}
		seen[result.DocumentID] = true
		candidates = append(candidates, retrievalCandidate{
			Doc:            *doc,
			VectorScore:    result.Score,
			LexicalScore:   keywordOverlapScore(*doc, query),
			FreshnessScore: freshnessScore(*doc),
		})
	}
	return candidates
}

func (r *RAGSystemManager) embedQuery(query string) ([]float32, error) {
	if cached, ok := r.cache.GetEmbedding(query); ok {
		return cached, nil
	}
	embedding, err := r.embedder.GenerateEmbedding(query)
	if err != nil {
		return nil, err
	}
	r.cache.SetEmbedding(query, embedding)
	return embedding, nil
}

func (r *RAGSystemManager) hasIndexedVectors() bool {
	if r == nil || r.hub == nil || r.hub.db == nil {
		return false
	}
	var count int
	if err := r.hub.db.QueryRow(`SELECT COUNT(*) FROM rag_vector_index`).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func scanRAGDocumentRow(scanner interface {
	Scan(dest ...interface{}) error
}) (RAGDocument, error) {
	var doc RAGDocument
	var createdAt, updatedAt int64
	err := scanner.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.Path, &doc.Category, &doc.Tags, &createdAt, &updatedAt, &doc.Status, &doc.VectorID, &doc.Metadata)
	if err != nil {
		return RAGDocument{}, err
	}
	doc.CreatedAt = time.Unix(createdAt, 0)
	doc.UpdatedAt = time.Unix(updatedAt, 0)
	return doc, nil
}

func buildFTSMatchQuery(query string) string {
	tokens := queryTokens(query)
	if len(tokens) == 0 {
		return strings.TrimSpace(query)
	}
	parts := make([]string, 0, len(tokens))
	for _, token := range tokens {
		parts = append(parts, token+"*")
	}
	return strings.Join(parts, " OR ")
}

func keywordOverlapScore(doc RAGDocument, query string) float64 {
	queryTokens := uniqueNormalizedTokens(query)
	if len(queryTokens) == 0 {
		return 0
	}
	contentTokens := make(map[string]bool)
	for _, token := range uniqueNormalizedTokens(doc.Title + " " + doc.Content + " " + doc.Tags) {
		contentTokens[token] = true
	}
	matches := 0
	for _, token := range queryTokens {
		if contentTokens[token] {
			matches++
		}
	}
	return math.Min(1.0, float64(matches)/float64(len(queryTokens)))
}

func freshnessScore(doc RAGDocument) float64 {
	updatedAt := doc.UpdatedAt
	if updatedAt.IsZero() {
		return 0
	}
	ageHours := time.Since(updatedAt).Hours()
	if ageHours <= 0 {
		return 1
	}
	return math.Max(0, math.Min(1, 1-(ageHours/(24*30))))
}

func uniqueNormalizedTokens(text string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, token := range queryTokens(text) {
		if seen[token] {
			continue
		}
		seen[token] = true
		result = append(result, token)
	}
	return result
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneRAGQuery(input *RAGQuery) *RAGQuery {
	if input == nil {
		return nil
	}
	clone := *input
	clone.Documents = append([]string(nil), input.Documents...)
	clone.Metadata = cloneStringMap(input.Metadata)
	if input.Runtime != nil {
		trace := *input.Runtime
		if input.Runtime.StageDurationsMs != nil {
			trace.StageDurationsMs = make(map[string]int64, len(input.Runtime.StageDurationsMs))
			for key, value := range input.Runtime.StageDurationsMs {
				trace.StageDurationsMs[key] = value
			}
		}
		clone.Runtime = &trace
	}
	return &clone
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
