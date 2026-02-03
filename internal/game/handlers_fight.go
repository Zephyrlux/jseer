package game

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"time"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

const noviceBossID = 13

func registerFightHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2411, handleChallengeBoss(state))
	s.Register(2404, handleReadyToFight(state))
	s.Register(2405, handleUseSkill(state))
	s.Register(2406, handleUsePetItem())
	s.Register(2407, handleChangePet(state))
	s.Register(2409, handleCatchMonster(state))
	s.Register(2410, handleEscapeFight(state))
}

func handleChallengeBoss(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		bossID := int(reader.ReadUint32BE())
		if bossID == 0 {
			bossID = noviceBossID
		}

		user := state.GetOrCreateUser(ctx.UserID)
		petID, level, dv, _, catchTime, skills := pickCurrentPet(user)
		stats := getStats(LoadPetDB().pets[int(petID)], int(level), int(dv), evSet{})
		if len(skills) == 0 {
			skills = getSkillsForLevel(LoadPetDB().pets[int(petID)], int(level))
		}
		bossLevel := 1
		bossStats := getStats(LoadPetDB().pets[bossID], bossLevel, 15, evSet{})
		bossSkills := getSkillsForLevel(LoadPetDB().pets[bossID], bossLevel)

		user.Fight = &FightState{
			UserID:       ctx.UserID,
			PlayerPetID:  petID,
			PlayerLevel:  level,
			PlayerHP:     stats.HP,
			PlayerMaxHP:  stats.MaxHP,
			PlayerCatch:  catchTime,
			PlayerSkills: skills,
			EnemyPetID:   uint32(bossID),
			EnemyLevel:   uint32(bossLevel),
			EnemyHP:      bossStats.HP,
			EnemyMaxHP:   bossStats.MaxHP,
			EnemyCatch:   0,
			EnemySkills:  bossSkills,
		}

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(2))
		buf.Write(buildFightUserInfo(ctx.UserID, pickNick(user, ctx.UserID)))
		binary.Write(buf, binary.BigEndian, uint32(1))
		buf.Write(buildSimpleFightPetInfo(int(petID), int(level), stats.HP, stats.MaxHP, int(catchTime), skills))
		binary.Write(buf, binary.BigEndian, uint32(0))
		protocol.WriteFixedString(buf, "", 16)
		binary.Write(buf, binary.BigEndian, uint32(1))
		buf.Write(buildSimpleFightPetInfo(bossID, bossLevel, bossStats.HP, bossStats.MaxHP, 0, bossSkills))
		ctx.Server.SendResponse(ctx.Conn, 2503, ctx.UserID, buf.Bytes())
	}
}

func handleReadyToFight(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Fight == nil {
			_ = handleChallengeBoss(state)(ctx)
		}
		f := user.Fight
		if f == nil {
			return
		}
		user.InFight = true
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		buf.Write(buildFightPetInfo(ctx.UserID, f.PlayerPetID, f.PlayerCatch, f.PlayerHP, f.PlayerMaxHP, f.PlayerLevel, 0))
		buf.Write(buildFightPetInfo(0, f.EnemyPetID, f.EnemyCatch, f.EnemyHP, f.EnemyMaxHP, f.EnemyLevel, 1))
		ctx.Server.SendResponse(ctx.Conn, 2504, ctx.UserID, buf.Bytes())
	}
}

func handleUseSkill(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		skillID := reader.ReadUint32BE()
		ctx.Server.SendResponse(ctx.Conn, 2405, ctx.UserID, []byte{})

		user := state.GetOrCreateUser(ctx.UserID)
		f := user.Fight
		if f == nil {
			sendFightOver(ctx, 0, 0)
			return
		}
		rand.Seed(time.Now().UnixNano())
		playerDamage := 8 + rand.Intn(6)
		enemyDamage := 4 + rand.Intn(5)

		f.EnemyHP -= playerDamage
		if f.EnemyHP < 0 {
			f.EnemyHP = 0
		}
		if f.EnemyHP > 0 {
			f.PlayerHP -= enemyDamage
			if f.PlayerHP < 0 {
				f.PlayerHP = 0
			}
		}

		buf := new(bytes.Buffer)
		buf.Write(buildAttackValue(ctx.UserID, skillID, 1, uint32(playerDamage), 0, f.PlayerHP, f.PlayerMaxHP, 0, 0, 0))
		buf.Write(buildAttackValue(0, 0, 1, uint32(enemyDamage), 0, f.EnemyHP, f.EnemyMaxHP, 0, 0, 0))
		ctx.Server.SendResponse(ctx.Conn, 2505, ctx.UserID, buf.Bytes())

		if f.EnemyHP == 0 || f.PlayerHP == 0 {
			winner := uint32(0)
			if f.EnemyHP == 0 {
				winner = ctx.UserID
			}
			sendFightOver(ctx, winner, 0)
			user.Fight = nil
			user.InFight = false
		}
	}
}

func handleUsePetItem() gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		itemID := reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, itemID)
		binary.Write(buf, binary.BigEndian, uint32(100))
		binary.Write(buf, binary.BigEndian, uint32(50))
		ctx.Server.SendResponse(ctx.Conn, 2406, ctx.UserID, buf.Bytes())
	}
}

func handleChangePet(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchTime := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		petID := user.CurrentPetID
		level := uint32(16)
		hp := uint32(100)
		maxHP := uint32(100)
		if petID == 0 && len(user.Pets) > 0 {
			petID = user.Pets[0].ID
			level = user.Pets[0].Level
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, petID)
		protocol.WriteFixedString(buf, "", 16)
		binary.Write(buf, binary.BigEndian, level)
		binary.Write(buf, binary.BigEndian, hp)
		binary.Write(buf, binary.BigEndian, maxHP)
		binary.Write(buf, binary.BigEndian, catchTime)
		ctx.Server.SendResponse(ctx.Conn, 2407, ctx.UserID, buf.Bytes())
	}
}

func handleCatchMonster(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		bossID := uint32(noviceBossID)
		if user.Fight != nil {
			bossID = user.Fight.EnemyPetID
		}
		catchTime := uint32(time.Now().Unix())
		user.Pets = append(user.Pets, Pet{
			ID:        bossID,
			CatchTime: catchTime,
			Level:     1,
			DV:        15,
			Exp:       0,
			HP:        50,
		})
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, catchTime)
		binary.Write(buf, binary.BigEndian, bossID)
		ctx.Server.SendResponse(ctx.Conn, 2409, ctx.UserID, buf.Bytes())
	}
}

func handleEscapeFight(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		user.Fight = nil
		user.InFight = false
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(1))
		ctx.Server.SendResponse(ctx.Conn, 2410, ctx.UserID, buf.Bytes())
	}
}

func buildFightUserInfo(userID uint32, nick string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, userID)
	protocol.WriteFixedString(buf, nick, 16)
	return buf.Bytes()
}

func buildSimpleFightPetInfo(petID int, level int, hp int, maxHP int, catchTime int, skills []int) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(petID))
	binary.Write(buf, binary.BigEndian, uint32(level))
	binary.Write(buf, binary.BigEndian, uint32(hp))
	binary.Write(buf, binary.BigEndian, uint32(maxHP))
	valid := 0
	for _, sid := range skills {
		if sid > 0 {
			valid++
		}
	}
	binary.Write(buf, binary.BigEndian, uint32(valid))
	for i := 0; i < 4; i++ {
		sid := 0
		if i < len(skills) {
			sid = skills[i]
		}
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
	binary.Write(buf, binary.BigEndian, uint32(0))
	return buf.Bytes()
}

func buildFightPetInfo(userID uint32, petID uint32, catchTime uint32, hp int, maxHP int, level uint32, catchable uint32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, userID)
	binary.Write(buf, binary.BigEndian, petID)
	protocol.WriteFixedString(buf, "", 16)
	binary.Write(buf, binary.BigEndian, catchTime)
	binary.Write(buf, binary.BigEndian, uint32(hp))
	binary.Write(buf, binary.BigEndian, uint32(maxHP))
	binary.Write(buf, binary.BigEndian, level)
	binary.Write(buf, binary.BigEndian, catchable)
	buf.Write(make([]byte, 6))
	return buf.Bytes()
}

func buildAttackValue(userID uint32, skillID uint32, atkTimes uint32, lostHP uint32, gainHP int, remainHP int, maxHP int, state uint32, isCrit uint32, petType uint32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, userID)
	binary.Write(buf, binary.BigEndian, skillID)
	binary.Write(buf, binary.BigEndian, atkTimes)
	binary.Write(buf, binary.BigEndian, lostHP)
	binary.Write(buf, binary.BigEndian, int32(gainHP))
	binary.Write(buf, binary.BigEndian, int32(remainHP))
	binary.Write(buf, binary.BigEndian, uint32(maxHP))
	binary.Write(buf, binary.BigEndian, state)
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, isCrit)
	buf.Write(make([]byte, 20))
	buf.Write(make([]byte, 6))
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, petType)
	return buf.Bytes()
}

func sendFightOver(ctx *gateway.Context, winner uint32, reason uint32) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, reason)
	binary.Write(buf, binary.BigEndian, winner)
	buf.Write(make([]byte, 20))
	ctx.Server.SendResponse(ctx.Conn, 2506, ctx.UserID, buf.Bytes())
}

func pickCurrentPet(user *User) (petID uint32, level uint32, dv uint32, exp int, catchTime uint32, skills []int) {
	petID = user.CurrentPetID
	catchTime = user.CatchID
	if petID == 0 && len(user.Pets) > 0 {
		petID = user.Pets[0].ID
		catchTime = user.Pets[0].CatchTime
		level = user.Pets[0].Level
		dv = user.Pets[0].DV
		exp = user.Pets[0].Exp
		skills = append([]int{}, user.Pets[0].Skills...)
	}
	if petID == 0 {
		petID = 7
	}
	if level == 0 {
		level = 5
	}
	if dv == 0 {
		dv = 31
	}
	return
}
