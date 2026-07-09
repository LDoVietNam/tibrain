package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
)

type config struct {
	Port          int
	AccountID     string
	ZoneID        string
	APIToken      string
	DefaultHost   string
	DefaultTunnel string
	OriginURL     string
}

type app struct {
	client *cloudflare.Client
	cfg    config
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type envelope struct {
	Success    bool            `json:"success"`
	Errors     []apiError      `json:"errors"`
	Messages   []apiError      `json:"messages"`
	Result     json.RawMessage `json:"result"`
	ResultInfo struct {
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
}

type zoneItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type dnsRecordItem struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

type tunnelItem struct {
	ID           string `json:"id"`
	AccountTag   string `json:"account_tag"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	DeletedAt    any    `json:"deleted_at"`
	RemoteConfig bool   `json:"remote_config"`
}

type deployRequest struct {
	AccountID    string `json:"account_id"`
	ZoneID       string `json:"zone_id"`
	Hostname     string `json:"hostname"`
	TunnelName   string `json:"tunnel_name"`
	OriginURL    string `json:"origin_url"`
	ValidateOnly bool   `json:"validate_only"`
	Proxied      bool   `json:"proxied"`
}

type deployResponse struct {
	AccountID       string   `json:"account_id"`
	ZoneID          string   `json:"zone_id"`
	Hostname        string   `json:"hostname"`
	TunnelName      string   `json:"tunnel_name"`
	TunnelID        string   `json:"tunnel_id"`
	TunnelStatus    string   `json:"tunnel_status"`
	DNSRecordID     string   `json:"dns_record_id,omitempty"`
	DNSRecordTarget string   `json:"dns_record_target,omitempty"`
	ValidateOnly    bool     `json:"validate_only"`
	Steps           []string `json:"steps"`
	Warnings        []string `json:"warnings,omitempty"`
}

type preflightProbe struct {
	RequiredScope string `json:"required_scope"`
	Endpoint      string `json:"endpoint"`
	Method        string `json:"method"`
	ValidateOnly  bool   `json:"validate_only"`
	Status        string `json:"status"`
	HTTPStatus    int    `json:"http_status,omitempty"`
	Detail        string `json:"detail,omitempty"`
}

type preflightResponse struct {
	AccountID    string         `json:"account_id"`
	ZoneID       string         `json:"zone_id"`
	Hostname     string         `json:"hostname"`
	TunnelName   string         `json:"tunnel_name"`
	OriginURL    string         `json:"origin_url,omitempty"`
	TokenID      string         `json:"token_id,omitempty"`
	TokenStatus  string         `json:"token_status,omitempty"`
	TokenOK      bool           `json:"token_ok"`
	DNSWrite     preflightProbe `json:"dns_write"`
	TunnelWrite  preflightProbe `json:"tunnel_write"`
	Ready        bool           `json:"ready"`
	Missing      []string       `json:"missing,omitempty"`
	Warnings     []string       `json:"warnings,omitempty"`
}

func main() {
	cfg := loadConfig()
	if cfg.APIToken == "" {
		log.Fatal("CLOUDFLARE_API_TOKEN is required")
	}
	if cfg.AccountID == "" {
		log.Fatal("CLOUDFLARE_ACCOUNT_ID is required")
	}

	client := cloudflare.NewClient(
		option.WithAPIToken(cfg.APIToken),
		option.WithMaxRetries(2),
	)

	a := &app{
		client: client,
		cfg:    cfg,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/zones", a.handleZones)
	mux.HandleFunc("/zones/", a.handleZoneSubroutes)
	mux.HandleFunc("/tunnels", a.handleTunnels)
	mux.HandleFunc("/preflight/scope-check", a.handleScopeCheck)
	mux.HandleFunc("/deploy/fixed-tunnel", a.handleDeployFixedTunnel)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("backend listening on http://localhost:%d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func loadConfig() config {
	port := 1820
	if v := strings.TrimSpace(os.Getenv("BACKEND_PORT")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			port = n
		}
	}

	return config{
		Port:          port,
		AccountID:     firstNonEmpty(os.Getenv("CLOUDFLARE_ACCOUNT_ID")),
		ZoneID:        firstNonEmpty(os.Getenv("CLOUDFLARE_ZONE_ID")),
		APIToken:      firstNonEmpty(os.Getenv("CLOUDFLARE_API_TOKEN")),
		DefaultHost:   firstNonEmpty(os.Getenv("TIBRAIN_PUBLIC_HOSTNAME")),
		DefaultTunnel: firstNonEmpty(os.Getenv("TIBRAIN_TUNNEL_NAME")),
		OriginURL:     firstNonEmpty(os.Getenv("TIBRAIN_ORIGIN_URL")),
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"account_id": a.cfg.AccountID,
		"zone_id":    a.cfg.ZoneID,
		"hostname":   a.cfg.DefaultHost,
		"tunnel":     a.cfg.DefaultTunnel,
	})
}

func (a *app) handleZones(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resp envelope
	if err := a.cfGet(r.Context(), "/zones?per_page=100&order=name&direction=asc", nil, &resp); err != nil {
		writeAPIError(w, err)
		return
	}

	var zones []zoneItem
	if err := json.Unmarshal(resp.Result, &zones); err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":     resp.Success,
		"total_count": resp.ResultInfo.TotalCount,
		"zones":       zones,
	})
}

func (a *app) handleZoneSubroutes(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/zones/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 || parts[1] != "dns-records" {
		http.NotFound(w, r)
		return
	}
	zoneID := parts[0]
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resp envelope
	if err := a.cfGet(r.Context(), fmt.Sprintf("/zones/%s/dns_records?per_page=100&order=name&direction=asc", zoneID), nil, &resp); err != nil {
		writeAPIError(w, err)
		return
	}

	var records []dnsRecordItem
	if err := json.Unmarshal(resp.Result, &records); err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":     resp.Success,
		"total_count": resp.ResultInfo.TotalCount,
		"zone_id":     zoneID,
		"records":     records,
	})
}

func (a *app) handleTunnels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var resp envelope
	path := fmt.Sprintf("/accounts/%s/cfd_tunnel?per_page=100", a.cfg.AccountID)
	if err := a.cfGet(r.Context(), path, nil, &resp); err != nil {
		writeAPIError(w, err)
		return
	}

	var tunnels []tunnelItem
	if err := json.Unmarshal(resp.Result, &tunnels); err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":     resp.Success,
		"total_count": resp.ResultInfo.TotalCount,
		"account_id":  a.cfg.AccountID,
		"tunnels":     tunnels,
	})
}

func (a *app) handleScopeCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := deployRequest{
		AccountID:  firstNonEmpty(r.URL.Query().Get("account_id"), a.cfg.AccountID),
		ZoneID:     firstNonEmpty(r.URL.Query().Get("zone_id"), a.cfg.ZoneID),
		Hostname:   firstNonEmpty(r.URL.Query().Get("hostname"), a.cfg.DefaultHost),
		TunnelName: firstNonEmpty(r.URL.Query().Get("tunnel_name"), a.cfg.DefaultTunnel),
		OriginURL:  firstNonEmpty(r.URL.Query().Get("origin_url"), a.cfg.OriginURL),
	}

	resp, err := a.preflightScopeCheck(r.Context(), req)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (a *app) handleDeployFixedTunnel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req deployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, err)
		return
	}

	req.AccountID = firstNonEmpty(req.AccountID, a.cfg.AccountID)
	req.ZoneID = firstNonEmpty(req.ZoneID, a.cfg.ZoneID)
	req.Hostname = firstNonEmpty(req.Hostname, a.cfg.DefaultHost)
	req.TunnelName = firstNonEmpty(req.TunnelName, a.cfg.DefaultTunnel)
	req.OriginURL = firstNonEmpty(req.OriginURL, a.cfg.OriginURL)
	if v := strings.TrimSpace(r.URL.Query().Get("validate_only")); v != "" {
		req.ValidateOnly = v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
	}

	if req.AccountID == "" || req.ZoneID == "" || req.Hostname == "" || req.TunnelName == "" || req.OriginURL == "" {
		http.Error(w, "account_id, zone_id, hostname, tunnel_name, and origin_url are required", http.StatusBadRequest)
		return
	}

	preflight, err := a.preflightScopeCheck(r.Context(), req)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	if !preflight.Ready && !req.ValidateOnly {
		writeJSON(w, http.StatusForbidden, map[string]any{
			"success":   false,
			"error":     "preflight scope check failed",
			"preflight": preflight,
		})
		return
	}

	resp, err := a.deployFixedTunnel(r.Context(), req)
	if err != nil {
		writeAPIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func (a *app) deployFixedTunnel(ctx context.Context, req deployRequest) (*deployResponse, error) {
	out := &deployResponse{
		AccountID:    req.AccountID,
		ZoneID:       req.ZoneID,
		Hostname:     req.Hostname,
		TunnelName:   req.TunnelName,
		ValidateOnly: req.ValidateOnly,
		Steps:        []string{},
	}

	tunnel, err := a.findTunnelByName(ctx, req.AccountID, req.TunnelName)
	if err != nil {
		return nil, err
	}
	if tunnel == nil {
		if req.ValidateOnly {
			out.Warnings = append(out.Warnings, "validate_only cannot create a new tunnel; no changes were made")
			out.TunnelStatus = "not_found"
			return out, nil
		}
		tunnel, err = a.createTunnel(ctx, req.AccountID, req.TunnelName)
		if err != nil {
			return nil, err
		}
		out.Steps = append(out.Steps, "created tunnel")
	} else {
		out.Steps = append(out.Steps, "reused existing tunnel")
	}

	out.TunnelID = tunnel.ID
	out.TunnelStatus = tunnel.Status

	if err := a.updateTunnelConfig(ctx, req.AccountID, tunnel.ID, req.Hostname, req.OriginURL, req.ValidateOnly); err != nil {
		return nil, err
	}
	out.Steps = append(out.Steps, "updated tunnel config")

	if req.ValidateOnly {
		out.Warnings = append(out.Warnings, "validate_only was used; DNS record was not changed")
		return out, nil
	}

	target := tunnel.ID + ".cfargotunnel.com"
	if err := a.upsertCNAME(ctx, req.ZoneID, req.Hostname, target); err != nil {
		return nil, err
	}
	out.DNSRecordTarget = target
	out.Steps = append(out.Steps, "upserted DNS CNAME")

	return out, nil
}

func (a *app) preflightScopeCheck(ctx context.Context, req deployRequest) (*preflightResponse, error) {
	out := &preflightResponse{
		AccountID:  req.AccountID,
		ZoneID:     req.ZoneID,
		Hostname:   req.Hostname,
		TunnelName: req.TunnelName,
		OriginURL:  req.OriginURL,
	}

	tokenInfo, err := a.verifyToken(ctx)
	if err != nil {
		return nil, err
	}
	out.TokenOK = true
	out.TokenID = tokenInfo.ID
	out.TokenStatus = tokenInfo.Status

	out.DNSWrite = a.probeDNSWrite(ctx, req.ZoneID, req.Hostname)
	out.TunnelWrite = a.probeTunnelWrite(ctx, req.AccountID, req.TunnelName, req.OriginURL)

	out.Missing = append(out.Missing, missingScope(out.DNSWrite)...)
	out.Missing = append(out.Missing, missingScope(out.TunnelWrite)...)
	out.Ready = len(out.Missing) == 0
	if out.Ready {
		out.Warnings = append(out.Warnings, "preflight passed; fixed tunnel flow should be able to read the zone and perform tunnel/DNS writes")
	}

	return out, nil
}

type tokenVerifyInfo struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func (a *app) verifyToken(ctx context.Context) (*tokenVerifyInfo, error) {
	var resp tokenVerifyInfo
	if err := a.cfGet(ctx, "/user/tokens/verify", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *app) probeDNSWrite(ctx context.Context, zoneID, hostname string) preflightProbe {
	const fakeRecordID = "00000000000000000000000000000000"
	probe := preflightProbe{
		RequiredScope: "DNS Edit",
		Endpoint:      fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, fakeRecordID),
		Method:        http.MethodPut,
		ValidateOnly:  false,
	}
	payload := map[string]any{
		"type":    "CNAME",
		"name":    hostname,
		"content": "invalid.invalid",
		"proxied": false,
		"ttl":     1,
	}

	var resp envelope
	err := a.cfPut(ctx, probe.Endpoint, payload, &resp)
	classified, code, detail := classifyScopeProbe(err)
	probe.Status = classified
	probe.HTTPStatus = code
	probe.Detail = detail
	return probe
}

func (a *app) probeTunnelWrite(ctx context.Context, accountID, tunnelName, originURL string) preflightProbe {
	const fakeTunnelID = "00000000-0000-0000-0000-000000000000"
	probe := preflightProbe{
		RequiredScope: "Tunnel Edit",
		Endpoint:      fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/configurations?validate_only=true", accountID, fakeTunnelID),
		Method:        http.MethodPut,
		ValidateOnly:  true,
	}
	payload := map[string]any{
		"config": map[string]any{
			"ingress": []map[string]any{
				{
					"hostname": tunnelName + ".invalid",
					"service":  originURL,
				},
				{
					"service": "http_status:404",
				},
			},
		},
	}

	var resp envelope
	err := a.cfPut(ctx, probe.Endpoint, payload, &resp)
	classified, code, detail := classifyScopeProbe(err)
	probe.Status = classified
	probe.HTTPStatus = code
	probe.Detail = detail
	return probe
}

func classifyScopeProbe(err error) (status string, httpStatus int, detail string) {
	if err == nil {
		return "ok", http.StatusOK, "request unexpectedly succeeded"
	}

	var apiErr *cloudflare.Error
	if errors.As(err, &apiErr) {
		httpStatus = apiErr.StatusCode
		switch apiErr.StatusCode {
		case http.StatusForbidden:
			return "missing_scope", httpStatus, "Cloudflare returned 403; token lacks the required write permission"
		case http.StatusNotFound:
			return "ok", httpStatus, "write permission is present; the resource id was intentionally fake"
		default:
			return "error", httpStatus, err.Error()
		}
	}

	return "error", http.StatusBadRequest, err.Error()
}

func missingScope(probe preflightProbe) []string {
	if probe.Status == "missing_scope" {
		return []string{probe.RequiredScope}
	}
	return nil
}

func (a *app) findTunnelByName(ctx context.Context, accountID, tunnelName string) (*tunnelItem, error) {
	var resp envelope
	path := fmt.Sprintf("/accounts/%s/cfd_tunnel?per_page=100", accountID)
	if err := a.cfGet(ctx, path, nil, &resp); err != nil {
		return nil, err
	}

	var tunnels []tunnelItem
	if err := json.Unmarshal(resp.Result, &tunnels); err != nil {
		return nil, err
	}
	for _, tunnel := range tunnels {
		if tunnel.Name == tunnelName && tunnel.DeletedAt == nil {
			copy := tunnel
			return &copy, nil
		}
	}
	return nil, nil
}

func (a *app) createTunnel(ctx context.Context, accountID, tunnelName string) (*tunnelItem, error) {
	payload := map[string]any{
		"name":       tunnelName,
		"config_src": "cloudflare",
	}
	var resp envelope
	path := fmt.Sprintf("/accounts/%s/cfd_tunnel", accountID)
	if err := a.cfPost(ctx, path, payload, &resp); err != nil {
		return nil, err
	}

	var tunnel tunnelItem
	if err := json.Unmarshal(resp.Result, &tunnel); err != nil {
		return nil, err
	}
	return &tunnel, nil
}

func (a *app) updateTunnelConfig(ctx context.Context, accountID, tunnelID, hostname, originURL string, validateOnly bool) error {
	payload := map[string]any{
		"config": map[string]any{
			"ingress": []map[string]any{
				{
					"hostname": hostname,
					"service":  originURL,
				},
				{
					"service": "http_status:404",
				},
			},
		},
	}
	path := fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/configurations", accountID, tunnelID)
	if validateOnly {
		path += "?validate_only=true"
	}
	var resp envelope
	return a.cfPut(ctx, path, payload, &resp)
}

func (a *app) upsertCNAME(ctx context.Context, zoneID, hostname, target string) error {
	listPath := fmt.Sprintf("/zones/%s/dns_records?type=CNAME&name=%s", zoneID, url.QueryEscape(hostname))
	var resp envelope
	if err := a.cfGet(ctx, listPath, nil, &resp); err != nil {
		return err
	}

	var records []dnsRecordItem
	if err := json.Unmarshal(resp.Result, &records); err != nil {
		return err
	}
	if len(records) > 0 {
		rec := records[0]
		payload := map[string]any{
			"type":    "CNAME",
			"name":    hostname,
			"content": target,
			"proxied": true,
			"ttl":     1,
		}
		return a.cfPut(ctx, fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, rec.ID), payload, &envelope{})
	}

	payload := map[string]any{
		"type":    "CNAME",
		"name":    hostname,
		"content": target,
		"proxied": true,
		"ttl":     1,
	}
	return a.cfPost(ctx, fmt.Sprintf("/zones/%s/dns_records", zoneID), payload, &envelope{})
}

func (a *app) cfGet(ctx context.Context, path string, body any, dst any) error {
	return a.client.Get(ctx, path, body, dst)
}

func (a *app) cfPost(ctx context.Context, path string, body any, dst any) error {
	return a.client.Post(ctx, path, body, dst)
}

func (a *app) cfPut(ctx context.Context, path string, body any, dst any) error {
	return a.client.Put(ctx, path, body, dst)
}

func writeAPIError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	var apiErr *cloudflare.Error
	if errors.As(err, &apiErr) {
		status = apiErr.StatusCode
	}
	writeJSON(w, status, map[string]any{
		"success": false,
		"error":   err.Error(),
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
