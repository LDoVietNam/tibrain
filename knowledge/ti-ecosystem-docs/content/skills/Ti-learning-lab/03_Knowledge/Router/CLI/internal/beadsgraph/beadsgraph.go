package beadsgraph

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ti/cli/internal/ticore"
)

type Status string

type Graph struct {
	ID        string    `json:"id"`
	PlanID    string    `json:"plan_id"`
	Project   string    `json:"project"`
	Goal      string    `json:"goal"`
	Summary   string    `json:"summary,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Epics     []Epic    `json:"epics"`
	Issues    []Issue   `json:"issues"`
}

type Epic struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Summary  string   `json:"summary,omitempty"`
	Priority int      `json:"priority"`
	Status   Status   `json:"status"`
	IssueIDs []string `json:"issue_ids,omitempty"`
}

type Issue struct {
	ID          string    `json:"id"`
	EpicID      string    `json:"epic_id"`
	StepID      string    `json:"step_id,omitempty"`
	Title       string    `json:"title"`
	Objective   string    `json:"objective,omitempty"`
	Files       []string  `json:"files,omitempty"`
	Commands    []string  `json:"commands,omitempty"`
	Acceptance  []string  `json:"acceptance,omitempty"`
	DependsOn   []string  `json:"depends_on,omitempty"`
	Priority    int       `json:"priority"`
	Status      Status    `json:"status"`
	RouteHints  []string  `json:"route_hints,omitempty"`
	LastSyncAt  time.Time `json:"last_sync_at,omitempty"`
	SyncSummary string    `json:"sync_summary,omitempty"`
}

type ReviewReport struct {
	GraphID           string   `json:"graph_id"`
	PlanID            string   `json:"plan_id"`
	EpicCount         int      `json:"epic_count"`
	IssueCount        int      `json:"issue_count"`
	ReadyCount        int      `json:"ready_count"`
	CompletedCount    int      `json:"completed_count"`
	BlockedCount      int      `json:"blocked_count"`
	MissingAcceptance []string `json:"missing_acceptance,omitempty"`
	OrphanIssues      []string `json:"orphan_issues,omitempty"`
	Recommendations   []string `json:"recommendations,omitempty"`
	ReadyIssues       []Issue  `json:"ready_issues,omitempty"`
}

type SyncReport struct {
	GraphID         string   `json:"graph_id"`
	PlanID          string   `json:"plan_id"`
	UpdatedIssues   []string `json:"updated_issues,omitempty"`
	CompletedIssues []string `json:"completed_issues,omitempty"`
	BlockedIssues   []string `json:"blocked_issues,omitempty"`
	ReadyIssues     []string `json:"ready_issues,omitempty"`
	GraphStatus     string   `json:"graph_status"`
	PlanState       string   `json:"plan_state"`
}

type Store struct{ path string }

func NewStore(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("empty beads graph store path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return &Store{path: path}, nil
}

func (s *Store) Save(g *Graph) error {
	if g == nil {
		return errors.New("nil graph")
	}
	g.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return err
	}
	archiveDir := filepath.Join(filepath.Dir(s.path), "graphs")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(archiveDir, g.PlanID+".json"), data, 0o644)
}

func (s *Store) LoadCurrent() (*Graph, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var g Graph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *Store) LoadByPlan(planID string) (*Graph, error) {
	data, err := os.ReadFile(filepath.Join(filepath.Dir(s.path), "graphs", planID+".json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var g Graph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func FromApprovedPlan(plan *ticore.ApprovedPlan) *Graph {
	if plan == nil {
		return nil
	}
	now := time.Now()
	epicID := fmt.Sprintf("epic-%s", plan.ID)
	g := &Graph{ID: fmt.Sprintf("graph-%s", plan.ID), PlanID: plan.ID, Project: plan.Project, Goal: plan.Goal, Summary: plan.Summary, CreatedAt: now, UpdatedAt: now}
	epic := Epic{ID: epicID, Title: firstNonEmpty(plan.Summary, plan.Goal, "Ti execution plan"), Summary: plan.Goal, Priority: 1, Status: Status("open")}
	issues := make([]Issue, 0, len(plan.Steps))
	var prevID string
	for i, step := range plan.Steps {
		issueID := fmt.Sprintf("issue-%s-%02d", plan.ID, i+1)
		issue := Issue{ID: issueID, EpicID: epicID, StepID: step.ID, Title: step.Title, Objective: step.Objective, Files: append([]string(nil), step.Files...), Commands: append([]string(nil), step.Commands...), Acceptance: append([]string(nil), step.ExitCriteria...), Priority: 1, Status: Status("open"), RouteHints: routeHints(plan)}
		if prevID != "" {
			issue.DependsOn = []string{prevID}
		}
		if len(issue.DependsOn) == 0 {
			issue.Status = Status("ready")
		}
		prevID = issueID
		epic.IssueIDs = append(epic.IssueIDs, issueID)
		issues = append(issues, issue)
	}
	epic.Status = Status(graphStatus(&Graph{Issues: issues}))
	g.Epics = []Epic{epic}
	g.Issues = issues
	return g
}

func ReadyIssues(g *Graph) []Issue {
	if g == nil {
		return nil
	}
	resolved := map[string]bool{}
	for _, issue := range g.Issues {
		if issue.Status == Status("completed") || issue.Status == Status("closed") {
			resolved[issue.ID] = true
		}
	}
	ready := make([]Issue, 0)
	for _, issue := range g.Issues {
		if issue.Status == Status("completed") || issue.Status == Status("closed") || issue.Status == Status("blocked") || issue.Status == Status("in_progress") || issue.Status == Status("review") {
			continue
		}
		depsReady := true
		for _, dep := range issue.DependsOn {
			if !resolved[dep] {
				depsReady = false
				break
			}
		}
		if depsReady {
			issue.Status = Status("ready")
			ready = append(ready, issue)
		}
	}
	sort.Slice(ready, func(i, j int) bool { return ready[i].ID < ready[j].ID })
	return ready
}

func Review(g *Graph) ReviewReport {
	r := ReviewReport{}
	if g == nil {
		r.Recommendations = []string{"Create a Beads graph from an approved plan first."}
		return r
	}
	r.GraphID = g.ID
	r.PlanID = g.PlanID
	r.EpicCount = len(g.Epics)
	r.IssueCount = len(g.Issues)
	issueIDs := map[string]bool{}
	linked := map[string]bool{}
	for _, issue := range g.Issues {
		issueIDs[issue.ID] = true
		if len(issue.Acceptance) == 0 {
			r.MissingAcceptance = append(r.MissingAcceptance, issue.ID)
		}
		switch issue.Status {
		case Status("completed"), Status("closed"):
			r.CompletedCount++
		case Status("blocked"):
			r.BlockedCount++
		}
	}
	for _, epic := range g.Epics {
		for _, id := range epic.IssueIDs {
			linked[id] = true
		}
	}
	for _, issue := range g.Issues {
		if !linked[issue.ID] {
			r.OrphanIssues = append(r.OrphanIssues, issue.ID)
		}
	}
	r.ReadyIssues = ReadyIssues(g)
	r.ReadyCount = len(r.ReadyIssues)
	if r.IssueCount == 0 {
		r.Recommendations = append(r.Recommendations, "Graph has no issues; re-file from an approved plan.")
	}
	if len(r.MissingAcceptance) > 0 {
		r.Recommendations = append(r.Recommendations, "Add acceptance criteria to issues missing exit criteria from the approved plan.")
	}
	if len(r.OrphanIssues) > 0 {
		r.Recommendations = append(r.Recommendations, "Attach orphan issues to an epic or regenerate the graph.")
	}
	if r.ReadyCount == 0 && r.IssueCount > 0 {
		r.Recommendations = append(r.Recommendations, "No ready issues are available; review dependencies or unblock the first incomplete step.")
	}
	if r.BlockedCount > 0 {
		r.Recommendations = append(r.Recommendations, "Blocked issues exist; resolve blockers before parallel execution expands.")
	}
	if len(r.Recommendations) == 0 {
		r.Recommendations = append(r.Recommendations, "Graph is coherent and has ready work.")
	}
	return r
}

func SyncWithApprovedPlan(g *Graph, plan *ticore.ApprovedPlan) SyncReport {
	rep := SyncReport{}
	if g == nil || plan == nil {
		return rep
	}
	rep.GraphID = g.ID
	rep.PlanID = g.PlanID
	rep.PlanState = string(plan.State)
	stepMap := map[string]ticore.ApprovedStep{}
	for _, s := range plan.Steps {
		stepMap[s.ID] = s
	}
	completed := map[string]bool{}
	for i := range g.Issues {
		issue := &g.Issues[i]
		step, ok := stepMap[issue.StepID]
		if !ok {
			issue.SyncSummary = "step missing from approved plan"
			continue
		}
		prev := issue.Status
		switch step.Status {
		case ticore.StepCompleted, ticore.StepSkipped:
			issue.Status = Status("completed")
			completed[issue.ID] = true
			rep.CompletedIssues = append(rep.CompletedIssues, issue.ID)
		case ticore.StepVerified:
			issue.Status = Status("review")
		case ticore.StepImplemented, ticore.StepInProgress:
			issue.Status = Status("in_progress")
		case ticore.StepBlocked:
			issue.Status = Status("blocked")
			rep.BlockedIssues = append(rep.BlockedIssues, issue.ID)
		default:
			issue.Status = Status("open")
		}
		issue.SyncSummary = strings.TrimSpace(step.LastNote)
		issue.LastSyncAt = time.Now()
		if issue.Status != prev {
			rep.UpdatedIssues = append(rep.UpdatedIssues, issue.ID)
		}
	}
	for i := range g.Issues {
		issue := &g.Issues[i]
		if issue.Status == Status("completed") {
			completed[issue.ID] = true
		}
	}
	for i := range g.Issues {
		issue := &g.Issues[i]
		if issue.Status == Status("completed") || issue.Status == Status("blocked") || issue.Status == Status("in_progress") || issue.Status == Status("review") {
			continue
		}
		depsReady := true
		for _, dep := range issue.DependsOn {
			if !completed[dep] {
				depsReady = false
				break
			}
		}
		if depsReady {
			issue.Status = Status("ready")
			rep.ReadyIssues = append(rep.ReadyIssues, issue.ID)
		} else {
			issue.Status = Status("open")
		}
	}
	rep.GraphStatus = graphStatus(g)
	g.UpdatedAt = time.Now()
	for i := range g.Epics {
		g.Epics[i].Status = Status(rep.GraphStatus)
	}
	return rep
}

func routeHints(plan *ticore.ApprovedPlan) []string {
	if plan == nil {
		return nil
	}
	hints := append([]string(nil), plan.RouteHints...)
	if len(hints) == 0 && plan.TaskType != "" {
		hints = append(hints, "task-type:"+plan.TaskType)
	}
	return hints
}

func graphStatus(g *Graph) string {
	if g == nil || len(g.Issues) == 0 {
		return "empty"
	}
	totalCompleted := 0
	totalBlocked := 0
	for _, issue := range g.Issues {
		switch issue.Status {
		case Status("completed"), Status("closed"):
			totalCompleted++
		case Status("blocked"):
			totalBlocked++
		}
	}
	switch {
	case totalCompleted == len(g.Issues):
		return "completed"
	case totalBlocked > 0:
		return "blocked"
	default:
		return "active"
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
