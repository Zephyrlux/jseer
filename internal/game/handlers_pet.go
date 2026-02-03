package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerPetHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2301, handleGetPetInfo(state))
	s.Register(2303, handleGetPetList())
	s.Register(2304, handlePetRelease(deps, state))
	s.Register(2305, handlePetShow(state))
	s.Register(2306, handlePetCure())
	s.Register(2309, handlePetBargeList(state))
	s.Register(2354, handleGetSoulBeadList())
}

func handleGetPetInfo(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		petID := uint32(7)
		level := 5
		dv := 31
		exp := 0
		if user.CurrentPetID != 0 {
			petID = user.CurrentPetID
		}
		targetCatch := catchID
		if targetCatch == 0 {
			targetCatch = user.CatchID
		}
		var skills []int
		if len(user.Pets) > 0 && targetCatch != 0 {
			for _, p := range user.Pets {
				if uint32(p.CatchTime) == targetCatch {
					petID = uint32(p.ID)
					level = int(p.Level)
					dv = int(p.DV)
					exp = p.Exp
					skills = append([]int{}, p.Skills...)
					break
				}
			}
		}
		catchTime := int(catchID)
		if catchTime == 0 {
			catchTime = int(targetCatch)
		}
		body := buildFullPetInfo(int(petID), catchTime, level, dv, exp, skills)
		ctx.Server.SendResponse(ctx.Conn, 2301, ctx.UserID, body)
	}
}

func handleGetPetList() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2303, ctx.UserID, buf.Bytes())
	}
}

func handlePetRelease(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchID := reader.ReadUint32BE()
		_ = reader.ReadUint32BE() // flag

		petType := int(catchID - 0x69686700)
		if petType < 1 || petType > 2000 {
			petType = 7
		}

		user := state.GetOrCreateUser(ctx.UserID)
		user.CurrentPetID = uint32(petType)
		user.CatchID = catchID
		savePlayer(deps, ctx.UserID, user)

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0)) // homeEnergy
		binary.Write(buf, binary.BigEndian, catchID)
		binary.Write(buf, binary.BigEndian, uint32(1))
		buf.Write(buildFullPetInfo(petType, int(catchID), 5, 31, 0, nil))
		ctx.Server.SendResponse(ctx.Conn, 2304, ctx.UserID, buf.Bytes())
	}
}

func handlePetShow(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		reqCatch := reader.ReadUint32BE()
		reqFlag := reader.ReadUint32BE()

		user := state.GetOrCreateUser(ctx.UserID)
		petID := user.CurrentPetID
		if petID == 0 {
			petID = 7
		}
		catchTime := user.CatchID
		if catchTime == 0 {
			catchTime = 0x69686700 + petID
		}
		if reqCatch > 0 {
			catchTime = reqCatch
		}
		dv := uint32(31)
		for _, p := range user.Pets {
			if p.CatchTime == catchTime {
				dv = p.DV
				petID = p.ID
				break
			}
		}

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, catchTime)
		binary.Write(buf, binary.BigEndian, petID)
		binary.Write(buf, binary.BigEndian, reqFlag)
		binary.Write(buf, binary.BigEndian, dv)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2305, ctx.UserID, buf.Bytes())
	}
}

func handlePetCure() gateway.Handler {
	return func(ctx *gateway.Context) {
		// Keep ack-only to match protocol expectations; healing is handled by 2310 for single pet.
		ctx.Server.SendResponse(ctx.Conn, 2306, ctx.UserID, []byte{})
	}
}

func handlePetBargeList(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		reqType := reader.ReadUint32BE()
		maxID := reader.ReadUint32BE()
		if maxID == 0 {
			maxID = 1500
		}

		user := state.GetOrCreateUser(ctx.UserID)
		caughtSet := map[int]bool{}
		for _, p := range user.Pets {
			caughtSet[int(p.ID)] = true
		}

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, maxID)
		binary.Write(buf, binary.BigEndian, reqType)
		for i := uint32(1); i <= maxID; i++ {
			binary.Write(buf, binary.BigEndian, uint32(0))
			encountered := uint32(0)
			caught := uint32(0)
			if caughtSet[int(i)] {
				encountered = 1
				caught = 1
			}
			binary.Write(buf, binary.BigEndian, encountered)
			binary.Write(buf, binary.BigEndian, caught)
			binary.Write(buf, binary.BigEndian, i)
		}
		ctx.Server.SendResponse(ctx.Conn, 2309, ctx.UserID, buf.Bytes())
	}
}

func handleGetSoulBeadList() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2354, ctx.UserID, buf.Bytes())
	}
}

func buildFullPetInfo(petID int, catchTime int, level int, dv int, exp int, skillsOverride []int) []byte {
	db := LoadPetDB()
	base := db.pets[petID]
	if level <= 0 {
		level = 5
	}
	if dv <= 0 {
		dv = 31
	}
	stats := getStats(base, level, dv, evSet{})
	expInfo := getExpInfo(base, level, exp)
	skills := skillsOverride
	if len(skills) == 0 {
		skills = getSkillsForLevel(base, level)
	} else {
		if len(skills) > 4 {
			skills = skills[len(skills)-4:]
		}
		for len(skills) < 4 {
			skills = append(skills, 0)
		}
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(petID))
	protocol.WriteFixedString(buf, "", 16)
	binary.Write(buf, binary.BigEndian, uint32(dv))
	binary.Write(buf, binary.BigEndian, uint32(0)) // nature
	binary.Write(buf, binary.BigEndian, uint32(level))
	binary.Write(buf, binary.BigEndian, uint32(expInfo.Exp))
	binary.Write(buf, binary.BigEndian, uint32(expInfo.LvExp))
	binary.Write(buf, binary.BigEndian, uint32(expInfo.NextLvExp))

	binary.Write(buf, binary.BigEndian, uint32(stats.HP))
	binary.Write(buf, binary.BigEndian, uint32(stats.MaxHP))
	binary.Write(buf, binary.BigEndian, uint32(stats.Attack))
	binary.Write(buf, binary.BigEndian, uint32(stats.Defence))
	binary.Write(buf, binary.BigEndian, uint32(stats.SA))
	binary.Write(buf, binary.BigEndian, uint32(stats.SD))
	binary.Write(buf, binary.BigEndian, uint32(stats.Speed))

	// EVs
	for i := 0; i < 6; i++ {
		binary.Write(buf, binary.BigEndian, uint32(0))
	}

	validCount := 0
	for i := 0; i < 4; i++ {
		if skills[i] > 0 {
			validCount++
		}
	}
	binary.Write(buf, binary.BigEndian, uint32(validCount))
	for i := 0; i < 4; i++ {
		sid := skills[i]
		pp := 0
		if sid > 0 {
			pp = getSkillPP(sid)
		}
		binary.Write(buf, binary.BigEndian, uint32(sid))
		binary.Write(buf, binary.BigEndian, uint32(pp))
	}

	binary.Write(buf, binary.BigEndian, uint32(catchTime))
	binary.Write(buf, binary.BigEndian, uint32(301))
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(level))
	protocol.WriteUint16BE(buf, 0)
	binary.Write(buf, binary.BigEndian, uint32(0))
	return buf.Bytes()
}
