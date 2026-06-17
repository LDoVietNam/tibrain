package routeagent

import (
	"fmt"
	"strings"
)

const (
	RouteSharedChat  = "sharedchat"
	RouteCLIProxyAPI = "cliproxyapi"
)

type Score struct {
	Route        string  `json:"route"`
	Available    bool    `json:"available"`
	AuthScore    float64 `json:"auth_score"`
	Success30m   float64 `json:"success_30m"`
	Success24h   float64 `json:"success_24h"`
	LatencyScore float64 `json:"latency_score"`
	TaskFitScore float64 `json:"task_fit_score"`
	PolicyScore  float64 `json:"policy_score"`
	FinalScore   float64 `json:"final_score"`
	Reason       string  `json:"reason"`
}

type Decision struct {
	Primary    string  `json:"primary"`
	Fallback   string  `json:"fallback,omitempty"`
	HasSession bool    `json:"has_session"`
	Profile    string  `json:"profile,omitempty"`
	Reason     string  `json:"reason"`
	Scores     []Score `json:"scores"`
}

type Engine struct{ store *Store }

func NewEngine(store *Store) *Engine { return &Engine{store: store} }

func (e *Engine) Decide(project, taskType, explicitProvider string, hasSession bool, profile string) (*Decision, error) {
	explicitProvider = normalizeRoute(explicitProvider)
	available := map[string]bool{RouteCLIProxyAPI: true, RouteSharedChat: hasSession}
	authScores := map[string]float64{RouteCLIProxyAPI: 1.0, RouteSharedChat: 0.15}
	if hasSession {
		authScores[RouteSharedChat] = 0.95
	}

	scores := make([]Score, 0, 2)
	for _, route := range []string{RouteSharedChat, RouteCLIProxyAPI} {
		routeStats, err := e.store.RouteStats(project, taskType, route)
		if err != nil {
			return nil, err
		}
		taskFit := taskFit(route, taskType)
		final := 0.30*authScores[route] + 0.20*routeStats.Success30m + 0.15*routeStats.Success24h + 0.15*routeStats.LatencyScore + 0.10*taskFit + 0.10*routeStats.PolicyScore
		reason := "historical score"
		if route == RouteSharedChat && hasSession {
			reason = "gfsessionid available"
		} else if route == RouteSharedChat && !hasSession {
			reason = "no gfsessionid"
		}
		scores = append(scores, Score{Route: route, Available: available[route], AuthScore: authScores[route], Success30m: routeStats.Success30m, Success24h: routeStats.Success24h, LatencyScore: routeStats.LatencyScore, TaskFitScore: taskFit, PolicyScore: routeStats.PolicyScore, FinalScore: final, Reason: reason})
	}

	d := &Decision{HasSession: hasSession, Profile: profile, Scores: scores}
	switch explicitProvider {
	case RouteSharedChat:
		d.Primary = RouteSharedChat
		d.Fallback = RouteCLIProxyAPI
		d.Reason = "explicit provider override: sharedchat"
	case RouteCLIProxyAPI, "proxy":
		d.Primary = RouteCLIProxyAPI
		d.Reason = "explicit provider override: cliproxyapi"
	default:
		if hasSession {
			d.Primary = RouteSharedChat
			d.Fallback = RouteCLIProxyAPI
			d.Reason = "gfsessionid detected, try sharedchat first then fallback to CLIProxyAPI"
		} else {
			d.Primary = RouteCLIProxyAPI
			d.Reason = "no gfsessionid detected, route directly to CLIProxyAPI"
		}
	}
	return d, nil
}

func ClassifyError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "401") || strings.Contains(msg, "403") || strings.Contains(msg, "session"):
		return "auth"
	case strings.Contains(msg, "429"):
		return "rate_limit"
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline"):
		return "timeout"
	case strings.Contains(msg, "502") || strings.Contains(msg, "503") || strings.Contains(msg, "504"):
		return "upstream"
	default:
		return "runtime"
	}
}

func normalizeRoute(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "", RouteSharedChat, RouteCLIProxyAPI:
		return v
	case "cliproxy", "proxy", "router":
		return RouteCLIProxyAPI
	default:
		return v
	}
}

func taskFit(route, taskType string) float64 {
	taskType = strings.ToLower(strings.TrimSpace(taskType))
	switch route {
	case RouteSharedChat:
		switch taskType {
		case "chat", "general", "plan":
			return 0.9
		case "code", "implement", "review":
			return 0.7
		default:
			return 0.75
		}
	case RouteCLIProxyAPI:
		switch taskType {
		case "code", "implement", "review":
			return 0.95
		case "chat", "general", "plan":
			return 0.8
		default:
			return 0.85
		}
	default:
		return 0.5
	}
}

func (d *Decision) Summary() string {
	if d == nil {
		return "no decision"
	}
	if d.Fallback != "" {
		return fmt.Sprintf("%s -> %s (%s)", d.Primary, d.Fallback, d.Reason)
	}
	return fmt.Sprintf("%s (%s)", d.Primary, d.Reason)
}
