package game

import (
	"encoding/json"
)

type fitmentPayload struct {
	ID     uint32 `json:"id"`
	X      uint32 `json:"x"`
	Y      uint32 `json:"y"`
	Dir    uint32 `json:"dir"`
	Status uint32 `json:"status"`
}

func decodeFitments(raw string) []Fitment {
	if raw == "" {
		return nil
	}
	var list []fitmentPayload
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil
	}
	out := make([]Fitment, 0, len(list))
	for _, f := range list {
		out = append(out, Fitment{
			ID:     f.ID,
			X:      f.X,
			Y:      f.Y,
			Dir:    f.Dir,
			Status: f.Status,
		})
	}
	return out
}

func encodeFitments(list []Fitment) string {
	if len(list) == 0 {
		return "[]"
	}
	payload := make([]fitmentPayload, 0, len(list))
	for _, f := range list {
		payload = append(payload, fitmentPayload{
			ID:     f.ID,
			X:      f.X,
			Y:      f.Y,
			Dir:    f.Dir,
			Status: f.Status,
		})
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "[]"
	}
	return string(data)
}
