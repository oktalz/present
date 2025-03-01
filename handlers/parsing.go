package handlers

import "github.com/goccy/go-json"

type RequestPayload struct {
	Block *int     `json:"block"`
	Code  []string `json:"code"`
	Slide int      `json:"slide"`
}

func parseJSONData(jsonString string) (RequestPayload, error) {
	var payload RequestPayload
	err := json.Unmarshal([]byte(jsonString), &payload)
	return payload, err
}
