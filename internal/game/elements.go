package game

import (
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type elementTable struct {
	once sync.Once
	eff  [][]float64
}

var globalElementTable elementTable

//go:embed data/elements.json
var elementsJSON []byte

type elementData struct {
	MaxType       int         `json:"max_type"`
	Effectiveness [][]float64 `json:"effectiveness"`
}

func elementMultiplier(atkType int, defType int) float64 {
	globalElementTable.once.Do(loadElementTable)
	if atkType <= 0 || defType <= 0 {
		return 1
	}
	if atkType >= len(globalElementTable.eff) || defType >= len(globalElementTable.eff[atkType]) {
		return 1
	}
	return globalElementTable.eff[atkType][defType]
}

func loadElementTable() {
	if eff := loadElementTableFromConfig(); eff != nil {
		globalElementTable.eff = eff
		return
	}
	var data elementData
	if err := json.Unmarshal(elementsJSON, &data); err == nil && len(data.Effectiveness) > 0 {
		globalElementTable.eff = data.Effectiveness
		return
	}

	const maxType = 26
	eff := make([][]float64, maxType+1)
	for i := 0; i <= maxType; i++ {
		eff[i] = make([]float64, maxType+1)
		for j := 0; j <= maxType; j++ {
			eff[i][j] = 1
		}
	}
	globalElementTable.eff = eff
}

type elementConfigFile struct {
	Types         []elementTypeEntry            `json:"types"`
	Effectiveness map[string]map[string]float64 `json:"effectiveness"`
}

type elementTypeEntry struct {
	ID int `json:"id"`
}

func loadElementTableFromConfig() [][]float64 {
	path := filepath.Join(configDir, "elements.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cfg elementConfigFile
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil
	}
	maxType := 0
	for _, t := range cfg.Types {
		if t.ID > maxType {
			maxType = t.ID
		}
	}
	for atkKey, defs := range cfg.Effectiveness {
		if id := parseInt(atkKey); id > maxType {
			maxType = id
		}
		for defKey := range defs {
			if id := parseInt(defKey); id > maxType {
				maxType = id
			}
		}
	}
	if maxType == 0 {
		return nil
	}
	eff := make([][]float64, maxType+1)
	for i := 0; i <= maxType; i++ {
		eff[i] = make([]float64, maxType+1)
		for j := 0; j <= maxType; j++ {
			eff[i][j] = 1
		}
	}
	for atkKey, defs := range cfg.Effectiveness {
		atkID := parseInt(atkKey)
		if atkID <= 0 || atkID > maxType {
			continue
		}
		for defKey, val := range defs {
			defID := parseInt(defKey)
			if defID <= 0 || defID > maxType {
				continue
			}
			eff[atkID][defID] = val
		}
	}
	return eff
}

func parseInt(raw string) int {
	if raw == "" {
		return 0
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return v
}
