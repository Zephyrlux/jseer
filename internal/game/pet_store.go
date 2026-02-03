package game

import (
	"context"
	"encoding/json"

	"jseer/internal/storage"
)

func encodePetSkills(skills []int) string {
	if len(skills) == 0 {
		return "[]"
	}
	data, err := json.Marshal(skills)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodePetSkills(raw string) []int {
	if raw == "" {
		return nil
	}
	var list []int
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil
	}
	return list
}

func upsertPet(deps *Deps, user *User, pet Pet) {
	if deps == nil || deps.Store == nil || user == nil || user.PlayerID == 0 {
		return
	}
	_, _ = deps.Store.UpsertPet(context.Background(), &storage.Pet{
		PlayerID:  user.PlayerID,
		SpeciesID: int(pet.ID),
		Level:     int(pet.Level),
		Exp:       pet.Exp,
		HP:        pet.HP,
		CatchTime: int64(pet.CatchTime),
		DV:        int(pet.DV),
		Skills:    encodePetSkills(pet.Skills),
		Nature:    "normal",
	})
}
