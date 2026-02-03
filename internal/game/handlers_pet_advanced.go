package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
)

func registerPetAdvancedHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2302, handleModifyPetName(state))
	s.Register(2307, handlePetStudySkill(state))
	s.Register(2308, handlePetDefault(state))
	s.Register(2310, handlePetOneCure(state))
	s.Register(2313, handleIsCollect())
	s.Register(2316, handlePetHatchGet())
	s.Register(2318, handlePetSetExp(state))
	s.Register(2319, handlePetGetExp(state))
}

func handleModifyPetName(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchTime := reader.ReadUint32BE()
		name := reader.ReadFixedString(16)
		user := state.GetOrCreateUser(ctx.UserID)
		if catchTime > 0 {
			for i := range user.Pets {
				if user.Pets[i].CatchTime == catchTime {
					user.Pets[i].Name = name
					break
				}
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2302, ctx.UserID, buf.Bytes())
	}
}

func handlePetStudySkill(state *State) gateway.Handler {
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
					break
				}
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2307, ctx.UserID, buf.Bytes())
	}
}

func handlePetDefault(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchTime := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if catchTime > 0 {
			for _, p := range user.Pets {
				if p.CatchTime == catchTime {
					user.CurrentPetID = p.ID
					user.CatchID = p.CatchTime
					break
				}
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2308, ctx.UserID, buf.Bytes())
	}
}

func handlePetOneCure(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		catchTime := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if catchTime > 0 {
			for i := range user.Pets {
				if user.Pets[i].CatchTime == catchTime {
					user.Pets[i].HP = 100
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

func handlePetSetExp(state *State) gateway.Handler {
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
					break
				}
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.ExpPool)
		ctx.Server.SendResponse(ctx.Conn, 2318, ctx.UserID, buf.Bytes())
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
