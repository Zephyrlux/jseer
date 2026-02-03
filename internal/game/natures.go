package game

import (
	"math/rand"
	"sync"
)

type naturesFile struct {
	Natures []natureEntry `json:"natures"`
}

type natureEntry struct {
	ID int `json:"id"`
}

var (
	natureOnce sync.Once
	natureIDs  []int
)

func loadNatures() {
	ids := make([]int, 0, 32)
	var cfg naturesFile
	if readConfigJSON("natures.json", &cfg) {
		for _, n := range cfg.Natures {
			if n.ID >= 0 {
				ids = append(ids, n.ID)
			}
		}
	}
	if len(ids) == 0 {
		for i := 0; i < 25; i++ {
			ids = append(ids, i)
		}
	}
	natureIDs = ids
}

func randNature() int {
	natureOnce.Do(loadNatures)
	if len(natureIDs) == 0 {
		return rand.Intn(25)
	}
	return natureIDs[rand.Intn(len(natureIDs))]
}
