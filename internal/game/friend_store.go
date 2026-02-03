package game

import (
	"encoding/json"
)

type friendEntry struct {
	UserID   uint32 `json:"userID"`
	TimePoke uint32 `json:"timePoke"`
}

func decodeFriends(raw string) []FriendInfo {
	if raw == "" {
		return nil
	}
	var entries []friendEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil
	}
	out := make([]FriendInfo, 0, len(entries))
	for _, e := range entries {
		out = append(out, FriendInfo{
			UserID:   e.UserID,
			TimePoke: e.TimePoke,
		})
	}
	return out
}

func encodeFriends(list []FriendInfo) string {
	if len(list) == 0 {
		return "[]"
	}
	entries := make([]friendEntry, 0, len(list))
	for _, f := range list {
		entries = append(entries, friendEntry{
			UserID:   f.UserID,
			TimePoke: f.TimePoke,
		})
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeBlacklist(raw string) []uint32 {
	if raw == "" {
		return nil
	}
	var list []uint32
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil
	}
	return list
}

func encodeBlacklist(list []uint32) string {
	if len(list) == 0 {
		return "[]"
	}
	data, err := json.Marshal(list)
	if err != nil {
		return "[]"
	}
	return string(data)
}
