package game

import (
	"encoding/json"
	"strconv"
)

func decodeTaskStatus(raw string) map[int]byte {
	out := make(map[int]byte)
	if raw == "" {
		return out
	}
	tmp := make(map[string]uint8)
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		return out
	}
	for k, v := range tmp {
		id, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		out[id] = byte(v)
	}
	return out
}

func encodeTaskStatus(m map[int]byte) string {
	if len(m) == 0 {
		return "{}"
	}
	tmp := make(map[string]uint8, len(m))
	for k, v := range m {
		tmp[strconv.Itoa(k)] = uint8(v)
	}
	data, err := json.Marshal(tmp)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func decodeTaskBufs(raw string) map[int]map[int]uint32 {
	out := make(map[int]map[int]uint32)
	if raw == "" {
		return out
	}
	tmp := make(map[string]map[string]uint32)
	if err := json.Unmarshal([]byte(raw), &tmp); err != nil {
		return out
	}
	for key, buf := range tmp {
		taskID, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		if out[taskID] == nil {
			out[taskID] = make(map[int]uint32)
		}
		for idxKey, val := range buf {
			idx, err := strconv.Atoi(idxKey)
			if err != nil {
				continue
			}
			out[taskID][idx] = val
		}
	}
	return out
}

func encodeTaskBufs(m map[int]map[int]uint32) string {
	if len(m) == 0 {
		return "{}"
	}
	tmp := make(map[string]map[string]uint32, len(m))
	for taskID, buf := range m {
		item := make(map[string]uint32, len(buf))
		for idx, val := range buf {
			item[strconv.Itoa(idx)] = val
		}
		tmp[strconv.Itoa(taskID)] = item
	}
	data, err := json.Marshal(tmp)
	if err != nil {
		return "{}"
	}
	return string(data)
}
