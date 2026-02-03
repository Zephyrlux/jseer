package game

import (
	"bytes"
	"encoding/binary"
	"time"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerNonoHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(9001, handleNonoOpen(deps, state))
	s.Register(9002, handleNonoChangeName(deps, state))
	s.Register(9003, handleNonoInfo(state))
	s.Register(9004, handleNonoChipMixture())
	s.Register(9007, handleNonoCure(deps, state))
	s.Register(9008, handleNonoExpadm())
	s.Register(9010, handleNonoImplementTool())
	s.Register(9012, handleNonoChangeColor(deps, state))
	s.Register(9013, handleNonoPlay(deps, state))
	s.Register(9014, handleNonoCloseOpen(deps, state))
	s.Register(9015, handleNonoExeList())
	s.Register(9016, handleNonoCharge(deps, state))
	s.Register(9017, handleNonoStartExe())
	s.Register(9018, handleNonoEndExe())
	s.Register(9019, handleNonoFollowOrHoom(deps, state))
	s.Register(9020, handleNonoOpenSuper(deps, state))
	s.Register(9021, handleNonoHelpExp())
	s.Register(9022, handleNonoMateChange())
	s.Register(9023, handleNonoGetChip())
	s.Register(9024, handleNonoAddEnergyMate(deps, state))
	s.Register(9025, handleGetDiamond())
	s.Register(9026, handleNonoAddExp(deps, state))
	s.Register(9027, handleNonoIsInfo(deps, state))
	s.Register(80001, handleNieoLogin(deps, state))
}

func handleNonoOpen(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		user.Nono.HasNono = true
		user.Nono.Flag = 1
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 9001, ctx.UserID, buf.Bytes())
	}
}

func handleNonoChangeName(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		name := reader.ReadFixedString(16)
		if name == "" {
			name = "NoNo"
		}
		user := state.GetOrCreateUser(ctx.UserID)
		user.Nono.Nick = name
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 9002, ctx.UserID, []byte{})
	}
}

func handleNonoInfo(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		n := &user.Nono
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, n.Flag)
		binary.Write(buf, binary.BigEndian, n.State)
		protocol.WriteFixedString(buf, pickString(n.Nick, "NONO"), 16)
		binary.Write(buf, binary.BigEndian, n.SuperNono)
		binary.Write(buf, binary.BigEndian, pickNonZero(n.Color, 0xFFFFFF))
		binary.Write(buf, binary.BigEndian, pickNonZero(n.Power, 10000))
		binary.Write(buf, binary.BigEndian, pickNonZero(n.Mate, 10000))
		binary.Write(buf, binary.BigEndian, n.IQ)
		protocol.WriteUint16BE(buf, n.AI)
		binary.Write(buf, binary.BigEndian, pickNonZero(n.Birth, uint32(time.Now().Unix())))
		binary.Write(buf, binary.BigEndian, n.ChargeTime)
		for i := 0; i < 20; i++ {
			b := n.Func[i]
			if b == 0 {
				b = 0xFF
			}
			buf.WriteByte(b)
		}
		binary.Write(buf, binary.BigEndian, n.SuperEnergy)
		binary.Write(buf, binary.BigEndian, pickNonZero(n.SuperLevel, 0))
		stage := n.SuperStage
		if stage == 0 {
			stage = 1
		}
		binary.Write(buf, binary.BigEndian, stage)
		ctx.Server.SendResponse(ctx.Conn, 9003, ctx.UserID, buf.Bytes())
	}
}

func handleNonoChipMixture() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 9004, ctx.UserID, buf.Bytes())
	}
}

func handleNonoCure(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Nono.MaxHP == 0 {
			user.Nono.MaxHP = 10000
		}
		user.Nono.HP = user.Nono.MaxHP
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 9007, ctx.UserID, buf.Bytes())
	}
}

func handleNonoExpadm() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 9008, ctx.UserID, buf.Bytes())
	}
}

func handleNonoImplementTool() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 9010, ctx.UserID, buf.Bytes())
	}
}

func handleNonoChangeColor(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		color := reader.ReadUint32BE()
		if color == 0 {
			color = 0xFFFFFF
		}
		user := state.GetOrCreateUser(ctx.UserID)
		user.Nono.Color = color
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 9012, ctx.UserID, []byte{})
	}
}

func handleNonoPlay(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		n := &user.Nono
		if n.Mate == 0 {
			n.Mate = 10000
		}
		if n.Mate < 100000 {
			n.Mate += 5000
			if n.Mate > 100000 {
				n.Mate = 100000
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, n.Power)
		protocol.WriteUint16BE(buf, n.AI)
		binary.Write(buf, binary.BigEndian, n.Mate)
		binary.Write(buf, binary.BigEndian, n.IQ)
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 9013, ctx.UserID, buf.Bytes())
	}
}

func handleNonoCloseOpen(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		action := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		user.Nono.State = action
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 9014, ctx.UserID, []byte{})
	}
}

func handleNonoExeList() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 9015, ctx.UserID, buf.Bytes())
	}
}

func handleNonoCharge(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Nono.SuperEnergy < 99999 {
			user.Nono.SuperEnergy += 1000
			if user.Nono.SuperEnergy > 99999 {
				user.Nono.SuperEnergy = 99999
			}
		}
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 9016, ctx.UserID, []byte{})
	}
}

func handleNonoStartExe() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 9017, ctx.UserID, []byte{})
	}
}

func handleNonoEndExe() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 9018, ctx.UserID, []byte{})
	}
}

func handleNonoFollowOrHoom(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		action := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		user.NonoFollowing = action == 1
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		stage := user.Nono.SuperStage
		if stage == 0 {
			stage = 1
		}
		binary.Write(buf, binary.BigEndian, stage)
		binary.Write(buf, binary.BigEndian, action)
		if action == 1 {
			protocol.WriteFixedString(buf, pickString(user.Nono.Nick, "NONO"), 16)
			binary.Write(buf, binary.BigEndian, pickNonZero(user.Nono.Color, 0xFFFFFF))
			binary.Write(buf, binary.BigEndian, pickNonZero(user.Nono.Power, 10000))
		}
		body := buf.Bytes()
		resp := protocol.BuildResponse(9019, ctx.UserID, 0, body)
		if user.MapID > 0 {
			state.BroadcastToMap(user.MapID, resp)
		} else {
			ctx.Server.SendResponse(ctx.Conn, 9019, ctx.UserID, body)
		}
	}
}

func handleNonoOpenSuper(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		user.Nono.SuperNono = 1
		if user.Nono.SuperLevel == 0 {
			user.Nono.SuperLevel = 1
		}
		if user.Nono.SuperStage == 0 {
			user.Nono.SuperStage = 1
		}
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 9020, ctx.UserID, []byte{})
	}
}

func handleNonoHelpExp() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 9021, ctx.UserID, []byte{})
	}
}

func handleNonoMateChange() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 9022, ctx.UserID, []byte{})
	}
}

func handleNonoGetChip() gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		chipType := reader.ReadUint32BE()
		if chipType == 0 {
			chipType = 1
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(1))
		binary.Write(buf, binary.BigEndian, chipType)
		binary.Write(buf, binary.BigEndian, uint32(1))
		ctx.Server.SendResponse(ctx.Conn, 9023, ctx.UserID, buf.Bytes())
	}
}

func handleNonoAddEnergyMate(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Nono.Power < 100000 {
			user.Nono.Power += 10000
			if user.Nono.Power > 100000 {
				user.Nono.Power = 100000
			}
		}
		if user.Nono.Mate < 100000 {
			user.Nono.Mate += 10000
			if user.Nono.Mate > 100000 {
				user.Nono.Mate = 100000
			}
		}
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 9024, ctx.UserID, []byte{})
	}
}

func handleGetDiamond() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(9999))
		ctx.Server.SendResponse(ctx.Conn, 9025, ctx.UserID, buf.Bytes())
	}
}

func handleNonoAddExp(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Nono.SuperLevel < 100 {
			user.Nono.SuperLevel++
		}
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 9026, ctx.UserID, []byte{})
	}
}

func handleNonoIsInfo(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		n := &user.Nono
		now := uint32(time.Now().Unix())
		if n.VipEndTime > 0 && n.VipEndTime < now {
			n.SuperNono = 0
		}
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(1))
		ctx.Server.SendResponse(ctx.Conn, 9027, ctx.UserID, buf.Bytes())
	}
}

func handleNieoLogin(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		n := &user.Nono
		now := uint32(time.Now().Unix())
		needActivate := n.SuperNono == 0 || (n.VipEndTime > 0 && n.VipEndTime < now)
		if needActivate {
			endTime := now + 30*24*60*60
			n.SuperNono = 1
			n.VipEndTime = endTime
			if n.SuperLevel == 0 {
				n.SuperLevel = 1
			}
			if n.SuperStage == 0 {
				n.SuperStage = 1
			}

			msg := "成功激活超能NONO！\n到期时间:" + time.Unix(int64(endTime), 0).Format("2006-01-02")
			msgBuf := new(bytes.Buffer)
			binary.Write(msgBuf, binary.BigEndian, uint32(len(msg)))
			msgBuf.WriteString(msg)
			ctx.Server.SendResponse(ctx.Conn, 80002, ctx.UserID, msgBuf.Bytes())

			vipBuf := new(bytes.Buffer)
			binary.Write(vipBuf, binary.BigEndian, ctx.UserID)
			binary.Write(vipBuf, binary.BigEndian, uint32(2))
			binary.Write(vipBuf, binary.BigEndian, n.AutoCharge)
			binary.Write(vipBuf, binary.BigEndian, endTime)
			ctx.Server.SendResponse(ctx.Conn, 8006, ctx.UserID, vipBuf.Bytes())
		}
		savePlayer(deps, ctx.UserID, user)

		statusBuf := new(bytes.Buffer)
		binary.Write(statusBuf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 80001, ctx.UserID, statusBuf.Bytes())
	}
}

func pickString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
