// Package router - additional client options for Ti-specific metadata.
package router

import "strings"

// WithTiMeta attaches Ti-specific metadata headers to outgoing requests.
// Each map entry is sent as an X-Ti-<key> header. Useful for routing
// telemetry (project, task-type, route, prompt hash, etc.).
func WithTiMeta(meta map[string]string) ClientOption {
	return func(c *Client) {
		if c.tiMeta == nil {
			c.tiMeta = make(map[string]string, len(meta))
		}
		for k, v := range meta {
			// Normalise key: ti-<key>
			key := strings.ToLower(strings.TrimSpace(k))
			if key == "" {
				continue
			}
			c.tiMeta[key] = v
		}
	}
}
