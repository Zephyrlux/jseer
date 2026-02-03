package game

import "sync"

// MapBossEntry describes a map boss configuration for MAP_BOSS (2021).
type MapBossEntry struct {
	BossPetID int  `json:"bossPetId"`
	Level     int  `json:"level"`
	HasShield bool `json:"hasShield"`
}

type mapBossFile struct {
	Maps map[string]map[string]MapBossEntry `json:"maps"`
}

// mapBossConfig maps mapID -> region -> boss entry.
// region is the param2 in client boss requests; for MAP_BOSS list it is returned as-is.
var (
	mapBossOnce   sync.Once
	mapBossConfig map[int]map[uint32]MapBossEntry
)

func getMapBossEntries(mapID int) map[uint32]MapBossEntry {
	if mapID <= 0 {
		return nil
	}
	mapBossOnce.Do(loadMapBossConfig)
	return mapBossConfig[mapID]
}

func loadMapBossConfig() {
	mapBossConfig = defaultMapBossConfig()
	var cfg mapBossFile
	if !readConfigJSON("map-boss.json", &cfg) {
		return
	}
	if len(cfg.Maps) == 0 {
		return
	}
	out := make(map[int]map[uint32]MapBossEntry)
	for mapKey, regionMap := range cfg.Maps {
		mapID := int(parseUint32(mapKey))
		if mapID == 0 {
			continue
		}
		regions := make(map[uint32]MapBossEntry)
		for regionKey, entry := range regionMap {
			regionID := parseUint32(regionKey)
			if regionID == 0 {
				continue
			}
			regions[regionID] = entry
		}
		out[mapID] = regions
	}
	if len(out) > 0 {
		mapBossConfig = out
	}
}

func defaultMapBossConfig() map[int]map[uint32]MapBossEntry {
	return map[int]map[uint32]MapBossEntry{
		12:  {0: {47, 10, true}, 1: {83, 5, false}},
		22:  {0: {34, 25, false}},
		21:  {0: {34, 25, false}},
		17:  {0: {42, 35, false}},
		40:  {0: {50, 65, false}},
		27:  {0: {69, 45, false}},
		32:  {0: {70, 70, false}},
		106: {0: {88, 60, false}},
		49:  {0: {113, 55, false}},
		314: {0: {132, 70, false}},
		53:  {0: {187, 50, false}},
		60:  {0: {216, 80, false}},
		325: {0: {264, 70, false}},
		61:  {0: {421, 70, false}},
		348: {0: {274, 75, false}, 1: {391, 70, false}, 2: {216, 80, false}, 3: {413, 75, false}},
		59:  {0: {347, 70, false}},
		16:  {0: {393, 75, false}},
		10:  {0: {4150, 80, false}},
	}
}
