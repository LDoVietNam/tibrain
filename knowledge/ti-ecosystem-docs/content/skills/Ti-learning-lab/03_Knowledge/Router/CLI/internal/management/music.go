package management

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// MusicGenerationRequest matches the OpenAI music generations request format.
type MusicGenerationRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Duration       int    `json:"duration,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
}

type musicProvider struct {
	ID      string
	BaseURL string
	Format  string
}

var musicProviders = map[string]musicProvider{
	"comfyui": {ID: "comfyui", BaseURL: "http://localhost:8188", Format: "comfyui"},
}

// handleMusicGenerations handles POST /v1/music/generations.
// Currently supports proxying to OpenAI-compatible providers.
// Full ComfyUI audio workflow support requires local node setup (ported from TS).
func (svc *Service) handleMusicGenerations(w http.ResponseWriter, r *http.Request) {
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

	var req MusicGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "parse request: "+err.Error(), "invalid_request_error")
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		writeOpenAIError(w, http.StatusBadRequest, "missing prompt", "invalid_request_error")
		return
	}

	providerID, modelID := parseLocalProviderModel(req.Model, "comfyui", "stable-audio-open")
	cfg, ok := musicProviders[providerID]
	if !ok {
		writeOpenAIError(w, http.StatusBadRequest, "unknown music provider: "+providerID, "invalid_request_error")
		return
	}
	if cfg.Format != "comfyui" {
		writeOpenAIError(w, http.StatusBadRequest, "unsupported music format: "+cfg.Format, "invalid_request_error")
		return
	}

	out, err := runComfyMusic(cfg.BaseURL, modelID, req)
	if err != nil {
		writeOpenAIError(w, http.StatusBadGateway, "music provider error: "+err.Error(), "server_error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"created": time.Now().Unix(),
		"data":    out,
	})
}

func runComfyMusic(baseURL, model string, req MusicGenerationRequest) ([]map[string]string, error) {
	duration := req.Duration
	if duration <= 0 {
		duration = 10
	}
	workflow := map[string]interface{}{
		"1": map[string]interface{}{"class_type": "CheckpointLoaderSimple", "inputs": map[string]interface{}{"ckpt_name": model}},
		"2": map[string]interface{}{"class_type": "CLIPTextEncode", "inputs": map[string]interface{}{"text": req.Prompt, "clip": []interface{}{"1", 1}}},
		"3": map[string]interface{}{"class_type": "CLIPTextEncode", "inputs": map[string]interface{}{"text": req.NegativePrompt, "clip": []interface{}{"1", 1}}},
		"4": map[string]interface{}{"class_type": "EmptyLatentAudio", "inputs": map[string]interface{}{"seconds": duration}},
		"5": map[string]interface{}{"class_type": "KSampler", "inputs": map[string]interface{}{"seed": time.Now().UnixNano(), "steps": 100, "cfg": 7, "sampler_name": "euler", "scheduler": "normal", "denoise": 1, "model": []interface{}{"1", 0}, "positive": []interface{}{"2", 0}, "negative": []interface{}{"3", 0}, "latent_image": []interface{}{"4", 0}}},
		"6": map[string]interface{}{"class_type": "VAEDecodeAudio", "inputs": map[string]interface{}{"samples": []interface{}{"5", 0}, "vae": []interface{}{"1", 2}}},
		"7": map[string]interface{}{"class_type": "SaveAudio", "inputs": map[string]interface{}{"filename_prefix": "tiroute_music", "audio": []interface{}{"6", 0}}},
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
		format := pickFormat(f.Format, "wav")
		out = append(out, map[string]string{"b64_json": base64.StdEncoding.EncodeToString(buf), "format": format})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no music output files")
	}
	return out, nil
}
