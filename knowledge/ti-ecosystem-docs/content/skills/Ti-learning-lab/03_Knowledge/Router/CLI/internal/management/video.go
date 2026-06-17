package management

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VideoGenerationRequest matches the OpenAI video generations request format.
type VideoGenerationRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Size   string `json:"size,omitempty"`
	Frames int    `json:"frames,omitempty"`
}

type videoProvider struct {
	ID      string
	BaseURL string
	Format  string
}

var videoProviders = map[string]videoProvider{
	"comfyui": {ID: "comfyui", BaseURL: "http://localhost:8188", Format: "comfyui"},
	"sdwebui": {ID: "sdwebui", BaseURL: "http://localhost:7860", Format: "sdwebui-video"},
}

// handleVideoGenerations handles POST /v1/videos/generations using local providers.
func (svc *Service) handleVideoGenerations(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		writeOpenAIError(w, http.StatusMethodNotAllowed, "method not allowed", "invalid_request_error")
		return
	}

	var req VideoGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "parse request: "+err.Error(), "invalid_request_error")
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		writeOpenAIError(w, http.StatusBadRequest, "missing prompt", "invalid_request_error")
		return
	}

	providerID, modelID := parseLocalProviderModel(req.Model, "comfyui", "animatediff")
	cfg, ok := videoProviders[providerID]
	if !ok {
		writeOpenAIError(w, http.StatusBadRequest, "unknown video provider: "+providerID, "invalid_request_error")
		return
	}

	var (
		result []map[string]string
		err    error
	)
	if cfg.Format == "comfyui" {
		result, err = runComfyVideo(cfg.BaseURL, modelID, req)
	} else {
		result, err = runSDWebUIVideo(cfg.BaseURL, req)
	}
	if err != nil {
		writeOpenAIError(w, http.StatusBadGateway, "video provider error: "+err.Error(), "server_error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"created": time.Now().Unix(),
		"data":    result,
	})
}

func runComfyVideo(baseURL, model string, req VideoGenerationRequest) ([]map[string]string, error) {
	width, height := parseSize(req.Size, 512, 512)
	frames := req.Frames
	if frames <= 0 {
		frames = 16
	}
	workflow := map[string]interface{}{
		"1": map[string]interface{}{"class_type": "CheckpointLoaderSimple", "inputs": map[string]interface{}{"ckpt_name": model}},
		"2": map[string]interface{}{"class_type": "CLIPTextEncode", "inputs": map[string]interface{}{"text": req.Prompt, "clip": []interface{}{"1", 1}}},
		"3": map[string]interface{}{"class_type": "CLIPTextEncode", "inputs": map[string]interface{}{"text": "", "clip": []interface{}{"1", 1}}},
		"4": map[string]interface{}{"class_type": "EmptyLatentImage", "inputs": map[string]interface{}{"width": width, "height": height, "batch_size": frames}},
		"5": map[string]interface{}{"class_type": "KSampler", "inputs": map[string]interface{}{"seed": time.Now().UnixNano(), "steps": 20, "cfg": 7, "sampler_name": "euler", "scheduler": "normal", "denoise": 1, "model": []interface{}{"1", 0}, "positive": []interface{}{"2", 0}, "negative": []interface{}{"3", 0}, "latent_image": []interface{}{"4", 0}}},
		"6": map[string]interface{}{"class_type": "VAEDecode", "inputs": map[string]interface{}{"samples": []interface{}{"5", 0}, "vae": []interface{}{"1", 2}}},
		"7": map[string]interface{}{"class_type": "SaveAnimatedWEBP", "inputs": map[string]interface{}{"filename_prefix": "tiroute_video", "fps": 8, "lossless": false, "quality": 80, "method": "default", "images": []interface{}{"6", 0}}},
	}
	promptID, err := submitComfyWorkflow(baseURL, workflow)
	if err != nil {
		return nil, err
	}
	history, err := pollComfyResult(baseURL, promptID, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	files := extractComfyOutputFiles(history)
	out := make([]map[string]string, 0, len(files))
	for _, f := range files {
		buf, err := fetchComfyOutput(baseURL, f.Filename, f.Subfolder, f.Type)
		if err != nil {
			continue
		}
		out = append(out, map[string]string{"b64_json": base64.StdEncoding.EncodeToString(buf), "format": pickFormat(f.Format, "webp")})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no video output files")
	}
	return out, nil
}

func runSDWebUIVideo(baseURL string, req VideoGenerationRequest) ([]map[string]string, error) {
	width, height := parseSize(req.Size, 512, 512)
	body := map[string]interface{}{
		"prompt":          req.Prompt,
		"negative_prompt": "",
		"width":           width,
		"height":          height,
		"steps":           20,
		"cfg_scale":       7,
		"frames":          maxInt(req.Frames, 16),
		"fps":             8,
	}
	resp, err := doJSONRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/animatediff/v1/generate", body, nil)
	if err != nil {
		return nil, err
	}
	videos := make([]map[string]string, 0)
	if v, ok := resp["video"].(string); ok && strings.TrimSpace(v) != "" {
		videos = append(videos, map[string]string{"b64_json": v, "format": "mp4"})
	}
	if imgs, ok := resp["images"].([]interface{}); ok {
		for _, it := range imgs {
			if s, ok := it.(string); ok {
				videos = append(videos, map[string]string{"b64_json": s, "format": "mp4"})
			}
		}
	}
	if len(videos) == 0 {
		return nil, fmt.Errorf("no sdwebui video output")
	}
	return videos, nil
}

func parseSize(size string, defW, defH int) (int, int) {
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return defW, defH
	}
	var w, h int
	if _, err := fmt.Sscanf(size, "%dx%d", &w, &h); err != nil || w <= 0 || h <= 0 {
		return defW, defH
	}
	return w, h
}

func maxInt(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}

func pickFormat(v, fallback string) string {
	if strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}

func doJSONRequest(method, endpoint string, payload map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := (&http.Client{Timeout: 120 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed (%d): %s", resp.StatusCode, string(respBody))
	}
	var out map[string]interface{}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func parseLocalProviderModel(rawModel, defProvider, defModel string) (string, string) {
	model := strings.TrimSpace(rawModel)
	if model == "" {
		return defProvider, defModel
	}
	if strings.Contains(model, "/") {
		parts := strings.SplitN(model, "/", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return defProvider, model
}
