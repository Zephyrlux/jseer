package game

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

const noviceBossID = 13

func init() {
	rand.Seed(time.Now().UnixNano())
}

func registerFightHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2411, handleChallengeBoss(state))
	s.Register(2404, handleReadyToFight(state))
	s.Register(2405, handleUseSkill(deps, state))
	s.Register(2406, handleUsePetItem())
	s.Register(2407, handleChangePet(state))
	s.Register(2409, handleCatchMonster(deps, state))
	s.Register(2410, handleEscapeFight(deps, state))
}

func handleChallengeBoss(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		bossID := int(reader.ReadUint32BE())
		if bossID == 0 {
			bossID = noviceBossID
		}

		user := state.GetOrCreateUser(ctx.UserID)
		player := resolveUserFightPet(user, user.CatchID, user.CurrentPetID)
		if user.CatchID == 0 {
			user.CatchID = player.CatchTime
		}
		bossLevel := 1
		bossHP := 0
		bossRewardID := 0
		bossRewardName := ""
		bossRewardCount := 0
		if cfg := GetSPTBossByID(bossID); cfg != nil {
			if cfg.Level > 0 {
				bossLevel = cfg.Level
			}
			if cfg.MaxHP > 0 {
				bossHP = cfg.MaxHP
			}
			if cfg.RewardItemID > 0 {
				bossRewardID = cfg.RewardItemID
				bossRewardName = cfg.RewardName
				bossRewardCount = cfg.RewardCount
			}
		}
		enemy := resolveEnemyFightPet(bossID, bossLevel)
		if bossHP > 0 {
			enemy.Stats.MaxHP = bossHP
			enemy.Stats.HP = bossHP
			enemy.CurrentHP = bossHP
		}
		if enemy.CatchTime == 0 {
			enemy.CatchTime = uint32(time.Now().Unix())
		}

		user.Fight = &FightState{
			UserID:        ctx.UserID,
			PlayerPetID:   player.ID,
			PlayerLevel:   player.Level,
			PlayerHP:      player.CurrentHP,
			PlayerMaxHP:   player.Stats.MaxHP,
			PlayerCatch:   player.CatchTime,
			PlayerSkills:  player.Skills,
			PlayerStats:   player.Stats,
			PlayerType:    player.Type,
			EnemyPetID:    enemy.ID,
			EnemyLevel:    enemy.Level,
			EnemyHP:       enemy.CurrentHP,
			EnemyMaxHP:    enemy.Stats.MaxHP,
			EnemyCatch:    enemy.CatchTime,
			EnemySkills:   enemy.Skills,
			EnemyStats:    enemy.Stats,
			EnemyType:     enemy.Type,
			EnemyRewardID: bossRewardID,
			EnemyRewardNm: bossRewardName,
			EnemyRewardCt: bossRewardCount,
		}

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(2))
		buf.Write(buildFightUserInfo(ctx.UserID, pickNick(user, ctx.UserID)))
		binary.Write(buf, binary.BigEndian, uint32(1))
		buf.Write(buildSimpleFightPetInfo(int(player.ID), int(player.Level), player.CurrentHP, player.Stats.MaxHP, int(player.CatchTime), player.Skills, int(user.MapID)))
		binary.Write(buf, binary.BigEndian, uint32(0))
		protocol.WriteFixedString(buf, "", 16)
		binary.Write(buf, binary.BigEndian, uint32(1))
		buf.Write(buildSimpleFightPetInfo(int(enemy.ID), int(enemy.Level), enemy.CurrentHP, enemy.Stats.MaxHP, int(enemy.CatchTime), enemy.Skills, 301))
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
		ensureFightStats(user, f)
		ensureFightStatus(f)
		user.InFight = true
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		buf.Write(buildFightPetInfo(ctx.UserID, f.PlayerPetID, f.PlayerCatch, f.PlayerHP, f.PlayerMaxHP, f.PlayerLevel, 0))
		buf.Write(buildFightPetInfo(0, f.EnemyPetID, f.EnemyCatch, f.EnemyHP, f.EnemyMaxHP, f.EnemyLevel, 1))
		ctx.Server.SendResponse(ctx.Conn, 2504, ctx.UserID, buf.Bytes())
	}
}

func handleUseSkill(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		reqSkillID := int(reader.ReadUint32BE())
		ctx.Server.SendResponse(ctx.Conn, 2405, ctx.UserID, []byte{})

		user := state.GetOrCreateUser(ctx.UserID)
		f := user.Fight
		if f == nil {
			sendFightOver(ctx, 0, 0)
			return
		}
		ensureFightStats(user, f)

		playerSkill := selectPlayerSkill(f.PlayerSkills, reqSkillID)
		enemySkill := selectRandomSkill(f.EnemySkills)
		playerSpeed := applyStageModifier(f.PlayerStats.Speed, f.PlayerStage.Spd)
		enemySpeed := applyStageModifier(f.EnemyStats.Speed, f.EnemyStage.Spd)
		playerFirst := isPlayerFirst(playerSkill, enemySkill, playerSpeed, enemySpeed)

		var first attackResult
		var second attackResult
		if playerFirst {
			first = executeAttack(ctx, f, true, playerSkill)
			if f.EnemyHP > 0 {
				second = executeAttack(ctx, f, false, enemySkill)
			} else {
				second = placeholderAttack(false, f)
			}
		} else {
			first = executeAttack(ctx, f, false, enemySkill)
			if f.PlayerHP > 0 {
				second = executeAttack(ctx, f, true, playerSkill)
			} else {
				second = placeholderAttack(true, f)
			}
		}

		buf := new(bytes.Buffer)
		firstStatus := f.PlayerStatus
		if first.UserID == 0 {
			firstStatus = f.EnemyStatus
		}
		secondStatus := f.PlayerStatus
		if second.UserID == 0 {
			secondStatus = f.EnemyStatus
		}
		buf.Write(buildAttackValue(first.UserID, first.SkillID, first.AtkTimes, first.LostHP, first.GainHP, first.RemainHP, first.MaxHP, first.State, first.IsCrit, first.PetType, first.Stage, firstStatus))
		buf.Write(buildAttackValue(second.UserID, second.SkillID, second.AtkTimes, second.LostHP, second.GainHP, second.RemainHP, second.MaxHP, second.State, second.IsCrit, second.PetType, second.Stage, secondStatus))
		ctx.Server.SendResponse(ctx.Conn, 2505, ctx.UserID, buf.Bytes())

		if f.EnemyHP == 0 || f.PlayerHP == 0 {
			winner := uint32(0)
			if f.EnemyHP == 0 {
				winner = ctx.UserID
			}
			updateFightResult(deps, user, f, winner == ctx.UserID)
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
		pet := resolveUserFightPet(user, catchTime, user.CurrentPetID)
		user.CurrentPetID = pet.ID
		if pet.CatchTime != 0 {
			user.CatchID = pet.CatchTime
		}
		if user.Fight != nil {
			user.Fight.PlayerPetID = pet.ID
			user.Fight.PlayerLevel = pet.Level
			user.Fight.PlayerCatch = pet.CatchTime
			user.Fight.PlayerSkills = pet.Skills
			user.Fight.PlayerStats = pet.Stats
			user.Fight.PlayerType = pet.Type
			user.Fight.PlayerHP = pet.CurrentHP
			user.Fight.PlayerMaxHP = pet.Stats.MaxHP
		}
		hp := uint32(pet.CurrentHP)
		maxHP := uint32(pet.Stats.MaxHP)
		respCatch := pet.CatchTime
		if respCatch == 0 {
			respCatch = catchTime
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, pet.ID)
		protocol.WriteFixedString(buf, "", 16)
		binary.Write(buf, binary.BigEndian, pet.Level)
		binary.Write(buf, binary.BigEndian, hp)
		binary.Write(buf, binary.BigEndian, maxHP)
		binary.Write(buf, binary.BigEndian, respCatch)
		ctx.Server.SendResponse(ctx.Conn, 2407, ctx.UserID, buf.Bytes())
	}
}

func handleCatchMonster(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		bossID := uint32(noviceBossID)
		if user.Fight != nil {
			bossID = user.Fight.EnemyPetID
		}
		catchTime := uint32(time.Now().Unix())
		if user.Fight != nil && user.Fight.EnemyCatch > 0 {
			catchTime = user.Fight.EnemyCatch
		}
		newPet := Pet{
			ID:        bossID,
			CatchTime: catchTime,
			Level:     1,
			DV:        15,
			Exp:       0,
			HP:        50,
		}
		user.Pets = append(user.Pets, newPet)
		upsertPet(deps, user, newPet)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, catchTime)
		binary.Write(buf, binary.BigEndian, bossID)
		ctx.Server.SendResponse(ctx.Conn, 2409, ctx.UserID, buf.Bytes())
	}
}

func handleEscapeFight(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Fight != nil {
			updateFightResult(deps, user, user.Fight, false)
		}
		user.Fight = nil
		user.InFight = false
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(1))
		ctx.Server.SendResponse(ctx.Conn, 2410, ctx.UserID, buf.Bytes())
	}
}

func updateFightResult(deps *Deps, user *User, f *FightState, won bool) {
	if user == nil || f == nil {
		return
	}
	for i := range user.Pets {
		if user.Pets[i].CatchTime != f.PlayerCatch {
			continue
		}
		p := &user.Pets[i]
		p.HP = f.PlayerHP
		if won {
			expGain := calculateExpGain(int(f.EnemyPetID), int(f.EnemyLevel), true)
			p.Exp += expGain
			base := LoadPetDB().pets[int(p.ID)]
			for {
				info := getExpInfo(base, int(p.Level), p.Exp)
				if p.Exp < info.NextLvExp {
					break
				}
				p.Exp -= info.NextLvExp
				p.Level++
			}
			stats := getStats(base, int(p.Level), int(p.DV), evSet{})
			if p.HP > stats.MaxHP {
				p.HP = stats.MaxHP
			}
		}
		upsertPet(deps, user, *p)
		if won && f.EnemyRewardID > 0 {
			grantItem(deps, user, f.EnemyRewardID, maxInt(1, f.EnemyRewardCt))
		}
		break
	}
}

func calculateExpGain(enemyID int, enemyLevel int, isWild bool) int {
	if enemyLevel <= 0 {
		enemyLevel = 1
	}
	baseExp := 50
	if base := LoadPetDB().pets[enemyID]; base != nil && base.BaseExp > 0 {
		baseExp = base.BaseExp
	}
	exp := int(float64(baseExp*enemyLevel) / 7.0)
	if isWild {
		exp = int(float64(exp) * 0.8)
	}
	if exp < 1 {
		exp = 1
	}
	return exp
}

func grantItem(deps *Deps, user *User, itemID int, count int) bool {
	if deps == nil || deps.Store == nil || user == nil || itemID <= 0 || count <= 0 {
		return false
	}
	if user.Items == nil {
		user.Items = make(map[int]*ItemInfo)
	}
	if isUniqueItem(itemID) && user.Items[itemID] != nil {
		return false
	}
	info := user.Items[itemID]
	if info == nil {
		info = &ItemInfo{Count: 0, ExpireTime: defaultItemExpire}
		user.Items[itemID] = info
	}
	info.Count += count
	upsertItem(deps, user, itemID)
	return true
}

func findPetByCatchTime(user *User, catchTime uint32) *Pet {
	if user == nil || catchTime == 0 {
		return nil
	}
	for i := range user.Pets {
		if user.Pets[i].CatchTime == catchTime {
			return &user.Pets[i]
		}
	}
	return nil
}

func buildFightUserInfo(userID uint32, nick string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, userID)
	protocol.WriteFixedString(buf, nick, 16)
	return buf.Bytes()
}

func buildSimpleFightPetInfo(petID int, level int, hp int, maxHP int, catchTime int, skills []int, catchMap int) []byte {
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
	if catchMap <= 0 {
		catchMap = 301
	}
	binary.Write(buf, binary.BigEndian, uint32(catchMap))
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

func buildAttackValue(userID uint32, skillID uint32, atkTimes uint32, lostHP uint32, gainHP int, remainHP int, maxHP int, state uint32, isCrit uint32, petType uint32, battleLv stageModifiers, status map[int]int) []byte {
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
	writeStatus(buf, status)
	writeBattleLv(buf, battleLv)
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, petType)
	return buf.Bytes()
}

func writeBattleLv(buf *bytes.Buffer, battleLv stageModifiers) {
	values := []int{
		battleLv.Atk,
		battleLv.Def,
		battleLv.SpA,
		battleLv.SpD,
		battleLv.Spd,
		battleLv.Acc,
	}
	for _, v := range values {
		if v > 127 {
			v = 127
		} else if v < -128 {
			v = -128
		}
		_ = buf.WriteByte(byte(int8(v)))
	}
}

func writeStatus(buf *bytes.Buffer, status map[int]int) {
	values := make([]byte, 20)
	if status != nil {
		for idx, turns := range status {
			if idx < 0 || idx >= len(values) {
				continue
			}
			if turns < 0 {
				turns = 0
			}
			if turns > 255 {
				turns = 255
			}
			values[idx] = byte(turns)
		}
	}
	buf.Write(values)
}

func sendFightOver(ctx *gateway.Context, winner uint32, reason uint32) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, reason)
	binary.Write(buf, binary.BigEndian, winner)
	buf.Write(make([]byte, 20))
	ctx.Server.SendResponse(ctx.Conn, 2506, ctx.UserID, buf.Bytes())
}

type fightPetSnapshot struct {
	ID        uint32
	Level     uint32
	DV        uint32
	CatchTime uint32
	Skills    []int
	Stats     petStats
	Type      int
	CurrentHP int
}

type attackResult struct {
	UserID   uint32
	SkillID  uint32
	AtkTimes uint32
	LostHP   uint32
	GainHP   int
	RemainHP int
	MaxHP    int
	State    uint32
	IsCrit   uint32
	PetType  uint32
	Stage    stageModifiers
}

type stageModifiers struct {
	Atk int
	Def int
	SpA int
	SpD int
	Spd int
	Acc int
	Eva int
}

type fightStateChange struct {
	Atk int
	Def int
	SpA int
	SpD int
	Spd int
	Acc int
	Eva int
}

const (
	statusParalysis = 0
	statusPoison    = 1
	statusBurn      = 2
	statusFreeze    = 5
	statusFear      = 6
	statusFatigue   = 7
	statusSleep     = 8
	statusPetrify   = 9
	statusConfuse   = 10
	statusBleed     = 16
)

const (
	effectSelfStage   = 4
	effectStageChange = 5
	effectDrain       = 1
	effectRecoil      = 6
	effectMultiHit    = 31
	effectFatigue     = 20
	effectMercy       = 8
	effectHpRatio     = 34
	effectPunishment  = 35
	effectParalysis   = 10
	effectBurn        = 12
	effectBind        = 11
	effectFear        = 15
	effectBleed       = 16
)

var stageStatIndex = map[int]fightStateChange{
	0: {Atk: 1},
	1: {Def: 1},
	2: {SpA: 1},
	3: {SpD: 1},
	4: {Spd: 1},
	5: {Acc: 1},
}

var stageMultipliers = map[int]float64{
	-6: 2.0 / 8.0,
	-5: 2.0 / 7.0,
	-4: 2.0 / 6.0,
	-3: 2.0 / 5.0,
	-2: 2.0 / 4.0,
	-1: 2.0 / 3.0,
	0:  1.0,
	1:  3.0 / 2.0,
	2:  4.0 / 2.0,
	3:  5.0 / 2.0,
	4:  6.0 / 2.0,
	5:  7.0 / 2.0,
	6:  8.0 / 2.0,
}

var accuracyMultipliers = map[int]float64{
	-6: 3.0 / 9.0,
	-5: 3.0 / 8.0,
	-4: 3.0 / 7.0,
	-3: 3.0 / 6.0,
	-2: 3.0 / 5.0,
	-1: 3.0 / 4.0,
	0:  1.0,
	1:  4.0 / 3.0,
	2:  5.0 / 3.0,
	3:  6.0 / 3.0,
	4:  7.0 / 3.0,
	5:  8.0 / 3.0,
	6:  9.0 / 3.0,
}

func clampStage(stage int) int {
	if stage > 6 {
		return 6
	}
	if stage < -6 {
		return -6
	}
	return stage
}

func applyStageModifier(base int, stage int) int {
	stage = clampStage(stage)
	mul := stageMultipliers[stage]
	if mul == 0 {
		mul = 1
	}
	return int(float64(base) * mul)
}

func calculateAccuracy(baseAccuracy int, attackerAcc int, defenderEva int) float64 {
	if baseAccuracy <= 0 {
		baseAccuracy = 100
	}
	netStage := clampStage(attackerAcc - defenderEva)
	mul := accuracyMultipliers[netStage]
	if mul == 0 {
		mul = 1
	}
	return float64(baseAccuracy) * mul
}

func checkHit(baseAccuracy int, attackerAcc int, defenderEva int) bool {
	acc := calculateAccuracy(baseAccuracy, attackerAcc, defenderEva)
	return rand.Float64()*100 < acc
}

func resolveUserFightPet(user *User, catchTime uint32, petID uint32) fightPetSnapshot {
	var picked *Pet
	if user != nil {
		if catchTime != 0 {
			for i := range user.Pets {
				if user.Pets[i].CatchTime == catchTime {
					picked = &user.Pets[i]
					break
				}
			}
		}
		if picked == nil && petID != 0 {
			for i := range user.Pets {
				if user.Pets[i].ID == petID {
					picked = &user.Pets[i]
					break
				}
			}
		}
		if picked == nil && len(user.Pets) > 0 {
			picked = &user.Pets[0]
		}
	}

	id := petID
	level := uint32(5)
	dv := uint32(31)
	currentHP := 0
	skills := []int{}
	if picked != nil {
		id = picked.ID
		level = picked.Level
		dv = picked.DV
		catchTime = picked.CatchTime
		skills = append([]int{}, picked.Skills...)
		currentHP = picked.HP
	}
	if id == 0 {
		id = 7
	}

	base := LoadPetDB().pets[int(id)]
	stats := getStats(base, int(level), int(dv), evSet{})
	skills = normalizeSkillList(skills, base, int(level))
	if currentHP <= 0 {
		currentHP = stats.MaxHP
	}
	if currentHP > stats.MaxHP {
		currentHP = stats.MaxHP
	}
	typ := 0
	if base != nil {
		typ = base.Type
	}
	return fightPetSnapshot{
		ID:        id,
		Level:     level,
		DV:        dv,
		CatchTime: catchTime,
		Skills:    skills,
		Stats:     stats,
		Type:      typ,
		CurrentHP: currentHP,
	}
}

func resolveEnemyFightPet(petID int, level int) fightPetSnapshot {
	if petID == 0 {
		petID = noviceBossID
	}
	if level <= 0 {
		level = 1
	}
	base := LoadPetDB().pets[petID]
	stats := getStats(base, level, 15, evSet{})
	skills := normalizeSkillList(nil, base, level)
	typ := 0
	if base != nil {
		typ = base.Type
	}
	return fightPetSnapshot{
		ID:        uint32(petID),
		Level:     uint32(level),
		DV:        15,
		CatchTime: 0,
		Skills:    skills,
		Stats:     stats,
		Type:      typ,
		CurrentHP: stats.HP,
	}
}

func normalizeSkillList(skills []int, base *PetBase, level int) []int {
	if len(skills) == 0 {
		return getSkillsForLevel(base, level)
	}
	if len(skills) > 4 {
		skills = skills[len(skills)-4:]
	}
	for len(skills) < 4 {
		skills = append(skills, 0)
	}
	return skills
}

func ensureFightStats(user *User, f *FightState) {
	if f == nil {
		return
	}
	if f.PlayerStats.MaxHP == 0 || len(f.PlayerSkills) == 0 {
		needHP := f.PlayerHP == 0 && f.PlayerMaxHP == 0
		player := resolveUserFightPet(user, f.PlayerCatch, f.PlayerPetID)
		f.PlayerPetID = player.ID
		f.PlayerLevel = player.Level
		f.PlayerCatch = player.CatchTime
		f.PlayerSkills = player.Skills
		f.PlayerStats = player.Stats
		f.PlayerType = player.Type
		if f.PlayerMaxHP == 0 {
			f.PlayerMaxHP = player.Stats.MaxHP
		}
		if needHP {
			f.PlayerHP = player.CurrentHP
		}
	}
	if f.EnemyStats.MaxHP == 0 || len(f.EnemySkills) == 0 {
		needHP := f.EnemyHP == 0 && f.EnemyMaxHP == 0
		enemy := resolveEnemyFightPet(int(f.EnemyPetID), int(f.EnemyLevel))
		f.EnemyPetID = enemy.ID
		f.EnemyLevel = enemy.Level
		f.EnemyCatch = enemy.CatchTime
		f.EnemySkills = enemy.Skills
		f.EnemyStats = enemy.Stats
		f.EnemyType = enemy.Type
		if f.EnemyMaxHP == 0 {
			f.EnemyMaxHP = enemy.Stats.MaxHP
		}
		if needHP {
			f.EnemyHP = enemy.CurrentHP
		}
	}
}

func ensureFightStatus(f *FightState) {
	if f == nil {
		return
	}
	if f.PlayerStatus == nil {
		f.PlayerStatus = make(map[int]int)
	}
	if f.EnemyStatus == nil {
		f.EnemyStatus = make(map[int]int)
	}
}

func selectPlayerSkill(skills []int, requested int) int {
	if requested > 0 && containsSkill(skills, requested) {
		return requested
	}
	for _, sid := range skills {
		if sid > 0 {
			return sid
		}
	}
	return 0
}

func selectRandomSkill(skills []int) int {
	valid := make([]int, 0, len(skills))
	for _, sid := range skills {
		if sid > 0 {
			valid = append(valid, sid)
		}
	}
	if len(valid) == 0 {
		return 0
	}
	return valid[rand.Intn(len(valid))]
}

func containsSkill(skills []int, target int) bool {
	for _, sid := range skills {
		if sid == target {
			return true
		}
	}
	return false
}

func getSkillPriority(skillID int) int {
	if info := getSkillInfo(skillID); info != nil {
		return info.Priority
	}
	return 0
}

func isPlayerFirst(playerSkill int, enemySkill int, playerSpeed int, enemySpeed int) bool {
	playerPriority := getSkillPriority(playerSkill)
	enemyPriority := getSkillPriority(enemySkill)
	if playerPriority != enemyPriority {
		return playerPriority > enemyPriority
	}
	if playerSpeed != enemySpeed {
		return playerSpeed > enemySpeed
	}
	return rand.Intn(2) == 0
}

func executeAttack(ctx *gateway.Context, f *FightState, player bool, skillID int) attackResult {
	if f == nil {
		return attackResult{}
	}
	ensureFightStatus(f)
	if player {
		if isActionBlockedByStatus(f.PlayerStatus) {
			decreaseStatusDurations(f.PlayerStatus)
			return attackResult{
				UserID:   ctx.UserID,
				SkillID:  0,
				AtkTimes: 0,
				LostHP:   0,
				GainHP:   0,
				RemainHP: f.PlayerHP,
				MaxHP:    f.PlayerMaxHP,
				State:    1,
				IsCrit:   0,
				PetType:  uint32(f.PlayerType),
				Stage:    f.PlayerStage,
			}
		}
		if f.PlayerFatigue > 0 {
			f.PlayerFatigue--
			return attackResult{
				UserID:   ctx.UserID,
				SkillID:  0,
				AtkTimes: 0,
				LostHP:   0,
				GainHP:   0,
				RemainHP: f.PlayerHP,
				MaxHP:    f.PlayerMaxHP,
				State:    1,
				IsCrit:   0,
				PetType:  uint32(f.PlayerType),
				Stage:    f.PlayerStage,
			}
		}
	} else {
		if isActionBlockedByStatus(f.EnemyStatus) {
			decreaseStatusDurations(f.EnemyStatus)
			return attackResult{
				UserID:   0,
				SkillID:  0,
				AtkTimes: 0,
				LostHP:   0,
				GainHP:   0,
				RemainHP: f.EnemyHP,
				MaxHP:    f.EnemyMaxHP,
				State:    1,
				IsCrit:   0,
				PetType:  uint32(f.EnemyType),
				Stage:    f.EnemyStage,
			}
		}
		if f.EnemyFatigue > 0 {
			f.EnemyFatigue--
			return attackResult{
				UserID:   0,
				SkillID:  0,
				AtkTimes: 0,
				LostHP:   0,
				GainHP:   0,
				RemainHP: f.EnemyHP,
				MaxHP:    f.EnemyMaxHP,
				State:    1,
				IsCrit:   0,
				PetType:  uint32(f.EnemyType),
				Stage:    f.EnemyStage,
			}
		}
	}
	if skillID < 0 {
		skillID = 0
	}
	attackerID := uint32(0)
	atkStats := f.EnemyStats
	defStats := f.PlayerStats
	atkLevel := int(f.EnemyLevel)
	atkHP := f.EnemyHP
	atkMaxHP := f.EnemyMaxHP
	atkType := f.EnemyType
	defType := f.PlayerType
	atkStage := f.EnemyStage
	defStage := f.PlayerStage
	if player {
		attackerID = ctx.UserID
		atkStats = f.PlayerStats
		defStats = f.EnemyStats
		atkLevel = int(f.PlayerLevel)
		atkHP = f.PlayerHP
		atkMaxHP = f.PlayerMaxHP
		atkType = f.PlayerType
		defType = f.EnemyType
		atkStage = f.PlayerStage
		defStage = f.EnemyStage
	}

	defHP := f.PlayerHP
	if player {
		defHP = f.EnemyHP
	}
	damage, state, crit, hitCount := calcDamage(atkStats, defStats, atkLevel, skillID, atkType, defType, defHP, atkStage, defStage)
	info := getSkillInfo(skillID)
	gainHP := 0
	recoilDamage := 0
	if state == 0 && hitCount > 0 {
		if info != nil && (info.SideEffect == effectSelfStage || info.SideEffect == effectStageChange) {
			applyStageEffects(f, player, info)
			if player {
				atkStage = f.PlayerStage
			} else {
				atkStage = f.EnemyStage
			}
		}
		if info != nil && info.SideEffect == effectMercy {
			if defHP-damage < 1 {
				damage = maxInt(0, defHP-1)
			}
		}
		if info != nil && info.SideEffect == effectDrain {
			gainHP = damage / 2
		}
		if info != nil && info.SideEffect == effectRecoil {
			divisor := 4
			if info.SideEffectArg != "" {
				args := parseEffectArgs(info.SideEffectArg)
				if len(args) >= 1 && args[0] > 0 {
					divisor = args[0]
				}
			}
			recoilDamage = damage / divisor
		}
	}
	if info != nil && state == 0 && hitCount > 0 {
		applyStatusEffect(f, player, info)
	}
	if player {
		f.EnemyHP = maxInt(0, f.EnemyHP-damage)
		if gainHP > 0 {
			f.PlayerHP = minInt(f.PlayerMaxHP, f.PlayerHP+gainHP)
		}
		if recoilDamage > 0 {
			f.PlayerHP = maxInt(0, f.PlayerHP-recoilDamage)
		}
	} else {
		f.PlayerHP = maxInt(0, f.PlayerHP-damage)
		if gainHP > 0 {
			f.EnemyHP = minInt(f.EnemyMaxHP, f.EnemyHP+gainHP)
		}
		if recoilDamage > 0 {
			f.EnemyHP = maxInt(0, f.EnemyHP-recoilDamage)
		}
	}
	if info != nil && info.SideEffect == effectFatigue && state == 0 && hitCount > 0 {
		turns := 1
		if info.SideEffectArg != "" {
			args := parseEffectArgs(info.SideEffectArg)
			if len(args) >= 2 && args[1] > 0 {
				turns = args[1]
			}
		}
		if player {
			f.PlayerFatigue = maxInt(f.PlayerFatigue, turns)
		} else {
			f.EnemyFatigue = maxInt(f.EnemyFatigue, turns)
		}
	}
	if state == 0 && hitCount > 0 {
		if player {
			applyEndTurnStatusDamage(f.PlayerStatus, &f.PlayerHP, f.PlayerMaxHP)
		} else {
			applyEndTurnStatusDamage(f.EnemyStatus, &f.EnemyHP, f.EnemyMaxHP)
		}
	}

	remainHP := atkHP
	if player {
		remainHP = f.PlayerHP
	} else {
		remainHP = f.EnemyHP
	}

	return attackResult{
		UserID:   attackerID,
		SkillID:  uint32(skillID),
		AtkTimes: hitCount,
		LostHP:   uint32(damage),
		GainHP:   gainHP,
		RemainHP: remainHP,
		MaxHP:    atkMaxHP,
		State:    state,
		IsCrit:   boolToUint32(crit),
		PetType:  uint32(atkType),
		Stage:    atkStage,
	}
}

func placeholderAttack(player bool, f *FightState) attackResult {
	if player {
		return attackResult{
			UserID:   f.UserID,
			SkillID:  0,
			AtkTimes: 0,
			LostHP:   0,
			GainHP:   0,
			RemainHP: f.PlayerHP,
			MaxHP:    f.PlayerMaxHP,
			State:    0,
			IsCrit:   0,
			PetType:  uint32(f.PlayerType),
			Stage:    f.PlayerStage,
		}
	}
	return attackResult{
		UserID:   0,
		SkillID:  0,
		AtkTimes: 0,
		LostHP:   0,
		GainHP:   0,
		RemainHP: f.EnemyHP,
		MaxHP:    f.EnemyMaxHP,
		State:    0,
		IsCrit:   0,
		PetType:  uint32(f.EnemyType),
		Stage:    f.EnemyStage,
	}
}

func calcDamage(atk petStats, def petStats, level int, skillID int, atkType int, defType int, defHP int, atkStage stageModifiers, defStage stageModifiers) (int, uint32, bool, uint32) {
	info := getSkillInfo(skillID)
	if info == nil {
		return 0, 0, false, 1
	}
	if level <= 0 {
		level = 1
	}
	if !info.MustHit {
		if !checkHit(info.Accuracy, atkStage.Acc, defStage.Eva) {
			return 0, 1, false, 0
		}
	}
	if info.Power <= 0 || info.Category == 4 {
		return 0, 0, false, 1
	}
	if info.SideEffect == effectHpRatio {
		ratio := 50
		if info.SideEffectArg != "" {
			args := parseEffectArgs(info.SideEffectArg)
			if len(args) > 0 {
				ratio = args[0]
			}
		} else if info.Power > 0 {
			ratio = info.Power
		}
		if ratio < 1 {
			ratio = 1
		}
		damage := defHP * ratio / 100
		if damage < 1 && defHP > 0 {
			damage = 1
		}
		return damage, 0, false, 1
	}
	atkVal := atk.Attack
	defVal := def.Defence
	if info.Category == 2 {
		atkVal = atk.SA
		defVal = def.SD
	}
	if info.Category == 2 {
		atkVal = applyStageModifier(atkVal, atkStage.SpA)
		defVal = applyStageModifier(defVal, defStage.SpD)
	} else {
		atkVal = applyStageModifier(atkVal, atkStage.Atk)
		defVal = applyStageModifier(defVal, defStage.Def)
	}
	if atkVal <= 0 {
		atkVal = 1
	}
	if defVal <= 0 {
		defVal = 1
	}
	power := info.Power
	if info.SideEffect == effectPunishment {
		power += sumPositiveStages(defStage) * 20
	}
	base := (float64(level)*0.4 + 2) * float64(power) * float64(atkVal) / float64(defVal) / 50.0
	base += 2
	effectiveness := elementMultiplier(info.Type, defType)
	stab := 1.0
	if info.Type > 0 && info.Type == atkType {
		stab = 1.5
	}
	crit := rand.Intn(16) == 0
	critMod := 1.0
	if crit {
		critMod = 1.5
	}
	randomMod := float64(85+rand.Intn(16)) / 100.0

	hitCount := uint32(1)
	if info.SideEffect == effectMultiHit {
		minHits := 2
		maxHits := 5
		if info.SideEffectArg != "" {
			args := parseEffectArgs(info.SideEffectArg)
			if len(args) >= 2 {
				minHits = args[0]
				maxHits = args[1]
			}
		}
		if maxHits < minHits {
			maxHits = minHits
		}
		if minHits < 1 {
			minHits = 1
		}
		hitCount = uint32(minHits + rand.Intn(maxHits-minHits+1))
	}

	final := base * effectiveness * stab * critMod * randomMod
	damage := int(final) * int(hitCount)
	if effectiveness > 0 && damage < 1 {
		damage = 1
	}
	if effectiveness == 0 {
		damage = 0
	}
	return damage, 0, crit, hitCount
}

func applyStageEffects(f *FightState, player bool, info *SkillInfo) {
	if info == nil {
		return
	}
	if info.SideEffect != effectSelfStage && info.SideEffect != effectStageChange {
		return
	}
	args := parseEffectArgs(info.SideEffectArg)
	if len(args) < 3 {
		return
	}
	stat := args[0]
	chance := args[1]
	stages := args[2]
	if chance <= 0 {
		chance = 100
	}
	if rand.Intn(100)+1 > chance {
		return
	}
	baseChange := stageStatIndex[stat]
	if baseChange == (fightStateChange{}) {
		return
	}
	change := fightStateChange{
		Atk: baseChange.Atk * stages,
		Def: baseChange.Def * stages,
		SpA: baseChange.SpA * stages,
		SpD: baseChange.SpD * stages,
		Spd: baseChange.Spd * stages,
		Acc: baseChange.Acc * stages,
	}
	if info.SideEffect == effectStageChange && stages < 0 {
		if player {
			f.EnemyStage = clampStageChange(f.EnemyStage, change)
		} else {
			f.PlayerStage = clampStageChange(f.PlayerStage, change)
		}
		return
	}
	if player {
		f.PlayerStage = clampStageChange(f.PlayerStage, change)
	} else {
		f.EnemyStage = clampStageChange(f.EnemyStage, change)
	}
}

func parseEffectArgs(arg string) []int {
	if arg == "" {
		return nil
	}
	parts := strings.FieldsFunc(arg, func(r rune) bool {
		return r == ' ' || r == ',' || r == '\t'
	})
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		if v, err := strconv.Atoi(p); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func applyStatusEffect(f *FightState, player bool, info *SkillInfo) {
	if f == nil || info == nil {
		return
	}
	var statusID int
	var turns int
	switch info.SideEffect {
	case effectParalysis:
		statusID = statusParalysis
		turns = 2
	case effectBurn:
		statusID = statusBurn
		turns = 3
	case effectBind:
		statusID = statusPoison
		turns = 4
	case effectFear:
		statusID = statusFear
		turns = 2
	case effectBleed:
		statusID = statusBleed
		turns = 3
	default:
		return
	}
	chance := 100
	if info.SideEffectArg != "" {
		args := parseEffectArgs(info.SideEffectArg)
		if len(args) >= 1 && args[0] > 0 {
			chance = args[0]
		}
		if len(args) >= 2 && args[1] > 0 {
			turns = args[1]
		}
	}
	if chance < 100 && rand.Intn(100)+1 > chance {
		return
	}
	if player {
		if f.EnemyStatus == nil {
			f.EnemyStatus = make(map[int]int)
		}
		f.EnemyStatus[statusID] = maxInt(f.EnemyStatus[statusID], turns)
	} else {
		if f.PlayerStatus == nil {
			f.PlayerStatus = make(map[int]int)
		}
		f.PlayerStatus[statusID] = maxInt(f.PlayerStatus[statusID], turns)
	}
}

func isActionBlockedByStatus(status map[int]int) bool {
	if status == nil {
		return false
	}
	return status[statusSleep] > 0 || status[statusFreeze] > 0 || status[statusPetrify] > 0 || status[statusConfuse] > 0 || status[statusParalysis] > 0 || status[statusFear] > 0
}

func decreaseStatusDurations(status map[int]int) {
	if status == nil {
		return
	}
	for k, v := range status {
		if v <= 0 {
			continue
		}
		status[k] = v - 1
	}
}

func applyEndTurnStatusDamage(status map[int]int, hp *int, maxHP int) {
	if status == nil || hp == nil || maxHP <= 0 {
		return
	}
	damage := 0
	if status[statusPoison] > 0 {
		damage += maxHP / 8
		status[statusPoison]--
	}
	if status[statusBurn] > 0 {
		damage += maxHP / 16
		status[statusBurn]--
	}
	if status[statusBleed] > 0 {
		damage += maxHP / 8
		status[statusBleed]--
	}
	if damage > 0 {
		*hp = maxInt(0, *hp-damage)
	}
}

func clampStageChange(cur stageModifiers, change fightStateChange) stageModifiers {
	cur.Atk = clampStage(cur.Atk + change.Atk)
	cur.Def = clampStage(cur.Def + change.Def)
	cur.SpA = clampStage(cur.SpA + change.SpA)
	cur.SpD = clampStage(cur.SpD + change.SpD)
	cur.Spd = clampStage(cur.Spd + change.Spd)
	cur.Acc = clampStage(cur.Acc + change.Acc)
	cur.Eva = clampStage(cur.Eva + change.Eva)
	return cur
}

func boolToUint32(v bool) uint32 {
	if v {
		return 1
	}
	return 0
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func sumPositiveStages(stage stageModifiers) int {
	sum := 0
	for _, v := range []int{stage.Atk, stage.Def, stage.SpA, stage.SpD, stage.Spd, stage.Acc, stage.Eva} {
		if v > 0 {
			sum += v
		}
	}
	return sum
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
