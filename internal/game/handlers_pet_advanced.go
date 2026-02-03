package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
)

func registerPetAdvancedHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2302, handleModifyPetName(deps, state))
	s.Register(2307, handlePetStudySkill(deps, state))
	s.Register(2308, handlePetDefault(deps, state))
	s.Register(2310, handlePetOneCure(deps, state))
	s.Register(2311, handlePetCollect())
	s.Register(2312, handlePetSkillSwitch())
	s.Register(2313, handleIsCollect())
	s.Register(2314, handlePetEvolution())
	s.Register(2315, handlePetHatch())
	s.Register(2316, handlePetHatchGet())
	s.Register(2318, handlePetSetExp(deps, state))
	s.Register(2319, handlePetGetExp(state))
	s.Register(2320, handlePetRoweiList())
	s.Register(2321, handlePetRowei())
	s.Register(2322, handlePetRetrieve())
	s.Register(2323, handlePetRoomShow())
	s.Register(2324, handlePetRoomList())
	s.Register(2325, handlePetRoomInfo())
	s.Register(2326, handleUsePetItemOutOfFight())
	s.Register(2327, handleUseSpeedupItem())
	s.Register(2328, handleSkillSort())
	s.Register(2329, handleUseAutoFightItem())
	s.Register(2330, handleOnOffAutoFight())
	s.Register(2331, handleUseEnergyXishou())
	s.Register(2332, handleUseStudyItem())
	s.Register(2343, handlePetResetNature())
	s.Register(2351, handlePetFusion())
	s.Register(2352, handleGetSoulBeadBuf())
	s.Register(2353, handleSetSoulBeadBuf())
	s.Register(2356, handleGetSoulBeadStatus())
	s.Register(2357, handleTransformSoulBead())
	s.Register(2358, handleSoulBeadToPet())
}

func handleModifyPetName(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchTime := reader.ReadUint32BE()
		name := reader.ReadFixedString(16)
		user := state.GetOrCreateUser(ctx.UserID)
		if catchTime > 0 {
			for i := range user.Pets {
				if user.Pets[i].CatchTime == catchTime {
					user.Pets[i].Name = name
					upsertPet(deps, user, user.Pets[i])
					break
				}
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2302, ctx.UserID, buf.Bytes())
	}
}

func handlePetStudySkill(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchTime := reader.ReadUint32BE()
		skillID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if catchTime > 0 && skillID > 0 {
			for i := range user.Pets {
				if user.Pets[i].CatchTime == catchTime {
					exists := false
					for _, sid := range user.Pets[i].Skills {
						if sid == int(skillID) {
							exists = true
							break
						}
					}
					if !exists {
						user.Pets[i].Skills = append(user.Pets[i].Skills, int(skillID))
					}
					upsertPet(deps, user, user.Pets[i])
					break
				}
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2307, ctx.UserID, buf.Bytes())
	}
}

func handlePetDefault(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchTime := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if catchTime > 0 {
			for _, p := range user.Pets {
				if p.CatchTime == catchTime {
					user.CurrentPetID = p.ID
					user.CatchID = p.CatchTime
					savePlayer(deps, ctx.UserID, user)
					break
				}
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2308, ctx.UserID, buf.Bytes())
	}
}

func handlePetOneCure(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchTime := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if catchTime > 0 {
			for i := range user.Pets {
				if user.Pets[i].CatchTime == catchTime {
					base := LoadPetDB().pets[int(user.Pets[i].ID)]
					stats := getStats(base, int(user.Pets[i].Level), int(user.Pets[i].DV), evSet{})
					user.Pets[i].HP = stats.MaxHP
					upsertPet(deps, user, user.Pets[i])
					break
				}
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2310, ctx.UserID, buf.Bytes())
	}
}

func handleIsCollect() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(1))
		ctx.Server.SendResponse(ctx.Conn, 2313, ctx.UserID, buf.Bytes())
	}
}

func handlePetHatchGet() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		for i := 0; i < 4; i++ {
			binary.Write(buf, binary.BigEndian, uint32(0))
		}
		ctx.Server.SendResponse(ctx.Conn, 2316, ctx.UserID, buf.Bytes())
	}
}

func handlePetSetExp(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchTime := reader.ReadUint32BE()
		expAmount := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if user.ExpPool >= expAmount {
			user.ExpPool -= expAmount
		} else {
			expAmount = user.ExpPool
			user.ExpPool = 0
		}
		if catchTime > 0 {
			for i := range user.Pets {
				if user.Pets[i].CatchTime == catchTime {
					user.Pets[i].Exp += int(expAmount)
					user.Pets[i].Skills = append([]int{}, user.Pets[i].Skills...)
					upsertPet(deps, user, user.Pets[i])
					break
				}
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.ExpPool)
		ctx.Server.SendResponse(ctx.Conn, 2318, ctx.UserID, buf.Bytes())
	}
}

func handlePetCollect() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2311, ctx.UserID, buf.Bytes())
	}
}

func handlePetSkillSwitch() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2312, ctx.UserID, buf.Bytes())
	}
}

func handlePetEvolution() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2314, ctx.UserID, buf.Bytes())
	}
}

func handlePetHatch() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2315, ctx.UserID, buf.Bytes())
	}
}

func handlePetRoweiList() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2320, ctx.UserID, buf.Bytes())
	}
}

func handlePetRowei() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2321, ctx.UserID, buf.Bytes())
	}
}

func handlePetRetrieve() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2322, ctx.UserID, buf.Bytes())
	}
}

func handlePetRoomShow() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2323, ctx.UserID, buf.Bytes())
	}
}

func handlePetRoomList() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2324, ctx.UserID, buf.Bytes())
	}
}

func handlePetRoomInfo() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2325, ctx.UserID, buf.Bytes())
	}
}

func handleUsePetItemOutOfFight() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2326, ctx.UserID, buf.Bytes())
	}
}

func handleUseSpeedupItem() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2327, ctx.UserID, buf.Bytes())
	}
}

func handleSkillSort() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2328, ctx.UserID, buf.Bytes())
	}
}

func handleUseAutoFightItem() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2329, ctx.UserID, buf.Bytes())
	}
}

func handleOnOffAutoFight() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2330, ctx.UserID, buf.Bytes())
	}
}

func handleUseEnergyXishou() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2331, ctx.UserID, buf.Bytes())
	}
}

func handleUseStudyItem() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2332, ctx.UserID, buf.Bytes())
	}
}

func handlePetResetNature() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2343, ctx.UserID, buf.Bytes())
	}
}

func handlePetFusion() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2351, ctx.UserID, buf.Bytes())
	}
}

func handleGetSoulBeadBuf() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2352, ctx.UserID, buf.Bytes())
	}
}

func handleSetSoulBeadBuf() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2353, ctx.UserID, buf.Bytes())
	}
}

func handleGetSoulBeadStatus() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2356, ctx.UserID, buf.Bytes())
	}
}

func handleTransformSoulBead() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2357, ctx.UserID, buf.Bytes())
	}
}

func handleSoulBeadToPet() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2358, ctx.UserID, buf.Bytes())
	}
}

func handlePetGetExp(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.ExpPool)
		ctx.Server.SendResponse(ctx.Conn, 2319, ctx.UserID, buf.Bytes())
	}
}
