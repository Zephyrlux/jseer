package game

import (
	"sort"
	"sync"
)

type mapOgresFile struct {
	Maps map[string]mapOgreMap `json:"maps"`
}

type mapOgreMap struct {
	Ogres []mapOgreEntry `json:"ogres"`
}

type mapOgreEntry struct {
	Slot  int `json:"slot"`
	PetID int `json:"petId"`
	Shiny int `json:"shiny"`
}

var (
	mapOgresOnce sync.Once
	mapOgresData map[uint32]map[int][2]uint32
)

func getMapOgreSlots(mapID uint32) map[int][2]uint32 {
	mapOgresOnce.Do(loadMapOgres)
	if slots, ok := mapOgresData[mapID]; ok {
		return slots
	}
	return map[int][2]uint32{}
}

func buildMapOgreBody(mapID uint32) []byte {
	slots := getMapOgreSlots(mapID)
	body := make([]byte, 9*8)
	offset := 0
	for i := 0; i < 9; i++ {
		if data, ok := slots[i]; ok {
			putU32(body, offset, data[0])
			offset += 4
			putU32(body, offset, data[1])
			offset += 4
		} else {
			putU32(body, offset, 0)
			offset += 4
			putU32(body, offset, 0)
			offset += 4
		}
	}
	return body
}

func putU32(buf []byte, offset int, value uint32) {
	buf[offset] = byte(value >> 24)
	buf[offset+1] = byte(value >> 16)
	buf[offset+2] = byte(value >> 8)
	buf[offset+3] = byte(value)
}

func loadMapOgres() {
	mapOgresData = defaultMapOgres()

	var cfg mapOgresFile
	if !readConfigJSON("map-ogres.json", &cfg) {
		return
	}

	out := make(map[uint32]map[int][2]uint32)
	for mapKey, entry := range cfg.Maps {
		mapID := parseUint32(mapKey)
		if mapID == 0 {
			continue
		}
		slots := normalizeOgreSlots(entry.Ogres)
		out[mapID] = slots
	}
	for mapID, slots := range out {
		mapOgresData[mapID] = slots
	}
}

func defaultMapOgres() map[uint32]map[int][2]uint32 {
	return map[uint32]map[int][2]uint32{
		8: {
			0: {10, 0},
			1: {58, 0},
		},
		515: {
			0: {10, 0},
			1: {58, 0},
		},
		301: {
			0: {1, 0},
			1: {4, 0},
			2: {7, 0},
			3: {10, 0},
		},
	}
}

func normalizeOgreSlots(ogres []mapOgreEntry) map[int][2]uint32 {
	slots := make(map[int][2]uint32)
	if len(ogres) == 0 {
		return slots
	}

	hasSmall := false
	for _, ogre := range ogres {
		if ogre.Slot >= 0 && ogre.Slot <= 8 {
			hasSmall = true
			break
		}
	}

	if hasSmall {
		for _, ogre := range ogres {
			if ogre.Slot < 0 || ogre.Slot > 8 {
				continue
			}
			slots[ogre.Slot] = [2]uint32{uint32(ogre.PetID), uint32(ogre.Shiny)}
		}
		if len(slots) >= 9 {
			return slots
		}
		extras := make([]mapOgreEntry, 0, len(ogres))
		for _, ogre := range ogres {
			if ogre.Slot > 8 {
				extras = append(extras, ogre)
			}
		}
		sort.Slice(extras, func(i, j int) bool { return extras[i].Slot < extras[j].Slot })
		next := 0
		for _, ogre := range extras {
			for next <= 8 {
				if _, ok := slots[next]; !ok {
					slots[next] = [2]uint32{uint32(ogre.PetID), uint32(ogre.Shiny)}
					next++
					break
				}
				next++
			}
			if next > 8 {
				break
			}
		}
		return slots
	}

	sort.Slice(ogres, func(i, j int) bool { return ogres[i].Slot < ogres[j].Slot })
	for i, ogre := range ogres {
		if i >= 9 {
			break
		}
		slots[i] = [2]uint32{uint32(ogre.PetID), uint32(ogre.Shiny)}
	}
	return slots
}
