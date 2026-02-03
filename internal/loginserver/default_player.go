package loginserver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type defaultPlayerFile struct {
	Player struct {
		Level     int   `json:"level"`
		Coins     int64 `json:"coins"`
		Gold      int64 `json:"gold"`
		Energy    int64 `json:"energy"`
		MapID     int   `json:"mapId"`
		PosX      int   `json:"posX"`
		PosY      int   `json:"posY"`
		TimeLimit int64 `json:"timeLimit"`
	} `json:"player"`
}

var (
	defaultPlayerOnce sync.Once
	defaultPlayerCfg  defaultPlayerFile
)

func loadDefaultPlayerConfig() defaultPlayerFile {
	defaultPlayerOnce.Do(func() {
		defaultPlayerCfg = defaultPlayerFile{}
		path := filepath.Join(".", "data", "config", "default-player.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			return
		}
		_ = json.Unmarshal(raw, &defaultPlayerCfg)
	})
	return defaultPlayerCfg
}
