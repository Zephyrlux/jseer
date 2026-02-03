package game

// MapBossEntry describes a map boss configuration for MAP_BOSS (2021).
type MapBossEntry struct {
	BossPetID int
	Level     int
	HasShield bool
}

// mapBossConfig maps mapID -> region -> boss entry.
// region is the param2 in client boss requests; for MAP_BOSS list it is returned as-is.
var mapBossConfig = map[int]map[uint32]MapBossEntry{
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

func getMapBossEntries(mapID int) map[uint32]MapBossEntry {
	if mapID <= 0 {
		return nil
	}
	return mapBossConfig[mapID]
}
