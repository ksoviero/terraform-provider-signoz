// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// CanonicalJSON parses JSON and returns a stable compact representation (sorted object keys).
// Use for comparing Terraform config with API payloads to avoid whitespace drift.
func CanonicalJSON(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		return "", err
	}
	// Encoder.Encode adds a trailing newline; trim for string attributes.
	out := bytes.TrimSpace(buf.Bytes())
	return string(out), nil
}
