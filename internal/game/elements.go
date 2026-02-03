package game

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type elementTable struct {
	once sync.Once
	eff  [][]float64
}

var globalElementTable elementTable

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
	const maxType = 26
	eff := make([][]float64, maxType+1)
	for i := 0; i <= maxType; i++ {
		eff[i] = make([]float64, maxType+1)
		for j := 0; j <= maxType; j++ {
			eff[i][j] = 1
		}
	}

	path := resolveElementsPath()
	if path != "" {
		parseElementTable(path, eff)
	}
	globalElementTable.eff = eff
}

func resolveElementsPath() string {
	if v := os.Getenv("JSEER_ELEMENTS_PATH"); v != "" {
		return v
	}
	if v := os.Getenv("JSEER_GAME_ROOT"); v != "" {
		p := filepath.Join(v, "seer_elements.lua")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	candidates := []string{
		filepath.Join("..", "Reseer-main", "luvit_version", "game", "seer_elements.lua"),
		filepath.Join("..", "Reseer-main", "game", "seer_elements.lua"),
		filepath.Join("Reseer-main", "luvit_version", "game", "seer_elements.lua"),
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

func parseElementTable(path string, eff [][]float64) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	nameToID := map[string]int{
		"GRASS":     1,
		"WATER":     2,
		"FIRE":      3,
		"FLYING":    4,
		"ELECTRIC":  5,
		"MACHINE":   6,
		"GROUND":    7,
		"NORMAL":    8,
		"ICE":       9,
		"PSYCHIC":   10,
		"FIGHTING":  11,
		"LIGHT":     12,
		"DARK":      13,
		"MYSTERY":   14,
		"DRAGON":    15,
		"HOLY":      16,
		"DIMENSION": 17,
		"ANCIENT":   18,
		"EVIL":      19,
		"NATURE":    20,
		"KING":      21,
		"CHAOS":     22,
		"DIVINE":    23,
		"CYCLE":     24,
		"BUG":       25,
		"VOID":      26,
	}

	re := regexp.MustCompile(`Elements\.EFFECTIVENESS\[E\.([A-Z_]+)\]\[E\.([A-Z_]+)\]\s*=\s*([0-9.]+)`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.Contains(line, "Elements.EFFECTIVENESS") {
			continue
		}
		m := re.FindStringSubmatch(line)
		if len(m) != 4 {
			continue
		}
		atkID := nameToID[m[1]]
		defID := nameToID[m[2]]
		if atkID == 0 || defID == 0 {
			continue
		}
		val, err := strconv.ParseFloat(m[3], 64)
		if err != nil {
			continue
		}
		if atkID < len(eff) && defID < len(eff[atkID]) {
			eff[atkID][defID] = val
		}
	}
}
