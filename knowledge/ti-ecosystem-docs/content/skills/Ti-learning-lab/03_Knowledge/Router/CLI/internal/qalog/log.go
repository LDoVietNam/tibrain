package qalog

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Entry struct {
	ID              string             `json:"id"`
	Timestamp       time.Time          `json:"timestamp"`
	Project         string             `json:"project,omitempty"`
	TaskType        string             `json:"task_type,omitempty"`
	Goal            string             `json:"goal,omitempty"`
	ApprovedPlanID  string             `json:"approved_plan_id,omitempty"`
	PromptHash      string             `json:"prompt_hash,omitempty"`
	PromptShape     string             `json:"prompt_shape,omitempty"`
	Route           string             `json:"route,omitempty"`
	Model           string             `json:"model,omitempty"`
	AnswerSummary   string             `json:"answer_summary,omitempty"`
	TotalTokens     int                `json:"total_tokens,omitempty"`
	Success         bool               `json:"success"`
	Quality         float64            `json:"quality,omitempty"`
	ScoreComponents map[string]float64 `json:"score_components,omitempty"`
	ErrorClass      string             `json:"error_class,omitempty"`
	Metadata        map[string]string  `json:"metadata,omitempty"`
	VerifyPassed    bool               `json:"verify_passed,omitempty"`
	VerifyCommands  int                `json:"verify_commands,omitempty"`
	VerifyFailures  int                `json:"verify_failures,omitempty"`
	VerifySummary   string             `json:"verify_summary,omitempty"`
}

type Stats struct {
	Total          int            `json:"total"`
	Successes      int            `json:"successes"`
	Failures       int            `json:"failures"`
	ByRoute        map[string]int `json:"by_route"`
	ByTaskType     map[string]int `json:"by_task_type"`
	ByPromptShape  map[string]int `json:"by_prompt_shape"`
	TotalTokens    int            `json:"total_tokens"`
	AverageQuality float64        `json:"average_quality"`
}

type Store struct{ path string }

func NewStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return &Store{path: path}, nil
}
func (s *Store) Close() error { return nil }

func (s *Store) Append(entry Entry) error {
	if entry.ID == "" {
		entry.ID = randomID()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}

func (s *Store) List(limit int) ([]Entry, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	out := []Entry{}
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var e Entry
		if json.Unmarshal([]byte(line), &e) == nil {
			out = append(out, e)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

func (s *Store) LatestExecutionForStep(planID, stepID string) (*Entry, error) {
	entries, err := s.List(0)
	if err != nil {
		return nil, err
	}
	planID = strings.TrimSpace(planID)
	stepID = strings.TrimSpace(stepID)
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if strings.TrimSpace(e.ApprovedPlanID) != planID {
			continue
		}
		if e.Metadata == nil || strings.TrimSpace(e.Metadata["phase"]) != "implement" {
			continue
		}
		if stepID != "" && strings.TrimSpace(e.Metadata["step_id"]) != stepID {
			continue
		}
		copyEntry := e
		return &copyEntry, nil
	}
	return nil, nil
}

func (s *Store) LatestByPhaseStep(planID, stepID, phase string) (*Entry, error) {
	entries, err := s.List(0)
	if err != nil {
		return nil, err
	}
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if strings.TrimSpace(e.ApprovedPlanID) != strings.TrimSpace(planID) {
			continue
		}
		if e.Metadata == nil || strings.TrimSpace(e.Metadata["phase"]) != strings.TrimSpace(phase) {
			continue
		}
		if stepID != "" && strings.TrimSpace(e.Metadata["step_id"]) != strings.TrimSpace(stepID) {
			continue
		}
		copyEntry := e
		return &copyEntry, nil
	}
	return nil, nil
}

func (s *Store) Stats() (*Stats, error) {
	entries, err := s.List(0)
	if err != nil {
		return nil, err
	}
	st := &Stats{ByRoute: map[string]int{}, ByTaskType: map[string]int{}, ByPromptShape: map[string]int{}}
	totalQuality := 0.0
	for _, e := range entries {
		st.Total++
		if e.Success {
			st.Successes++
		} else {
			st.Failures++
		}
		if e.Route != "" {
			st.ByRoute[e.Route]++
		}
		if e.TaskType != "" {
			st.ByTaskType[e.TaskType]++
		}
		if e.PromptShape != "" {
			st.ByPromptShape[e.PromptShape]++
		}
		st.TotalTokens += e.TotalTokens
		totalQuality += e.Quality
	}
	if st.Total > 0 {
		st.AverageQuality = totalQuality / float64(st.Total)
	}
	return st, nil
}

func randomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
