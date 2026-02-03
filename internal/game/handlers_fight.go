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

	"go.uber.org/zap"
)

const noviceBossID = 13

func init() {
	rand.Seed(time.Now().UnixNano())
}

func registerFightHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2401, handleInviteToFight(state))
	s.Register(2402, handleInviteFightCancel(state))
	s.Register(2403, handleHandleFightInvite(state))
	s.Register(2411, handleChallengeBoss(state))
	s.Register(2404, handleReadyToFight(state))
	s.Register(2405, handleUseSkill(deps, state))
	s.Register(2406, handleUsePetItem(state))
	s.Register(2407, handleChangePet(state))
	s.Register(2408, handleFightNpcMonster(state))
	s.Register(2409, handleCatchMonster(deps, state))
	s.Register(2410, handleEscapeFight(deps, state))
	s.Register(2412, handleAttackBoss(state))
	s.Register(2413, handlePetKingJoin())
	s.Register(2427, handleNpcJoin())
	s.Register(2431, handleStartPetWar())
	s.Register(2441, handleLoadPercent())
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
		player.CatchTime = ensureCatchTime(player.CatchTime, player.ID)
		if user.CatchID == 0 {
			user.CatchID = player.CatchTime
		}
		if user.CurrentPetID == 0 {
			user.CurrentPetID = player.ID
		}
		player.CatchTime = ensureCatchTime(player.CatchTime, player.ID)
		if user.CatchID == 0 {
			user.CatchID = player.CatchTime
		}
		if user.CurrentPetID == 0 {
			user.CurrentPetID = player.ID
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
			PlayerDV:      player.DV,
			PlayerHP:      player.CurrentHP,
			PlayerMaxHP:   player.Stats.MaxHP,
			PlayerCatch:   player.CatchTime,
			PlayerSkills:  player.Skills,
			PlayerStats:   player.Stats,
			PlayerType:    player.Type,
			EnemyPetID:    enemy.ID,
			EnemyLevel:    enemy.Level,
			EnemyDV:       enemy.DV,
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
		buf.Write(buildSimpleFightPetInfo(int(player.ID), int(player.Level), player.CurrentHP, player.Stats.MaxHP, int(player.CatchTime), player.Skills, int(user.MapID), int(player.ID)))
		binary.Write(buf, binary.BigEndian, uint32(0))
		protocol.WriteFixedString(buf, "", 16)
		binary.Write(buf, binary.BigEndian, uint32(1))
		buf.Write(buildSimpleFightPetInfo(int(enemy.ID), int(enemy.Level), enemy.CurrentHP, enemy.Stats.MaxHP, int(enemy.CatchTime), enemy.Skills, 301, int(enemy.ID)))
		ctx.Server.SendResponse(ctx.Conn, 2503, ctx.UserID, buf.Bytes())
		// sendPveFightStart(ctx, user, user.Fight) // Removed: Should not send 2504 here, wait for 2404
	}
}

func handleReadyToFight(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Fight == nil {
			// handleChallengeBoss(state)(ctx) // Removed: Don't auto-challenge if state missing
			return
		}
		f := user.Fight
		if f == nil {
			return
		}
		if f.OpponentUserID != 0 {
			opp := state.GetOrCreateUser(f.OpponentUserID)
			body := buildNoteStartFightPvP(ctx.UserID, f, f.OpponentUserID, opp.Fight)
			if len(body) > 0 {
				ctx.Server.SendResponse(ctx.Conn, 2504, ctx.UserID, body)
			}
			return
		}
		ensureFightStats(user, f)
		ensureFightSkillPP(f)
		ensureFightStatus(f)
		user.InFight = true
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		buf.Write(buildFightPetInfo(ctx.UserID, f.PlayerPetID, f.PlayerCatch, f.PlayerHP, f.PlayerMaxHP, f.PlayerLevel, 0))
		buf.Write(buildFightPetInfo(0, f.EnemyPetID, f.EnemyCatch, f.EnemyHP, f.EnemyMaxHP, f.EnemyLevel, 1))
		ctx.Server.SendResponse(ctx.Conn, 2504, ctx.UserID, buf.Bytes())
		sendFightPetInfo(ctx, user, f)
	}
}

func handleUseSkill(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		reqSkillID := int(reader.ReadUint32BE())
		ctx.Server.SendResponse(ctx.Conn, 2405, ctx.UserID, []byte{})

		deps.Logger.Info("fight_use_skill", zap.Uint32("uid", ctx.UserID), zap.Int("skill_id", reqSkillID))

		user := state.GetOrCreateUser(ctx.UserID)
		f := user.Fight
		if f == nil {
			sendFightOver(ctx, 0, 0)
			return
		}
		ensureFightStats(user, f)
		ensureFightStatus(f)
		ensureFightSkillPP(f)
		if f.OpponentUserID != 0 {
			handleUseSkillPvP(ctx, deps, state, user, f, reqSkillID)
			return
		}

		f.Turn++
		playerSkill := pickSkillWithEncore(reqSkillID, f.PlayerSkills, f.PlayerSkillPP, &f.PlayerEncoreSkill, &f.PlayerEncoreTurns)
		enemySkill := pickEncoreSkill(&f.EnemyEncoreSkill, &f.EnemyEncoreTurns, f.EnemySkillPP)
		if enemySkill == 0 {
			enemySkill = selectAISkill(f)
		}
		if enemySkill == 0 {
			enemySkill = selectRandomSkillWithPP(f.EnemySkills, f.EnemySkillPP)
		}
		if playerSkill > 0 {
			if pp, maxPP, ok := consumeSkillPP(f.PlayerSkillPP, playerSkill); ok {
				sendSkillPPUpdate(ctx, playerSkill, pp, maxPP)
			}
		}
		if enemySkill > 0 {
			consumeSkillPP(f.EnemySkillPP, enemySkill)
		}
		f.PlayerLastSkill = playerSkill
		f.EnemyLastSkill = enemySkill

		applyTurnStatusDamage(f.PlayerStatus, &f.PlayerHP, f.PlayerMaxHP, &f.PlayerBoundTurns)
		applyTurnStatusDamage(f.EnemyStatus, &f.EnemyHP, f.EnemyMaxHP, &f.EnemyBoundTurns)

		playerSpeed := effectiveSpeed(f.PlayerStats.Speed, f.PlayerStage.Spd, f.PlayerStatus)
		enemySpeed := effectiveSpeed(f.EnemyStats.Speed, f.EnemyStage.Spd, f.EnemyStatus)
		playerFirst := isPlayerFirst(playerSkill, enemySkill, playerSpeed, enemySpeed)

		var first attackResult
		var second attackResult
		if f.PlayerHP <= 0 || f.EnemyHP <= 0 {
			first = placeholderAttack(true, f)
			second = placeholderAttack(false, f)
		} else {
			playerCanAct := canAct(f, true)
			enemyCanAct := canAct(f, false)
			if playerFirst {
				if playerCanAct {
					first = executeAttack(ctx, f, true, playerSkill, true)
				} else {
					first = cannotActAttack(ctx.UserID, true, f)
				}
				if f.EnemyHP > 0 {
					if enemyCanAct {
						second = executeAttack(ctx, f, false, enemySkill, false)
					} else {
						second = cannotActAttack(ctx.UserID, false, f)
					}
				} else {
					second = placeholderAttack(false, f)
				}
			} else {
				if enemyCanAct {
					first = executeAttack(ctx, f, false, enemySkill, true)
				} else {
					first = cannotActAttack(ctx.UserID, false, f)
				}
				if f.PlayerHP > 0 {
					if playerCanAct {
						second = executeAttack(ctx, f, true, playerSkill, false)
					} else {
						second = cannotActAttack(ctx.UserID, true, f)
					}
				} else {
					second = placeholderAttack(true, f)
				}
			}
		}

		oppID := f.OpponentUserID
		firstStatus := f.PlayerStatus
		if first.UserID == 0 {
			firstStatus = f.EnemyStatus
		}
		secondStatus := f.PlayerStatus
		if second.UserID == 0 {
			secondStatus = f.EnemyStatus
		}
		firstUserID := first.UserID
		if oppID != 0 && firstUserID == 0 {
			firstUserID = oppID
		}
		secondUserID := second.UserID
		if oppID != 0 && secondUserID == 0 {
			secondUserID = oppID
		}
		buf := new(bytes.Buffer)
		buf.Write(buildAttackValue(firstUserID, first.SkillID, first.AtkTimes, first.LostHP, first.GainHP, first.RemainHP, first.MaxHP, first.State, first.IsCrit, first.PetType, first.Stage, firstStatus))
		buf.Write(buildAttackValue(secondUserID, second.SkillID, second.AtkTimes, second.LostHP, second.GainHP, second.RemainHP, second.MaxHP, second.State, second.IsCrit, second.PetType, second.Stage, secondStatus))
		ctx.Server.SendResponse(ctx.Conn, 2505, ctx.UserID, buf.Bytes())
		if oppID != 0 {
			if conn, ok := state.GetConn(oppID); ok {
				ctx.Server.SendResponse(conn, 2505, oppID, buf.Bytes())
			}
			syncPvPFightState(state, user, f)
		}

		if f.EnemyHP == 0 || f.PlayerHP == 0 {
			winner := uint32(0)
			if f.EnemyHP == 0 {
				winner = ctx.UserID
			} else if oppID != 0 {
				winner = oppID
			}
			if oppID != 0 {
				updateFightHP(deps, user, f)
				sendNoteUpdateProp(ctx, user, f.PlayerCatch)
				var oppCatch uint32
				if opp := state.GetOrCreateUser(oppID); opp != nil {
					if opp.Fight != nil {
						oppCatch = opp.Fight.PlayerCatch
						updateFightHP(deps, opp, opp.Fight)
					}
				}
				sendFightOver(ctx, winner, 0)
				if conn, ok := state.GetConn(oppID); ok {
					ctx.Server.SendResponse(conn, 2506, oppID, buildFightOverBody(0, winner))
					if opp := state.GetOrCreateUser(oppID); opp != nil {
						if body := buildNoteUpdatePropBody(opp, oppCatch); len(body) > 0 {
							ctx.Server.SendResponse(conn, 2508, oppID, body)
						}
					}
				}
				user.Fight = nil
				user.InFight = false
				if opp := state.GetOrCreateUser(oppID); opp != nil {
					opp.Fight = nil
					opp.InFight = false
				}
				return
			}
			learned := updateFightResult(deps, user, f, winner == ctx.UserID)
			sendFightOver(ctx, winner, 0)
			sendNoteUpdateProp(ctx, user, f.PlayerCatch)
			sendNoteUpdateSkill(ctx, learned)
			user.Fight = nil
			user.InFight = false
		}
	}
}

func handleUsePetItem(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		itemID := reader.ReadUint32BE()
		heal := 50
		hp := 100
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Fight != nil {
			if user.Fight.PlayerMaxHP > 0 {
				user.Fight.PlayerHP = minInt(user.Fight.PlayerMaxHP, user.Fight.PlayerHP+heal)
				hp = user.Fight.PlayerHP
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, itemID)
		binary.Write(buf, binary.BigEndian, uint32(hp))
		binary.Write(buf, binary.BigEndian, uint32(heal))
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
			user.Fight.PlayerDV = pet.DV
			user.Fight.PlayerCatch = pet.CatchTime
			user.Fight.PlayerSkills = pet.Skills
			user.Fight.PlayerStats = pet.Stats
			user.Fight.PlayerType = pet.Type
			user.Fight.PlayerHP = pet.CurrentHP
			user.Fight.PlayerMaxHP = pet.Stats.MaxHP
			user.Fight.PlayerSkillPP = nil
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
		if user.Fight == nil {
			return
		}
		if user.Fight.OpponentUserID != 0 {
			return
		}
		bossID := user.Fight.EnemyPetID
		catchTime := uint32(time.Now().Unix())
		if user.Fight.EnemyCatch > 0 {
			catchTime = user.Fight.EnemyCatch
		}
		level := user.Fight.EnemyLevel
		if level == 0 {
			level = 1
		}
		dv := uint32(1 + rand.Intn(31))
		base := LoadPetDB().pets[int(bossID)]
		stats := getStats(base, int(level), int(dv), evSet{})
		newPet := Pet{
			ID:        bossID,
			CatchTime: catchTime,
			Level:     level,
			DV:        dv,
			Exp:       0,
			HP:        stats.MaxHP,
		}
		user.Pets = append(user.Pets, newPet)
		upsertPet(deps, user, newPet)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, catchTime)
		binary.Write(buf, binary.BigEndian, bossID)
		ctx.Server.SendResponse(ctx.Conn, 2409, ctx.UserID, buf.Bytes())

		learned := updateFightResult(deps, user, user.Fight, true)
		sendFightOver(ctx, ctx.UserID, 0)
		sendNoteUpdateProp(ctx, user, user.Fight.PlayerCatch)
		sendNoteUpdateSkill(ctx, learned)
		user.Fight = nil
		user.InFight = false
	}
}

func handleEscapeFight(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Fight != nil {
			if user.Fight.OpponentUserID == 0 {
				updateFightResult(deps, user, user.Fight, false)
			} else {
				updateFightHP(deps, user, user.Fight)
				if opp := state.GetOrCreateUser(user.Fight.OpponentUserID); opp != nil && opp.Fight != nil {
					updateFightHP(deps, opp, opp.Fight)
					opp.Fight = nil
					opp.InFight = false
					if conn, ok := state.GetConn(opp.ID); ok {
						ctx.Server.SendResponse(conn, 2506, opp.ID, buildFightOverBody(0, 0))
					}
				}
			}
		}
		user.Fight = nil
		user.InFight = false
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(1))
		ctx.Server.SendResponse(ctx.Conn, 2410, ctx.UserID, buf.Bytes())
	}
}

func handleUseSkillPvP(ctx *gateway.Context, deps *Deps, state *State, user *User, f *FightState, reqSkillID int) {
	if ctx == nil || state == nil || user == nil || f == nil {
		return
	}
	oppID := f.OpponentUserID
	if oppID == 0 {
		return
	}
	ensureFightStats(user, f)
	ensureFightStatus(f)
	ensureFightSkillPP(f)

	playerSkill := pickSkillWithEncore(reqSkillID, f.PlayerSkills, f.PlayerSkillPP, &f.PlayerEncoreSkill, &f.PlayerEncoreTurns)
	if playerSkill == 0 {
		playerSkill = selectRandomSkillWithPP(f.PlayerSkills, f.PlayerSkillPP)
	}
	if playerSkill > 0 {
		if pp, maxPP, ok := consumeSkillPP(f.PlayerSkillPP, playerSkill); ok {
			sendSkillPPUpdate(ctx, playerSkill, pp, maxPP)
		}
	}

	state.fightMu.Lock()
	defer state.fightMu.Unlock()

	f.PendingSkillID = playerSkill
	opp := state.GetOrCreateUser(oppID)
	if opp == nil || opp.Fight == nil || opp.Fight.OpponentUserID != ctx.UserID {
		f.PendingSkillID = 0
		sendFightOver(ctx, 0, 0)
		return
	}
	if opp.Fight.PendingSkillID == 0 {
		return
	}
	enemySkill := opp.Fight.PendingSkillID
	f.PendingSkillID = 0
	opp.Fight.PendingSkillID = 0

	if enemySkill > 0 {
		consumeSkillPP(f.EnemySkillPP, enemySkill)
	}
	f.PlayerLastSkill = playerSkill
	f.EnemyLastSkill = enemySkill
	opp.Fight.PlayerLastSkill = enemySkill
	opp.Fight.EnemyLastSkill = playerSkill
	f.Turn++
	opp.Fight.Turn++

	applyTurnStatusDamage(f.PlayerStatus, &f.PlayerHP, f.PlayerMaxHP, &f.PlayerBoundTurns)
	applyTurnStatusDamage(f.EnemyStatus, &f.EnemyHP, f.EnemyMaxHP, &f.EnemyBoundTurns)

	playerSpeed := effectiveSpeed(f.PlayerStats.Speed, f.PlayerStage.Spd, f.PlayerStatus)
	enemySpeed := effectiveSpeed(f.EnemyStats.Speed, f.EnemyStage.Spd, f.EnemyStatus)
	playerFirst := isPlayerFirst(playerSkill, enemySkill, playerSpeed, enemySpeed)

	var first attackResult
	var second attackResult
	if f.PlayerHP <= 0 || f.EnemyHP <= 0 {
		first = placeholderAttack(true, f)
		second = placeholderAttack(false, f)
	} else {
		playerCanAct := canAct(f, true)
		enemyCanAct := canAct(f, false)
		if playerFirst {
			if playerCanAct {
				first = executeAttack(ctx, f, true, playerSkill, true)
			} else {
				first = cannotActAttack(ctx.UserID, true, f)
			}
			if f.EnemyHP > 0 {
				if enemyCanAct {
					second = executeAttack(ctx, f, false, enemySkill, false)
				} else {
					second = cannotActAttack(ctx.UserID, false, f)
				}
			} else {
				second = placeholderAttack(false, f)
			}
		} else {
			if enemyCanAct {
				first = executeAttack(ctx, f, false, enemySkill, true)
			} else {
				first = cannotActAttack(ctx.UserID, false, f)
			}
			if f.PlayerHP > 0 {
				if playerCanAct {
					second = executeAttack(ctx, f, true, playerSkill, false)
				} else {
					second = cannotActAttack(ctx.UserID, true, f)
				}
			} else {
				second = placeholderAttack(true, f)
			}
		}
	}

	firstStatus := f.PlayerStatus
	if first.UserID == 0 {
		firstStatus = f.EnemyStatus
	}
	secondStatus := f.PlayerStatus
	if second.UserID == 0 {
		secondStatus = f.EnemyStatus
	}
	firstUserID := first.UserID
	if firstUserID == 0 {
		firstUserID = oppID
	}
	secondUserID := second.UserID
	if secondUserID == 0 {
		secondUserID = oppID
	}

	buf := new(bytes.Buffer)
	buf.Write(buildAttackValue(firstUserID, first.SkillID, first.AtkTimes, first.LostHP, first.GainHP, first.RemainHP, first.MaxHP, first.State, first.IsCrit, first.PetType, first.Stage, firstStatus))
	buf.Write(buildAttackValue(secondUserID, second.SkillID, second.AtkTimes, second.LostHP, second.GainHP, second.RemainHP, second.MaxHP, second.State, second.IsCrit, second.PetType, second.Stage, secondStatus))
	ctx.Server.SendResponse(ctx.Conn, 2505, ctx.UserID, buf.Bytes())
	if conn, ok := state.GetConn(oppID); ok {
		ctx.Server.SendResponse(conn, 2505, oppID, buf.Bytes())
	}
	syncPvPFightState(state, user, f)

	if f.EnemyHP == 0 || f.PlayerHP == 0 {
		winner := uint32(0)
		if f.EnemyHP == 0 {
			winner = ctx.UserID
		} else {
			winner = oppID
		}
		updateFightHP(deps, user, f)
		sendNoteUpdateProp(ctx, user, f.PlayerCatch)
		var oppCatch uint32
		if opp.Fight != nil {
			oppCatch = opp.Fight.PlayerCatch
			updateFightHP(deps, opp, opp.Fight)
		}
		sendFightOver(ctx, winner, 0)
		if conn, ok := state.GetConn(oppID); ok {
			ctx.Server.SendResponse(conn, 2506, oppID, buildFightOverBody(0, winner))
			if body := buildNoteUpdatePropBody(opp, oppCatch); len(body) > 0 {
				ctx.Server.SendResponse(conn, 2508, oppID, body)
			}
		}
		user.Fight = nil
		user.InFight = false
		opp.Fight = nil
		opp.InFight = false
	}
}

func handleInviteToFight(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		mode := reader.ReadUint32BE()

		user := state.GetOrCreateUser(ctx.UserID)
		user.PendingInviteTo = targetID
		user.PendingInviteMode = mode

		ack := new(bytes.Buffer)
		binary.Write(ack, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2401, ctx.UserID, ack.Bytes())

		if targetID == 0 {
			return
		}
		if conn, ok := state.GetConn(targetID); ok {
			note := new(bytes.Buffer)
			binary.Write(note, binary.BigEndian, ctx.UserID)
			protocol.WriteFixedString(note, pickNick(user, ctx.UserID), 16)
			binary.Write(note, binary.BigEndian, mode)
			ctx.Server.SendResponse(conn, 2501, ctx.UserID, note.Bytes())
		}
	}
}

func handleInviteFightCancel(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if user.PendingInviteTo == targetID {
			user.PendingInviteTo = 0
			user.PendingInviteMode = 0
		}
		ctx.Server.SendResponse(ctx.Conn, 2402, ctx.UserID, []byte{})
	}
}

func handleHandleFightInvite(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		inviterID := reader.ReadUint32BE()
		result := reader.ReadUint32BE()
		_ = reader.ReadUint32BE() // mode

		responder := state.GetOrCreateUser(ctx.UserID)
		ack := new(bytes.Buffer)
		binary.Write(ack, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2403, ctx.UserID, ack.Bytes())

		if inviterID == 0 {
			return
		}
		if conn, ok := state.GetConn(inviterID); ok {
			note := new(bytes.Buffer)
			binary.Write(note, binary.BigEndian, ctx.UserID)
			protocol.WriteFixedString(note, pickNick(responder, ctx.UserID), 16)
			binary.Write(note, binary.BigEndian, result)
			ctx.Server.SendResponse(conn, 2502, ctx.UserID, note.Bytes())
		}

		if result != 1 {
			return
		}

		inviter := state.GetOrCreateUser(inviterID)
		initPvPFightState(state, inviterID, ctx.UserID, inviter, responder)

		bodyInviter, bodyResponder := buildNoteReadyToFightPvP(inviterID, ctx.UserID, inviter, responder)
		if conn, ok := state.GetConn(inviterID); ok {
			ctx.Server.SendResponse(conn, 2503, inviterID, bodyInviter)
			body := buildNoteStartFightPvP(inviterID, inviter.Fight, ctx.UserID, responder.Fight)
			if len(body) > 0 {
				ctx.Server.SendResponse(conn, 2504, inviterID, body)
			}
		}
		ctx.Server.SendResponse(ctx.Conn, 2503, ctx.UserID, bodyResponder)
		body := buildNoteStartFightPvP(ctx.UserID, responder.Fight, inviterID, inviter.Fight)
		if len(body) > 0 {
			ctx.Server.SendResponse(ctx.Conn, 2504, ctx.UserID, body)
		}
	}
}

func handleFightNpcMonster(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		slot := int(reader.ReadUint32BE())
		user := state.GetOrCreateUser(ctx.UserID)
		mapID := user.MapID
		if mapID == 0 {
			mapID = 1
		}
		if mapID == 0 {
			mapID = 1
		}

		enemyID := 0
		enemyLevel := 5
		if slots := getMapOgreSlots(mapID); len(slots) > 0 {
			if slotData, ok := slots[slot]; ok && slotData[0] > 0 {
				enemyID = int(slotData[0])
			} else {
				for _, v := range slots {
					if v[0] > 0 {
						enemyID = int(v[0])
						break
					}
				}
			}
		}
		if enemyID == 0 {
			enemyID = noviceBossID
		}

		player := resolveUserFightPet(user, user.CatchID, user.CurrentPetID)
		enemy := resolveEnemyFightPet(enemyID, enemyLevel)
		if enemy.CatchTime == 0 {
			enemy.CatchTime = uint32(time.Now().Unix())
		}

		user.Fight = &FightState{
			UserID:       ctx.UserID,
			PlayerPetID:  player.ID,
			PlayerLevel:  player.Level,
			PlayerDV:     player.DV,
			PlayerHP:     player.CurrentHP,
			PlayerMaxHP:  player.Stats.MaxHP,
			PlayerCatch:  player.CatchTime,
			PlayerSkills: player.Skills,
			PlayerStats:  player.Stats,
			PlayerType:   player.Type,
			EnemyPetID:   enemy.ID,
			EnemyLevel:   enemy.Level,
			EnemyHP:      enemy.CurrentHP,
			EnemyMaxHP:   enemy.Stats.MaxHP,
			EnemyCatch:   enemy.CatchTime,
			EnemySkills:  enemy.Skills,
			EnemyStats:   enemy.Stats,
			EnemyType:    enemy.Type,
		}

		ctx.Server.SendResponse(ctx.Conn, 2408, ctx.UserID, []byte{})

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(2))
		buf.Write(buildFightUserInfo(ctx.UserID, pickNick(user, ctx.UserID)))
		binary.Write(buf, binary.BigEndian, uint32(1))
		buf.Write(buildSimpleFightPetInfo(int(player.ID), int(player.Level), player.CurrentHP, player.Stats.MaxHP, int(player.CatchTime), player.Skills, int(mapID), int(player.ID)))
		binary.Write(buf, binary.BigEndian, uint32(0))
		protocol.WriteFixedString(buf, "", 16)
		binary.Write(buf, binary.BigEndian, uint32(1))
		buf.Write(buildSimpleFightPetInfo(int(enemy.ID), int(enemy.Level), enemy.CurrentHP, enemy.Stats.MaxHP, int(enemy.CatchTime), enemy.Skills, int(mapID), int(enemy.ID)))
		ctx.Server.SendResponse(ctx.Conn, 2503, ctx.UserID, buf.Bytes())
		sendPveFightStart(ctx, user, user.Fight)
	}
}

func sendPveFightStart(ctx *gateway.Context, user *User, f *FightState) {
	// Deprecated: Logic moved to handleReadyToFight (CMD 2404)
	// Do not send 2504 or 2301 here. Client will request 2404 when ready.
}

func sendFightPetInfo(ctx *gateway.Context, user *User, f *FightState) {
	if ctx == nil || user == nil || f == nil {
		return
	}
	var body []byte
	if pet := findPetByCatchTime(user, f.PlayerCatch); pet != nil {
		body = buildFullPetInfo(int(pet.ID), int(pet.CatchTime), int(pet.Level), int(pet.DV), pet.Exp, pet.Skills)
	} else {
		body = buildFullPetInfo(int(f.PlayerPetID), int(f.PlayerCatch), int(f.PlayerLevel), int(f.PlayerDV), 0, f.PlayerSkills)
	}
	if len(body) > 0 {
		ctx.Server.SendResponse(ctx.Conn, 2301, ctx.UserID, body)
	}
}

func handleAttackBoss(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		mapID := user.MapID
		reader := NewReader(ctx.Body)
		region := uint32(0)
		if reader.Remaining() >= 4 {
			region = reader.ReadUint32BE()
		}
		entries := getMapBossEntries(int(mapID))
		entry, ok := entries[region]
		if !ok || !entry.HasShield {
			ctx.Server.SendResponse(ctx.Conn, 2412, ctx.UserID, make([]byte, 4))
			return
		}

		maxHP := 0
		if boss := GetSPTBossByID(entry.BossPetID); boss != nil && boss.MaxHP > 0 {
			maxHP = boss.MaxHP
		} else {
			base := LoadPetDB().pets[entry.BossPetID]
			stats := getStats(base, entry.Level, 15, evSet{})
			maxHP = stats.MaxHP
		}
		if maxHP <= 0 {
			maxHP = 1
		}
		key := bossShieldKey(mapID, region)
		currentHP := int(user.BossShield[key])
		if currentHP <= 0 {
			currentHP = maxHP
		}
		damage := maxHP / 4
		if damage < 1 {
			damage = 1
		}
		newHP := currentHP - damage
		if newHP < 0 {
			newHP = 0
		}
		user.BossShield[key] = uint32(newHP)

		resp := new(bytes.Buffer)
		binary.Write(resp, binary.BigEndian, uint32(newHP))
		ctx.Server.SendResponse(ctx.Conn, 2412, ctx.UserID, resp.Bytes())

		body := buildMapBossListForUser(user, int(mapID))
		ctx.Server.SendResponse(ctx.Conn, 2021, ctx.UserID, body)
	}
}

func handlePetKingJoin() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2413, ctx.UserID, []byte{})
	}
}

func handleNpcJoin() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2427, ctx.UserID, []byte{})
	}
}

func handleStartPetWar() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2431, ctx.UserID, []byte{})
	}
}

func handleLoadPercent() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(100))
		ctx.Server.SendResponse(ctx.Conn, 2441, ctx.UserID, buf.Bytes())
	}
}

func updateFightResult(deps *Deps, user *User, f *FightState, won bool) []int {
	if user == nil || f == nil {
		return nil
	}
	var learned []int
	for i := range user.Pets {
		if user.Pets[i].CatchTime != f.PlayerCatch {
			continue
		}
		p := &user.Pets[i]
		oldLevel := p.Level
		oldSkills := append([]int{}, p.Skills...)
		if len(oldSkills) == 0 {
			base := LoadPetDB().pets[int(p.ID)]
			oldSkills = getSkillsForLevel(base, int(oldLevel))
		}
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
			if p.Level > oldLevel {
				newSkills := getSkillsForLevel(base, int(p.Level))
				learned = diffSkills(newSkills, oldSkills)
				if len(learned) > 0 {
					cur := append([]int{}, oldSkills...)
					cur = normalizeSkillList(cur, base, int(oldLevel))
					for _, sid := range learned {
						placed := false
						for idx := range cur {
							if cur[idx] == 0 {
								cur[idx] = sid
								placed = true
								break
							}
						}
						if placed {
							continue
						}
					}
					p.Skills = cur
				}
			}
		}
		upsertPet(deps, user, *p)
		if won && f.EnemyRewardID > 0 {
			grantItem(deps, user, f.EnemyRewardID, maxInt(1, f.EnemyRewardCt))
		}
		break
	}
	return learned
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

func buildSimpleFightPetInfo(petID int, level int, hp int, maxHP int, catchTime int, skills []int, catchMap int, skinID int) []byte {
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
	// 对齐 Lua：catchMap/catchRect/catchLevel/skinID
	if catchMap == 0 {
		catchMap = 301
	}
	binary.Write(buf, binary.BigEndian, uint32(catchMap))
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(level))
	binary.Write(buf, binary.BigEndian, uint32(skinID))
	return buf.Bytes()
}

func buildFightPetInfo(userID uint32, petID uint32, catchTime uint32, hp int, maxHP int, level uint32, catchable uint32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, userID)
	binary.Write(buf, binary.BigEndian, petID)

	name := ""
	if base := LoadPetDB().pets[int(petID)]; base != nil {
		name = base.Name
	}
	protocol.WriteFixedString(buf, name, 16)

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
	body := buildFightOverBody(reason, winner)
	ctx.Server.SendResponse(ctx.Conn, 2506, ctx.UserID, body)
}

func buildFightOverBody(reason uint32, winner uint32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, reason)
	binary.Write(buf, binary.BigEndian, winner)
	buf.Write(make([]byte, 20))
	return buf.Bytes()
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
	statusIceSeal   = 15
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
	effectPoison      = 11
	effectBurn        = 12
	effectFreeze      = 14
	effectFlinch      = 15
	effectConfuse     = 16
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

func effectiveSpeed(base int, stage int, status map[int]int) int {
	speed := applyStageModifier(base, stage)
	if status != nil && status[statusParalysis] > 0 {
		speed /= 2
	}
	if speed < 1 {
		speed = 1
	}
	return speed
}

func calculateAccuracy(baseAccuracy int, attackerAcc int, defenderEva int) float64 {
	if baseAccuracy <= 0 {
		baseAccuracy = 100
	}
	netStage := clampStage(attackerAcc - defenderEva)
	mul := stageMultipliers[netStage]
	if mul == 0 {
		mul = 1
	}
	return float64(baseAccuracy) * mul
}

func checkHit(baseAccuracy int, attackerAcc int, defenderEva int) bool {
	if baseAccuracy >= 100 {
		return true
	}
	acc := calculateAccuracy(baseAccuracy, attackerAcc, defenderEva)
	return rand.Float64()*100 < acc
}

func resolveUserFightPet(user *User, catchTime uint32, petID uint32) fightPetSnapshot {
	var picked *Pet
	if user != nil {
		// Priority:
		// 1) Explicit petID
		// 2) Explicit catchTime
		// 3) CatchID (Last captured/selected)
		// 4) CurrentPetID
		// 5) First pet in bag
		if petID > 0 {
			for i := range user.Pets {
				if user.Pets[i].ID == petID {
					picked = &user.Pets[i]
					break
				}
			}
		}
		if picked == nil && catchTime > 0 {
			for i := range user.Pets {
				if user.Pets[i].CatchTime == catchTime {
					picked = &user.Pets[i]
					break
				}
			}
		}
		if picked == nil && user.CatchID > 0 {
			for i := range user.Pets {
				if user.Pets[i].CatchTime == user.CatchID {
					picked = &user.Pets[i]
					break
				}
			}
		}
		if picked == nil && user.CurrentPetID > 0 {
			for i := range user.Pets {
				if user.Pets[i].ID == user.CurrentPetID {
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
		f.PlayerDV = player.DV
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
		f.EnemyDV = enemy.DV
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

func ensureFightSkillPP(f *FightState) {
	if f == nil {
		return
	}
	if f.PlayerSkillPP == nil {
		f.PlayerSkillPP = make(map[int]int)
		for _, sid := range f.PlayerSkills {
			if sid > 0 {
				f.PlayerSkillPP[sid] = getSkillPP(sid)
			}
		}
	}
	if f.EnemySkillPP == nil {
		f.EnemySkillPP = make(map[int]int)
		for _, sid := range f.EnemySkills {
			if sid > 0 {
				f.EnemySkillPP[sid] = getSkillPP(sid)
			}
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

func selectPlayerSkillWithPP(skills []int, requested int, pp map[int]int) int {
	if requested > 0 && containsSkill(skills, requested) {
		if pp == nil || pp[requested] > 0 {
			return requested
		}
	}
	for _, sid := range skills {
		if sid > 0 && (pp == nil || pp[sid] > 0) {
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

func selectRandomSkillWithPP(skills []int, pp map[int]int) int {
	valid := make([]int, 0, len(skills))
	for _, sid := range skills {
		if sid > 0 && (pp == nil || pp[sid] > 0) {
			valid = append(valid, sid)
		}
	}
	if len(valid) == 0 {
		return 0
	}
	return valid[rand.Intn(len(valid))]
}

func pickSkillWithEncore(requested int, skills []int, pp map[int]int, encoreSkill *int, encoreTurns *int) int {
	if encoreSkill != nil && encoreTurns != nil && *encoreTurns > 0 && *encoreSkill > 0 {
		if pp != nil && pp[*encoreSkill] <= 0 {
			*encoreTurns = 0
			*encoreSkill = 0
		} else {
			*encoreTurns--
			if *encoreTurns <= 0 {
				*encoreSkill = 0
			}
			return *encoreSkill
		}
	}
	return selectPlayerSkillWithPP(skills, requested, pp)
}

func pickEncoreSkill(encoreSkill *int, encoreTurns *int, pp map[int]int) int {
	if encoreSkill == nil || encoreTurns == nil {
		return 0
	}
	if *encoreTurns <= 0 || *encoreSkill <= 0 {
		return 0
	}
	if pp != nil && pp[*encoreSkill] <= 0 {
		return 0
	}
	*encoreTurns--
	if *encoreTurns <= 0 {
		*encoreSkill = 0
	}
	return *encoreSkill
}

func selectAISkill(f *FightState) int {
	if f == nil {
		return 0
	}
	best := 0
	bestScore := -1.0
	for _, sid := range f.EnemySkills {
		if sid <= 0 {
			continue
		}
		if f.EnemySkillPP != nil && f.EnemySkillPP[sid] <= 0 {
			continue
		}
		info := getSkillInfo(sid)
		if info == nil {
			continue
		}
		score := 10.0
		if info.Power > 0 && info.Category != 4 {
			typeMod := elementMultiplier(info.Type, f.PlayerType)
			score = float64(info.Power) * typeMod * (float64(info.Accuracy) / 100.0)
			if f.PlayerMaxHP > 0 && float64(f.PlayerHP)/float64(f.PlayerMaxHP) < 0.3 {
				score *= 1.5
			}
		}
		if score > bestScore {
			bestScore = score
			best = sid
		}
	}
	return best
}

func consumeSkillPP(pp map[int]int, skillID int) (int, int, bool) {
	if pp == nil || skillID <= 0 {
		return 0, 0, false
	}
	maxPP := getSkillPP(skillID)
	cur := pp[skillID]
	if cur == 0 {
		cur = maxPP
	}
	if cur <= 0 {
		pp[skillID] = 0
		return 0, maxPP, true
	}
	cur--
	pp[skillID] = cur
	return cur, maxPP, true
}

func sendSkillPPUpdate(ctx *gateway.Context, skillID int, pp int, maxPP int) {
	if ctx == nil || skillID <= 0 {
		return
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, ctx.UserID)
	binary.Write(buf, binary.BigEndian, uint32(skillID))
	binary.Write(buf, binary.BigEndian, uint32(pp))
	binary.Write(buf, binary.BigEndian, uint32(maxPP))
	ctx.Server.SendResponse(ctx.Conn, 2507, ctx.UserID, buf.Bytes())
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

func checkCrit(info *SkillInfo, attackerHP int, attackerMaxHP int, defenderHP int, defenderMaxHP int, atkStage stageModifiers, isFirst bool) bool {
	if info == nil {
		return rand.Intn(16) == 0
	}
	if info.CritAtkFirst && isFirst {
		return true
	}
	if info.CritAtkSecond && !isFirst {
		return true
	}
	if info.CritSelfHalfHp && attackerMaxHP > 0 && attackerHP < attackerMaxHP/2 {
		return true
	}
	if info.CritFoeHalfHp && defenderMaxHP > 0 && defenderHP < defenderMaxHP/2 {
		return true
	}
	critRate := info.CritRate
	if critRate <= 0 {
		critRate = 1
	}
	bonus := 0
	if atkStage.Spd > 0 {
		bonus = atkStage.Spd
	}
	threshold := critRate + bonus
	if threshold < 1 {
		threshold = 1
	}
	return rand.Intn(16)+1 <= threshold
}

func executeAttack(ctx *gateway.Context, f *FightState, player bool, skillID int, isFirst bool) attackResult {
	if f == nil {
		return attackResult{}
	}
	ensureFightStatus(f)
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
	defStatus := f.PlayerStatus
	atkDV := int(f.EnemyDV)
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
		defStatus = f.EnemyStatus
		atkDV = int(f.PlayerDV)
	}

	info := getSkillInfo(skillID)
	if info == nil {
		return placeholderAttack(player, f)
	}
	if !info.MustHit {
		if !checkHit(info.Accuracy, atkStage.Acc, defStage.Eva) {
			return attackResult{
				UserID:   attackerID,
				SkillID:  uint32(skillID),
				AtkTimes: 0,
				LostHP:   0,
				GainHP:   0,
				RemainHP: atkHP,
				MaxHP:    atkMaxHP,
				State:    1,
				IsCrit:   0,
				PetType:  uint32(atkType),
				Stage:    atkStage,
			}
		}
	}

	hitCount := uint32(1)
	if info.SideEffect == effectMultiHit {
		minHits := 2
		maxHits := 5
		args := parseEffectArgs(info.SideEffectArg)
		if len(args) >= 2 {
			minHits = args[0]
			maxHits = args[1]
		}
		if maxHits < minHits {
			maxHits = minHits
		}
		if minHits < 1 {
			minHits = 1
		}
		hitCount = uint32(minHits + rand.Intn(maxHits-minHits+1))
	}

	defHP := f.PlayerHP
	defMaxHP := f.PlayerMaxHP
	if player {
		defHP = f.EnemyHP
		defMaxHP = f.EnemyMaxHP
	}
	totalDamage := 0
	critHit := false
	for i := uint32(0); i < hitCount; i++ {
		isCrit := checkCrit(info, atkHP, atkMaxHP, defHP, defMaxHP, atkStage, isFirst)
		if isCrit {
			critHit = true
		}
		damage := calcDamagePower(atkStats, defStats, atkLevel, atkDV, info, atkType, defType, atkStage, defStage, defStatus, isCrit)
		if info.SideEffect == effectHpRatio {
			ratio := 50
			args := parseEffectArgs(info.SideEffectArg)
			if len(args) > 0 {
				ratio = args[0]
			} else if info.Power > 0 {
				ratio = info.Power
			}
			if ratio < 1 {
				ratio = 1
			}
			damage = defHP * ratio / 100
			if damage < 1 && defHP > 0 {
				damage = 1
			}
		} else if info.SideEffect == effectPunishment {
			extra := sumPositiveStages(defStage) * 20
			infoCopy := *info
			infoCopy.Power += extra
			damage = calcDamagePower(atkStats, defStats, atkLevel, atkDV, &infoCopy, atkType, defType, atkStage, defStage, defStatus, isCrit)
		}
		totalDamage += damage
	}

	if info.SideEffect == effectMercy && defHP-totalDamage < 1 {
		totalDamage = maxInt(0, defHP-1)
	}
	if totalDamage < 0 {
		totalDamage = 0
	}

	gainHP := 0
	recoilDamage := 0
	if totalDamage > 0 {
		if info.SideEffect == effectDrain {
			gainHP = totalDamage / 2
		}
		if info.SideEffect == effectRecoil {
			divisor := 4
			args := parseEffectArgs(info.SideEffectArg)
			if len(args) >= 1 && args[0] > 0 {
				divisor = args[0]
			}
			recoilDamage = totalDamage / divisor
		}
	}

	if player {
		f.EnemyHP = maxInt(0, f.EnemyHP-totalDamage)
		if gainHP > 0 {
			f.PlayerHP = minInt(f.PlayerMaxHP, f.PlayerHP+gainHP)
		}
		if recoilDamage > 0 {
			f.PlayerHP = maxInt(0, f.PlayerHP-recoilDamage)
		}
	} else {
		f.PlayerHP = maxInt(0, f.PlayerHP-totalDamage)
		if gainHP > 0 {
			f.EnemyHP = minInt(f.EnemyMaxHP, f.EnemyHP+gainHP)
		}
		if recoilDamage > 0 {
			f.EnemyHP = maxInt(0, f.EnemyHP-recoilDamage)
		}
	}

	if info.SideEffect == effectSelfStage || info.SideEffect == effectStageChange {
		applyStageEffects(f, player, info)
		if player {
			atkStage = f.PlayerStage
		} else {
			atkStage = f.EnemyStage
		}
	}
	if info.SideEffect != effectDrain && info.SideEffect != effectSelfStage && info.SideEffect != effectStageChange &&
		info.SideEffect != effectRecoil && info.SideEffect != effectMercy && info.SideEffect != effectMultiHit &&
		info.SideEffect != effectHpRatio && info.SideEffect != effectPunishment {
		applySkillEffect(f, player, info, totalDamage)
	}
	if info.SideEffect == effectFatigue {
		turns := 1
		args := parseEffectArgs(info.SideEffectArg)
		if len(args) >= 2 && args[1] > 0 {
			turns = args[1]
		}
		if player {
			f.PlayerFatigue = maxInt(f.PlayerFatigue, turns)
		} else {
			f.EnemyFatigue = maxInt(f.EnemyFatigue, turns)
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
		LostHP:   uint32(totalDamage),
		GainHP:   gainHP,
		RemainHP: remainHP,
		MaxHP:    atkMaxHP,
		State:    0,
		IsCrit:   boolToUint32(critHit),
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

func cannotActAttack(userID uint32, player bool, f *FightState) attackResult {
	if player {
		return attackResult{
			UserID:   userID,
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
func calcDamagePower(atk petStats, def petStats, level int, dv int, info *SkillInfo, atkType int, defType int, atkStage stageModifiers, defStage stageModifiers, defStatus map[int]int, isCrit bool) int {
	if info == nil {
		return 0
	}
	if level <= 0 {
		level = 1
	}
	if info.DmgBindLv {
		return level
	}
	if info.Power <= 0 || info.Category == 4 {
		return 0
	}
	power := info.Power
	if info.PwrBindDv > 0 {
		mult := 5
		if info.PwrBindDv == 2 {
			mult = 10
		}
		if dv <= 0 {
			dv = 15
		}
		power = dv * mult
	}
	if info.PwrDouble && hasMajorStatus(defStatus) {
		power *= 2
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
	base := (float64(level)*0.4 + 2) * float64(power) * float64(atkVal) / float64(defVal) / 50.0
	base += 2
	effectiveness := elementMultiplier(info.Type, defType)
	stab := 1.0
	if info.Type > 0 && info.Type == atkType {
		stab = 1.5
	}
	critMod := 1.0
	if isCrit {
		critMod = 1.5
	}
	randomMod := float64(85+rand.Intn(16)) / 100.0
	final := base * effectiveness * stab * critMod * randomMod
	damage := int(final)
	if effectiveness > 0 && damage < 1 {
		damage = 1
	}
	if effectiveness == 0 {
		damage = 0
	}
	return damage
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

func hasMajorStatus(status map[int]int) bool {
	if status == nil {
		return false
	}
	for _, id := range []int{statusParalysis, statusPoison, statusBurn, statusFreeze, statusFear, statusSleep, statusPetrify, statusConfuse, statusBleed, statusIceSeal} {
		if status[id] > 0 {
			return true
		}
	}
	return false
}

func applySkillEffect(f *FightState, player bool, info *SkillInfo, damage int) {
	if f == nil || info == nil || info.SideEffect <= 0 {
		return
	}
	effectID := info.SideEffect
	args := parseEffectArgs(info.SideEffectArg)
	if applyDirectEffect(f, player, effectID, args) {
		return
	}
	effect := getSkillEffect(effectID)
	if effect == nil {
		return
	}
	if len(args) == 0 {
		args = parseEffectArgs(effect.Args)
	}
	applyEffectByEID(f, player, effect.Eid, args, damage)
}

func applyDirectEffect(f *FightState, player bool, effectID int, args []int) bool {
	switch effectID {
	case effectParalysis:
		applyStatusWithChance(f, player, statusParalysis, args, 10, 999)
		return true
	case effectPoison:
		applyStatusWithChance(f, player, statusPoison, args, 10, 999)
		return true
	case effectBurn:
		applyStatusWithChance(f, player, statusBurn, args, 10, 999)
		return true
	case effectFreeze:
		applyStatusWithChance(f, player, statusFreeze, args, 10, 3)
		return true
	case effectFlinch:
		chance := 10
		if len(args) >= 1 && args[0] > 0 {
			chance = args[0]
		}
		if chance < 100 && rand.Intn(100)+1 > chance {
			return true
		}
		setFlinch(f, player)
		return true
	case effectConfuse:
		applyStatusWithChance(f, player, statusConfuse, args, 10, 3)
		return true
	default:
		return false
	}
}

func applyStatusWithChance(f *FightState, player bool, statusID int, args []int, defaultChance int, defaultTurns int) {
	if f == nil {
		return
	}
	chance := defaultChance
	turns := defaultTurns
	if len(args) >= 1 && args[0] > 0 {
		chance = args[0]
	}
	if len(args) >= 2 && args[1] > 0 {
		turns = args[1]
	}
	if chance < 100 && rand.Intn(100)+1 > chance {
		return
	}
	if player {
		if f.EnemyStatus == nil {
			f.EnemyStatus = make(map[int]int)
		}
		if hasMajorStatus(f.EnemyStatus) {
			return
		}
		f.EnemyStatus[statusID] = maxInt(f.EnemyStatus[statusID], turns)
	} else {
		if f.PlayerStatus == nil {
			f.PlayerStatus = make(map[int]int)
		}
		if hasMajorStatus(f.PlayerStatus) {
			return
		}
		f.PlayerStatus[statusID] = maxInt(f.PlayerStatus[statusID], turns)
	}
}

func setFlinch(f *FightState, player bool) {
	if f == nil {
		return
	}
	if player {
		f.EnemyFlinch = true
	} else {
		f.PlayerFlinch = true
	}
}

func setBound(f *FightState, player bool, turns int) {
	if turns <= 0 {
		turns = 4
	}
	if player {
		if f.EnemyBoundTurns < turns {
			f.EnemyBoundTurns = turns
		}
	} else {
		if f.PlayerBoundTurns < turns {
			f.PlayerBoundTurns = turns
		}
	}
}

func applyEffectByEID(f *FightState, player bool, eid int, args []int, damage int) {
	if f == nil {
		return
	}
	switch eid {
	case 1:
		healPercent := 50
		if len(args) >= 1 && args[0] > 0 {
			healPercent = args[0]
		}
		heal := damage * healPercent / 100
		if heal <= 0 {
			return
		}
		if player {
			f.PlayerHP = minInt(f.PlayerMaxHP, f.PlayerHP+heal)
		} else {
			f.EnemyHP = minInt(f.EnemyMaxHP, f.EnemyHP+heal)
		}
	case 2:
		stat := 1
		stages := 1
		if len(args) >= 1 {
			stat = args[0]
		}
		if len(args) >= 2 {
			stages = args[1]
		}
		applyStageChangeTo(f, player, true, stat, -stages)
	case 3, 4:
		stat := 0
		stages := 1
		if len(args) >= 1 {
			stat = args[0]
		}
		if len(args) >= 2 {
			stages = args[1]
		}
		applyStageChangeTo(f, player, false, stat, stages)
	case 5:
		stat := 4
		chance := 100
		stages := 1
		if len(args) >= 1 {
			stat = args[0]
		}
		if len(args) >= 2 {
			chance = args[1]
		}
		if len(args) >= 3 {
			stages = args[2]
		}
		if chance < 100 && rand.Intn(100)+1 > chance {
			return
		}
		applyStageChangeTo(f, player, true, stat, -stages)
	case 6:
		recoilPercent := 25
		if len(args) >= 1 && args[0] > 0 {
			recoilPercent = args[0]
		}
		recoil := damage * recoilPercent / 100
		if player {
			f.PlayerHP = maxInt(0, f.PlayerHP-recoil)
		} else {
			f.EnemyHP = maxInt(0, f.EnemyHP-recoil)
		}
	case 7:
		if player {
			f.EnemyHP = minInt(f.EnemyMaxHP, f.PlayerHP)
		} else {
			f.PlayerHP = minInt(f.PlayerMaxHP, f.EnemyHP)
		}
	case 8:
		if player {
			if f.EnemyHP <= 0 {
				f.EnemyHP = 1
			}
		} else {
			if f.PlayerHP <= 0 {
				f.PlayerHP = 1
			}
		}
	case 9:
		minDamage := 20
		maxDamage := 80
		if len(args) >= 1 {
			minDamage = args[0]
		}
		if len(args) >= 2 {
			maxDamage = args[1]
		}
		if damage >= minDamage && damage <= maxDamage {
			applyStageChangeTo(f, player, false, 0, 1)
		}
	case 10:
		applyStatusWithChance(f, player, statusParalysis, args, 10, 999)
	case 11:
		chance := 100
		if len(args) >= 1 && args[0] > 0 {
			chance = args[0]
		}
		if chance < 100 && rand.Intn(100)+1 > chance {
			return
		}
		setBound(f, player, 4)
	case 12:
		applyStatusWithChance(f, player, statusBurn, args, 10, 999)
	case 13:
		applyStatusWithChance(f, player, statusPoison, args, 10, 999)
	case 14:
		chance := 100
		if len(args) >= 1 && args[0] > 0 {
			chance = args[0]
		}
		if chance < 100 && rand.Intn(100)+1 > chance {
			return
		}
		setBound(f, player, 4)
	case 15, 29:
		chance := 10
		if len(args) >= 1 && args[0] > 0 {
			chance = args[0]
		}
		if chance < 100 && rand.Intn(100)+1 > chance {
			return
		}
		setFlinch(f, player)
	case 20:
		turns := 1
		if len(args) >= 2 && args[1] > 0 {
			turns = args[1]
		}
		if player {
			f.PlayerFatigue = maxInt(f.PlayerFatigue, turns)
		} else {
			f.EnemyFatigue = maxInt(f.EnemyFatigue, turns)
		}
	case 31:
		return
	case 33:
		if player {
			reduceSkillPP(f.EnemySkillPP, f.EnemyLastSkill)
		} else {
			reduceSkillPP(f.PlayerSkillPP, f.PlayerLastSkill)
		}
	case 34:
		turns := 2
		if len(args) >= 1 && args[0] > 0 {
			turns = args[0]
		}
		if player {
			f.EnemyEncoreSkill = f.EnemyLastSkill
			f.EnemyEncoreTurns = maxInt(f.EnemyEncoreTurns, turns)
		} else {
			f.PlayerEncoreSkill = f.PlayerLastSkill
			f.PlayerEncoreTurns = maxInt(f.PlayerEncoreTurns, turns)
		}
	case 35:
		return
	}
}

func reduceSkillPP(pp map[int]int, skillID int) {
	if pp == nil || skillID <= 0 {
		return
	}
	cur := pp[skillID]
	if cur == 0 {
		cur = getSkillPP(skillID)
	}
	if cur <= 0 {
		pp[skillID] = 0
		return
	}
	cur--
	if cur < 0 {
		cur = 0
	}
	pp[skillID] = cur
}

func applyStageChangeTo(f *FightState, player bool, targetDefender bool, stat int, stages int) {
	if f == nil {
		return
	}
	base := stageStatIndex[stat]
	if base == (fightStateChange{}) {
		return
	}
	change := fightStateChange{
		Atk: base.Atk * stages,
		Def: base.Def * stages,
		SpA: base.SpA * stages,
		SpD: base.SpD * stages,
		Spd: base.Spd * stages,
		Acc: base.Acc * stages,
	}
	if player {
		if targetDefender {
			f.EnemyStage = clampStageChange(f.EnemyStage, change)
		} else {
			f.PlayerStage = clampStageChange(f.PlayerStage, change)
		}
	} else {
		if targetDefender {
			f.PlayerStage = clampStageChange(f.PlayerStage, change)
		} else {
			f.EnemyStage = clampStageChange(f.EnemyStage, change)
		}
	}
}

func canAct(f *FightState, player bool) bool {
	if f == nil {
		return false
	}
	var status map[int]int
	if player {
		status = f.PlayerStatus
	} else {
		status = f.EnemyStatus
	}
	if status == nil {
		status = make(map[int]int)
		if player {
			f.PlayerStatus = status
		} else {
			f.EnemyStatus = status
		}
	}
	if player {
		if f.PlayerFatigue > 0 {
			f.PlayerFatigue--
			return false
		}
	} else {
		if f.EnemyFatigue > 0 {
			f.EnemyFatigue--
			return false
		}
	}
	if status[statusSleep] > 0 {
		status[statusSleep]--
		if status[statusSleep] <= 0 {
			delete(status, statusSleep)
		}
		return false
	}
	if status[statusPetrify] > 0 {
		status[statusPetrify]--
		if status[statusPetrify] <= 0 {
			delete(status, statusPetrify)
		}
		return false
	}
	if status[statusIceSeal] > 0 {
		status[statusIceSeal]--
		if status[statusIceSeal] <= 0 {
			delete(status, statusIceSeal)
		}
		return false
	}
	if status[statusFreeze] > 0 {
		status[statusFreeze]--
		if status[statusFreeze] <= 0 {
			delete(status, statusFreeze)
		}
		return false
	}
	if status[statusParalysis] > 0 {
		if rand.Intn(4) == 0 {
			return false
		}
	}
	if status[statusFear] > 0 {
		status[statusFear]--
		if status[statusFear] <= 0 {
			delete(status, statusFear)
		}
		if rand.Intn(2) == 0 {
			return false
		}
	}
	if status[statusConfuse] > 0 {
		status[statusConfuse]--
		if status[statusConfuse] <= 0 {
			delete(status, statusConfuse)
		}
		if rand.Intn(3) == 0 {
			return false
		}
	}
	if player && f.PlayerFlinch {
		f.PlayerFlinch = false
		return false
	}
	if !player && f.EnemyFlinch {
		f.EnemyFlinch = false
		return false
	}
	return true
}

func applyTurnStatusDamage(status map[int]int, hp *int, maxHP int, boundTurns *int) int {
	if status == nil || hp == nil || maxHP <= 0 {
		return 0
	}
	damage := 0
	if status[statusPoison] > 0 {
		damage += maxHP / 8
		status[statusPoison]--
		if status[statusPoison] <= 0 {
			delete(status, statusPoison)
		}
	}
	if status[statusBurn] > 0 {
		damage += maxHP / 16
		status[statusBurn]--
		if status[statusBurn] <= 0 {
			delete(status, statusBurn)
		}
	}
	if status[statusFreeze] > 0 {
		damage += maxHP / 16
		status[statusFreeze]--
		if status[statusFreeze] <= 0 {
			delete(status, statusFreeze)
		}
	}
	if status[statusBleed] > 0 {
		damage += maxHP / 8
		status[statusBleed]--
		if status[statusBleed] <= 0 {
			delete(status, statusBleed)
		}
	}
	if boundTurns != nil && *boundTurns > 0 {
		damage += maxHP / 16
		*boundTurns--
	}
	if damage > 0 {
		*hp = maxInt(0, *hp-damage)
	}
	return damage
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

func initPvPFightState(state *State, inviterID uint32, responderID uint32, inviter *User, responder *User) {
	if state == nil || inviter == nil || responder == nil {
		return
	}
	invPlayer := resolveUserFightPet(inviter, inviter.CatchID, inviter.CurrentPetID)
	resPlayer := resolveUserFightPet(responder, responder.CatchID, responder.CurrentPetID)
	invPlayer.CatchTime = ensureCatchTime(invPlayer.CatchTime, invPlayer.ID)
	resPlayer.CatchTime = ensureCatchTime(resPlayer.CatchTime, resPlayer.ID)

	inviter.Fight = &FightState{
		UserID:         inviterID,
		OpponentUserID: responderID,
		PlayerPetID:    invPlayer.ID,
		PlayerLevel:    invPlayer.Level,
		PlayerDV:       invPlayer.DV,
		PlayerHP:       invPlayer.CurrentHP,
		PlayerMaxHP:    invPlayer.Stats.MaxHP,
		PlayerCatch:    invPlayer.CatchTime,
		PlayerSkills:   invPlayer.Skills,
		PlayerStats:    invPlayer.Stats,
		PlayerType:     invPlayer.Type,
		EnemyPetID:     resPlayer.ID,
		EnemyLevel:     resPlayer.Level,
		EnemyDV:        resPlayer.DV,
		EnemyHP:        resPlayer.CurrentHP,
		EnemyMaxHP:     resPlayer.Stats.MaxHP,
		EnemyCatch:     resPlayer.CatchTime,
		EnemySkills:    resPlayer.Skills,
		EnemyStats:     resPlayer.Stats,
		EnemyType:      resPlayer.Type,
	}
	responder.Fight = &FightState{
		UserID:         responderID,
		OpponentUserID: inviterID,
		PlayerPetID:    resPlayer.ID,
		PlayerLevel:    resPlayer.Level,
		PlayerDV:       resPlayer.DV,
		PlayerHP:       resPlayer.CurrentHP,
		PlayerMaxHP:    resPlayer.Stats.MaxHP,
		PlayerCatch:    resPlayer.CatchTime,
		PlayerSkills:   resPlayer.Skills,
		PlayerStats:    resPlayer.Stats,
		PlayerType:     resPlayer.Type,
		EnemyPetID:     invPlayer.ID,
		EnemyLevel:     invPlayer.Level,
		EnemyDV:        invPlayer.DV,
		EnemyHP:        invPlayer.CurrentHP,
		EnemyMaxHP:     invPlayer.Stats.MaxHP,
		EnemyCatch:     invPlayer.CatchTime,
		EnemySkills:    invPlayer.Skills,
		EnemyStats:     invPlayer.Stats,
		EnemyType:      invPlayer.Type,
	}
	inviter.InFight = true
	responder.InFight = true
}

func buildNoteReadyToFightPvP(inviterID uint32, responderID uint32, inviter *User, responder *User) (inviterBody []byte, responderBody []byte) {
	if inviter == nil || responder == nil {
		return nil, nil
	}
	invPet := resolveUserFightPet(inviter, inviter.CatchID, inviter.CurrentPetID)
	resPet := resolveUserFightPet(responder, responder.CatchID, responder.CurrentPetID)
	invPet.CatchTime = ensureCatchTime(invPet.CatchTime, invPet.ID)
	resPet.CatchTime = ensureCatchTime(resPet.CatchTime, resPet.ID)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(2))
	buf.Write(buildFightUserInfo(inviterID, pickNick(inviter, inviterID)))
	binary.Write(buf, binary.BigEndian, uint32(1))
	buf.Write(buildSimpleFightPetInfo(int(invPet.ID), int(invPet.Level), invPet.CurrentHP, invPet.Stats.MaxHP, int(invPet.CatchTime), invPet.Skills, int(inviter.MapID), int(invPet.ID)))
	buf.Write(buildFightUserInfo(responderID, pickNick(responder, responderID)))
	binary.Write(buf, binary.BigEndian, uint32(1))
	buf.Write(buildSimpleFightPetInfo(int(resPet.ID), int(resPet.Level), resPet.CurrentHP, resPet.Stats.MaxHP, int(resPet.CatchTime), resPet.Skills, int(responder.MapID), int(resPet.ID)))
	inviterBody = buf.Bytes()

	const blockSize = 20 + 4 + 72
	if len(inviterBody) < 4+blockSize*2 {
		return inviterBody, inviterBody
	}
	responderBody = make([]byte, len(inviterBody))
	copy(responderBody[0:4], inviterBody[0:4])
	copy(responderBody[4:4+blockSize], inviterBody[4+blockSize:4+blockSize*2])
	copy(responderBody[4+blockSize:4+blockSize*2], inviterBody[4:4+blockSize])
	return inviterBody, responderBody
}

func buildNoteStartFightPvP(selfID uint32, selfFight *FightState, otherID uint32, otherFight *FightState) []byte {
	if selfFight == nil || otherFight == nil {
		return nil
	}
	if selfFight.PlayerMaxHP <= 0 {
		selfFight.PlayerMaxHP = 1
	}
	if otherFight.PlayerMaxHP <= 0 {
		otherFight.PlayerMaxHP = 1
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(0))
	buf.Write(buildFightPetInfo(selfID, selfFight.PlayerPetID, selfFight.PlayerCatch, selfFight.PlayerHP, selfFight.PlayerMaxHP, selfFight.PlayerLevel, 0))
	buf.Write(buildFightPetInfo(otherID, otherFight.PlayerPetID, otherFight.PlayerCatch, otherFight.PlayerHP, otherFight.PlayerMaxHP, otherFight.PlayerLevel, 1))
	return buf.Bytes()
}

func syncPvPFightState(state *State, user *User, f *FightState) {
	if state == nil || user == nil || f == nil || f.OpponentUserID == 0 {
		return
	}
	opp := state.GetOrCreateUser(f.OpponentUserID)
	if opp == nil || opp.Fight == nil {
		return
	}
	opp.Fight.PlayerHP = f.EnemyHP
	opp.Fight.PlayerMaxHP = f.EnemyMaxHP
	opp.Fight.PlayerStatus = cloneStatusMap(f.EnemyStatus)
	opp.Fight.PlayerStage = f.EnemyStage
	opp.Fight.PlayerFatigue = f.EnemyFatigue
	opp.Fight.PlayerSkillPP = cloneSkillPP(f.EnemySkillPP)
	opp.Fight.PlayerBoundTurns = f.EnemyBoundTurns
	opp.Fight.PlayerFlinch = f.EnemyFlinch
	opp.Fight.PlayerEncoreSkill = f.EnemyEncoreSkill
	opp.Fight.PlayerEncoreTurns = f.EnemyEncoreTurns
	opp.Fight.PlayerLastSkill = f.EnemyLastSkill
	opp.Fight.EnemyHP = f.PlayerHP
	opp.Fight.EnemyMaxHP = f.PlayerMaxHP
	opp.Fight.EnemyStatus = cloneStatusMap(f.PlayerStatus)
	opp.Fight.EnemyStage = f.PlayerStage
	opp.Fight.EnemyFatigue = f.PlayerFatigue
	opp.Fight.EnemySkillPP = cloneSkillPP(f.PlayerSkillPP)
	opp.Fight.EnemyBoundTurns = f.PlayerBoundTurns
	opp.Fight.EnemyFlinch = f.PlayerFlinch
	opp.Fight.EnemyEncoreSkill = f.PlayerEncoreSkill
	opp.Fight.EnemyEncoreTurns = f.PlayerEncoreTurns
	opp.Fight.EnemyLastSkill = f.PlayerLastSkill
}

func updateFightHP(deps *Deps, user *User, f *FightState) {
	if user == nil || f == nil {
		return
	}
	for i := range user.Pets {
		if user.Pets[i].CatchTime != f.PlayerCatch {
			continue
		}
		p := &user.Pets[i]
		if f.PlayerHP < 0 {
			f.PlayerHP = 0
		}
		p.HP = f.PlayerHP
		upsertPet(deps, user, *p)
		break
	}
}

func sendNoteUpdateProp(ctx *gateway.Context, user *User, catchTime uint32) {
	if ctx == nil || user == nil {
		return
	}
	body := buildNoteUpdatePropBody(user, catchTime)
	if len(body) == 0 {
		return
	}
	ctx.Server.SendResponse(ctx.Conn, 2508, ctx.UserID, body)
}

func buildNoteUpdatePropBody(user *User, catchTime uint32) []byte {
	if user == nil {
		return nil
	}
	pet := findPetByCatchTime(user, catchTime)
	if pet == nil && len(user.Pets) > 0 {
		pet = &user.Pets[0]
	}
	if pet == nil {
		return nil
	}
	return buildNoteUpdateProp(*pet)
}

func buildNoteUpdateProp(pet Pet) []byte {
	base := LoadPetDB().pets[int(pet.ID)]
	stats := getStats(base, int(pet.Level), int(pet.DV), evSet{})
	expInfo := getExpInfo(base, int(pet.Level), pet.Exp)
	hp := pet.HP
	if hp <= 0 {
		hp = stats.MaxHP
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, pet.CatchTime)
	binary.Write(buf, binary.BigEndian, pet.ID)
	binary.Write(buf, binary.BigEndian, pet.Level)
	binary.Write(buf, binary.BigEndian, uint32(expInfo.Exp))
	binary.Write(buf, binary.BigEndian, uint32(expInfo.LvExp))
	binary.Write(buf, binary.BigEndian, uint32(expInfo.NextLvExp))
	binary.Write(buf, binary.BigEndian, uint32(hp))
	binary.Write(buf, binary.BigEndian, uint32(stats.MaxHP))
	binary.Write(buf, binary.BigEndian, uint32(stats.Attack))
	binary.Write(buf, binary.BigEndian, uint32(stats.Defence))
	binary.Write(buf, binary.BigEndian, uint32(stats.SA))
	binary.Write(buf, binary.BigEndian, uint32(stats.SD))
	binary.Write(buf, binary.BigEndian, uint32(stats.Speed))
	for i := 0; i < 7; i++ {
		binary.Write(buf, binary.BigEndian, uint32(0))
	}
	return buf.Bytes()
}

func sendNoteUpdateSkill(ctx *gateway.Context, skills []int) {
	if ctx == nil || len(skills) == 0 {
		return
	}
	for _, sid := range skills {
		if sid <= 0 {
			continue
		}
		pp := getSkillPP(sid)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, uint32(sid))
		binary.Write(buf, binary.BigEndian, uint32(pp))
		binary.Write(buf, binary.BigEndian, uint32(pp))
		ctx.Server.SendResponse(ctx.Conn, 2507, ctx.UserID, buf.Bytes())
	}
}

func cloneStatusMap(src map[int]int) map[int]int {
	if src == nil {
		return nil
	}
	dst := make(map[int]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneSkillPP(src map[int]int) map[int]int {
	if src == nil {
		return nil
	}
	dst := make(map[int]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func diffSkills(newSkills []int, oldSkills []int) []int {
	seen := make(map[int]struct{}, len(oldSkills))
	for _, sid := range oldSkills {
		if sid > 0 {
			seen[sid] = struct{}{}
		}
	}
	var out []int
	for _, sid := range newSkills {
		if sid <= 0 {
			continue
		}
		if _, ok := seen[sid]; ok {
			continue
		}
		seen[sid] = struct{}{}
		out = append(out, sid)
	}
	return out
}

func ensureCatchTime(ct uint32, petID uint32) uint32 {
	if ct != 0 {
		return ct
	}
	return 0x69686700 + petID
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
