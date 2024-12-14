package handlers

import "github.com/goccy/go-json"

type RequestPayload struct {
	Slide int      `json:"slide"`
	Code  []string `json:"code"`
	Block *int     `json:"block"`
}

func parseJSONData(jsonString string) (RequestPayload, error) {
	var payload RequestPayload
	err := json.Unmarshal([]byte(jsonString), &payload)
	return payload, err
}
