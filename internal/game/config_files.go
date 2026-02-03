package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

const configDir = "./data/config"

func readConfigJSON(name string, out any) bool {
	path := filepath.Join(configDir, name)
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return false
	}
	return true
}

func parseUint32(raw string) uint32 {
	if raw == "" {
		return 0
	}
	v, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(v)
}
