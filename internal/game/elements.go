package game

import (
	_ "embed"
	"encoding/json"
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
