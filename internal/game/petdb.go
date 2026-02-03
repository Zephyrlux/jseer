package game

import (
	"encoding/xml"
	"errors"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LearnableMove struct {
	ID    int
	Level int
}

type PetBase struct {
	ID         int
	Type       int
	BaseExp    int
	Name       string
	Hp         int
	Atk        int
	Def        int
	SpAtk      int
	SpDef      int
	Spd        int
	GrowthType int
	Learnable  []LearnableMove
}

type SkillInfo struct {
	ID            int
	PP            int
	Power         int
	Type          int
	Category      int
	Accuracy      int
	Priority      int
	SideEffect    int
	SideEffectArg string
	MustHit       bool
}

type PetDB struct {
	mu     sync.RWMutex
	loaded bool
	pets   map[int]*PetBase
	skills map[int]*SkillInfo
	names  map[string]int
}

var globalPetDB = &PetDB{}

func LoadPetDB() *PetDB {
	globalPetDB.mu.Lock()
	defer globalPetDB.mu.Unlock()
	if globalPetDB.loaded {
		return globalPetDB
	}
	globalPetDB.pets = make(map[int]*PetBase)
	globalPetDB.skills = make(map[int]*SkillInfo)
	globalPetDB.names = make(map[string]int)
	dataRoot := resolveDataRoot()
	_ = loadSkills(filepath.Join(dataRoot, "skills.xml"), globalPetDB.skills)
	_ = loadPets(filepath.Join(dataRoot, "spt.xml"), globalPetDB.pets)
	for id, pet := range globalPetDB.pets {
		if pet != nil && pet.Name != "" {
			globalPetDB.names[pet.Name] = id
		}
	}
	globalPetDB.loaded = true
	return globalPetDB
}

func resolveDataRoot() string {
	if v := os.Getenv("JSEER_DATA_ROOT"); v != "" {
		return v
	}
	candidates := []string{
		filepath.Join("..", "Reseer-main", "luvit_version", "data"),
		filepath.Join("..", "Reseer-main", "data"),
		filepath.Join("Reseer-main", "luvit_version", "data"),
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return "."
}

func loadSkills(path string, out map[int]*SkillInfo) error {
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
		if se.Name.Local != "Move" {
			continue
		}
		id := 0
		pp := 35
		power := 0
		typ := 0
		category := 0
		accuracy := 100
		priority := 0
		sideEffect := 0
		sideArg := ""
		mustHit := false
		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "ID":
				id, _ = strconv.Atoi(attr.Value)
			case "MaxPP":
				pp, _ = strconv.Atoi(attr.Value)
			case "Power":
				power, _ = strconv.Atoi(attr.Value)
			case "Type":
				typ, _ = strconv.Atoi(attr.Value)
			case "Category":
				category, _ = strconv.Atoi(attr.Value)
			case "Accuracy":
				accuracy, _ = strconv.Atoi(attr.Value)
			case "Priority":
				priority, _ = strconv.Atoi(attr.Value)
			case "SideEffect":
				sideEffect, _ = strconv.Atoi(attr.Value)
			case "SideEffectArg":
				sideArg = attr.Value
			case "MustHit":
				mustHit = attr.Value == "1"
			}
		}
		if id > 0 {
			out[id] = &SkillInfo{
				ID:            id,
				PP:            pp,
				Power:         power,
				Type:          typ,
				Category:      category,
				Accuracy:      accuracy,
				Priority:      priority,
				SideEffect:    sideEffect,
				SideEffectArg: sideArg,
				MustHit:       mustHit,
			}
		}
	}
	return nil
}

func loadPets(path string, out map[int]*PetBase) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	var current *PetBase
	inLearnable := false
	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := t.Name.Local
			switch name {
			case "Monster":
				current = &PetBase{}
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "ID":
						current.ID, _ = strconv.Atoi(attr.Value)
					case "Type":
						current.Type, _ = strconv.Atoi(attr.Value)
					case "YieldingExp":
						current.BaseExp, _ = strconv.Atoi(attr.Value)
					case "DefName":
						current.Name = attr.Value
					case "Hp", "HP":
						current.Hp, _ = strconv.Atoi(attr.Value)
					case "Atk":
						current.Atk, _ = strconv.Atoi(attr.Value)
					case "Def":
						current.Def, _ = strconv.Atoi(attr.Value)
					case "SpAtk", "SpA":
						current.SpAtk, _ = strconv.Atoi(attr.Value)
					case "SpDef", "SpD":
						current.SpDef, _ = strconv.Atoi(attr.Value)
					case "Spd", "Speed":
						current.Spd, _ = strconv.Atoi(attr.Value)
					case "GrowthType":
						current.GrowthType, _ = strconv.Atoi(attr.Value)
					}
				}
			case "LearnableMoves":
				if current != nil {
					inLearnable = true
				}
			case "Move":
				if current != nil && inLearnable {
					moveID := 0
					moveLv := 0
					for _, attr := range t.Attr {
						switch attr.Name.Local {
						case "ID":
							moveID, _ = strconv.Atoi(attr.Value)
						case "LearningLv":
							moveLv, _ = strconv.Atoi(attr.Value)
						}
					}
					if moveID > 0 {
						current.Learnable = append(current.Learnable, LearnableMove{ID: moveID, Level: moveLv})
					}
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "LearnableMoves":
				inLearnable = false
			case "Monster":
				if current != nil && current.ID > 0 {
					out[current.ID] = current
				}
				current = nil
			}
		}
	}
	return nil
}

func createStarterPet(petID int, level int) *PetInstance {
	db := LoadPetDB()
	base := db.pets[petID]
	if level == 0 {
		level = 5
	}
	if base == nil {
		return &PetInstance{
			ID:         petID,
			Level:      level,
			DV:         31,
			Nature:     rand.Intn(25),
			Exp:        0,
			HP:         100,
			MaxHP:      100,
			Attack:     20,
			Defence:    20,
			SA:         20,
			SD:         20,
			Speed:      20,
			Skills:     []int{10001, 0, 0, 0},
			CatchMap:   301,
			CatchLevel: level,
		}
	}
	dv := rand.Intn(32)
	nature := rand.Intn(25)
	stats := getStats(base, level, dv, evSet{})
	skills := getSkillsForLevel(base, level)
	return &PetInstance{
		ID:         petID,
		Level:      level,
		DV:         dv,
		Nature:     nature,
		Exp:        0,
		HP:         stats.HP,
		MaxHP:      stats.MaxHP,
		Attack:     stats.Attack,
		Defence:    stats.Defence,
		SA:         stats.SA,
		SD:         stats.SD,
		Speed:      stats.Speed,
		Skills:     skills,
		CatchMap:   301,
		CatchLevel: level,
	}
}

type evSet struct {
	HP  int
	Atk int
	Def int
	SpA int
	SpD int
	Spd int
}

type petStats struct {
	HP      int
	MaxHP   int
	Attack  int
	Defence int
	SA      int
	SD      int
	Speed   int
}

func getStats(base *PetBase, level int, dv int, ev evSet) petStats {
	if base == nil {
		return petStats{HP: 20, MaxHP: 20, Attack: 10, Defence: 10, SA: 10, SD: 10, Speed: 10}
	}
	if level <= 0 {
		level = 1
	}
	hp := ((base.Hp*2 + dv + ev.HP/4) * level / 100) + level + 10
	atk := ((base.Atk*2 + dv + ev.Atk/4) * level / 100) + 5
	def := ((base.Def*2 + dv + ev.Def/4) * level / 100) + 5
	spa := ((base.SpAtk*2 + dv + ev.SpA/4) * level / 100) + 5
	spd := ((base.SpDef*2 + dv + ev.SpD/4) * level / 100) + 5
	speed := ((base.Spd*2 + dv + ev.Spd/4) * level / 100) + 5
	return petStats{
		HP:      hp,
		MaxHP:   hp,
		Attack:  atk,
		Defence: def,
		SA:      spa,
		SD:      spd,
		Speed:   speed,
	}
}

func getSkillsForLevel(base *PetBase, level int) []int {
	if base == nil {
		return []int{10001, 0, 0, 0}
	}
	skills := make([]int, 0, len(base.Learnable))
	for _, mv := range base.Learnable {
		if mv.Level <= level {
			skills = append(skills, mv.ID)
		}
	}
	if len(skills) > 4 {
		skills = skills[len(skills)-4:]
	}
	for len(skills) < 4 {
		skills = append(skills, 0)
	}
	return skills
}

type expInfo struct {
	Exp       int
	LvExp     int
	NextLvExp int
}

func getExpInfo(base *PetBase, level int, currentLevelExp int) expInfo {
	if base == nil {
		return expInfo{Exp: currentLevelExp, LvExp: 0, NextLvExp: 100}
	}
	g := base.GrowthType
	next := 0
	switch g {
	case 0:
		next = int(float64(level*level*level) * 0.8)
	case 1:
		next = level * level * level
	case 2:
		next = int(float64(level*level*level) * 1.2)
	case 3:
		next = int(float64(level*level*level) * 1.5)
	default:
		next = level * level * level
	}
	total := 0
	for lv := 1; lv <= level-1; lv++ {
		switch g {
		case 0:
			total += int(float64(lv*lv*lv) * 0.8)
		case 1:
			total += lv * lv * lv
		case 2:
			total += int(float64(lv*lv*lv) * 1.2)
		case 3:
			total += int(float64(lv*lv*lv) * 1.5)
		default:
			total += lv * lv * lv
		}
	}
	total += currentLevelExp
	return expInfo{
		Exp:       currentLevelExp,
		LvExp:     currentLevelExp,
		NextLvExp: next,
	}
}

type PetInstance struct {
	ID         int
	Level      int
	DV         int
	Nature     int
	Exp        int
	HP         int
	MaxHP      int
	Attack     int
	Defence    int
	SA         int
	SD         int
	Speed      int
	Skills     []int
	CatchMap   int
	CatchLevel int
}

func getSkillPP(skillID int) int {
	db := LoadPetDB()
	if s := db.skills[skillID]; s != nil {
		return s.PP
	}
	return 20
}

func getSkillInfo(skillID int) *SkillInfo {
	db := LoadPetDB()
	return db.skills[skillID]
}

func findPetIDByName(name string) int {
	if name == "" {
		return 0
	}
	db := LoadPetDB()
	if id := db.names[name]; id > 0 {
		return id
	}
	return 0
}

func sanitizeName(name string) string {
	return strings.TrimRight(name, "\x00")
}
