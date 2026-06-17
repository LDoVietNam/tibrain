package management

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type comfyOutputFile struct {
	Filename  string
	Subfolder string
	Type      string
	Format    string
}

func submitComfyWorkflow(baseURL string, workflow map[string]interface{}) (string, error) {
	payload, _ := json.Marshal(map[string]interface{}{"prompt": workflow})
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/prompt", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("comfy submit failed (%d): %s", resp.StatusCode, string(body))
	}
	var out struct {
		PromptID string `json:"prompt_id"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if strings.TrimSpace(out.PromptID) == "" {
		return "", fmt.Errorf("comfy submit returned empty prompt_id")
	}
	return out.PromptID, nil
}

func pollComfyResult(baseURL, promptID string, timeout time.Duration) (map[string]interface{}, error) {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 30 * time.Second}
	for time.Now().Before(deadline) {
		historyURL := strings.TrimRight(baseURL, "/") + "/history/" + url.PathEscape(promptID)
		resp, err := client.Get(historyURL)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var history map[string]interface{}
				if json.Unmarshal(body, &history) == nil {
					if entry, ok := history[promptID].(map[string]interface{}); ok {
						return entry, nil
					}
					if len(history) == 1 {
						for _, v := range history {
							if entry, ok := v.(map[string]interface{}); ok {
								return entry, nil
							}
						}
					}
				}
			}
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("comfy poll timeout for prompt_id=%s", promptID)
}

func extractComfyOutputFiles(historyEntry map[string]interface{}) []comfyOutputFile {
	files := make([]comfyOutputFile, 0)
	outputs, _ := historyEntry["outputs"].(map[string]interface{})
	for _, nodeOutRaw := range outputs {
		nodeOut, ok := nodeOutRaw.(map[string]interface{})
		if !ok {
			continue
		}
		collect := func(key string, defaultFormat string) {
			arr, _ := nodeOut[key].([]interface{})
			for _, item := range arr {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				filename, _ := m["filename"].(string)
				subfolder, _ := m["subfolder"].(string)
				typeVal, _ := m["type"].(string)
				format := defaultFormat
				if f, ok := m["format"].(string); ok && strings.TrimSpace(f) != "" {
					format = f
				}
				if filename != "" {
					files = append(files, comfyOutputFile{Filename: filename, Subfolder: subfolder, Type: typeVal, Format: format})
				}
			}
		}
		collect("images", "png")
		collect("gifs", "webp")
		collect("audio", "wav")
		collect("videos", "mp4")
	}
	return files
}

func fetchComfyOutput(baseURL, filename, subfolder, fileType string) ([]byte, error) {
	q := url.Values{}
	q.Set("filename", filename)
	if strings.TrimSpace(subfolder) != "" {
		q.Set("subfolder", subfolder)
	}
	if strings.TrimSpace(fileType) != "" {
		q.Set("type", fileType)
	}
	viewURL := strings.TrimRight(baseURL, "/") + "/view?" + q.Encode()
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Get(viewURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("comfy view failed (%d): %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}
