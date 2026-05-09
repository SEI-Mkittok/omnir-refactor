package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func decodeJSONPatch(r *http.Request, dst any) (map[string]json.RawMessage, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return nil, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func patchFieldIsNull(raw map[string]json.RawMessage, field string) bool {
	value, ok := raw[field]
	return ok && bytes.Equal(bytes.TrimSpace(value), []byte("null"))
}
