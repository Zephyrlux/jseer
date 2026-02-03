package game

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type SkillEffect struct {
	ID    int
	Eid   int
	Stat  int
	Times int
	Args  string
	Desc  string
	Desc2 string
	Item  int
}

type SkillEffectDB struct {
	mu      sync.RWMutex
	loaded  bool
	effects map[int]*SkillEffect
}

var globalSkillEffects = &SkillEffectDB{}

func LoadSkillEffects() *SkillEffectDB {
	globalSkillEffects.mu.Lock()
	defer globalSkillEffects.mu.Unlock()
	if globalSkillEffects.loaded {
		return globalSkillEffects
	}
	globalSkillEffects.effects = make(map[int]*SkillEffect)
	path := resolveSkillEffectsPath()
	if path != "" {
		_ = loadSkillEffects(path, globalSkillEffects.effects)
	}
	globalSkillEffects.loaded = true
	return globalSkillEffects
}

func getSkillEffect(effectID int) *SkillEffect {
	if effectID <= 0 {
		return nil
	}
	db := LoadSkillEffects()
	return db.effects[effectID]
}

func resolveSkillEffectsPath() string {
	if v := os.Getenv("JSEER_SKILL_EFFECTS_PATH"); v != "" {
		return v
	}
	path := filepath.Join(resolveDataRoot(), "skill_effects.xml")
	if st, err := os.Stat(path); err == nil && !st.IsDir() {
		return path
	}
	return ""
}

func loadSkillEffects(path string, out map[int]*SkillEffect) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local != "NewSeIdx" {
			continue
		}
		effect := &SkillEffect{}
		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "Idx":
				effect.ID, _ = strconv.Atoi(attr.Value)
			case "Eid":
				effect.Eid, _ = strconv.Atoi(attr.Value)
			case "Stat":
				effect.Stat, _ = strconv.Atoi(attr.Value)
			case "Times":
				effect.Times, _ = strconv.Atoi(attr.Value)
			case "Args":
				effect.Args = attr.Value
			case "Desc", "Des":
				effect.Desc = attr.Value
			case "Desc2":
				effect.Desc2 = attr.Value
			case "ItemId":
				effect.Item, _ = strconv.Atoi(attr.Value)
			}
		}
		if effect.ID > 0 {
			out[effect.ID] = effect
		}
	}
	return nil
}
