package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
)

func registerTeamPKHandlers(s *gateway.Server) {
	s.Register(4001, handleTeamPKSign())
	s.Register(4002, handleTeamPKRegister())
	s.Register(4003, handleTeamPKJoin())
	s.Register(4004, handleTeamPKShot())
	s.Register(4005, handleTeamPKRefreshDistance())
	s.Register(4006, handleTeamPKWin())
	s.Register(4007, handleTeamPKNote())
	s.Register(4008, handleTeamPKFreeze())
	s.Register(4009, handleTeamPKUnfreeze())
	s.Register(4010, handleTeamPKBeShot())
	s.Register(4011, handleTeamPKGetBuildingInfo())
	s.Register(4012, handleTeamPKSituation())
	s.Register(4013, handleTeamPKResult())
	s.Register(4014, handleTeamPKUseShield())
	s.Register(4017, handleTeamPKWeekyScore())
	s.Register(4018, handleTeamPKHistory())
	s.Register(4019, handleTeamPKSomeoneJoinInfo())
	s.Register(4020, handleTeamPKNoPet())
	s.Register(4022, handleTeamPKActive())
	s.Register(4023, handleTeamPKActiveNoteGetItem())
	s.Register(4024, handleTeamPKActiveGetAttack())
	s.Register(4025, handleTeamPKActiveGetStone())
	s.Register(4101, handleTeamPKTeamCharts())
	s.Register(4102, handleTeamPKSeerCharts())
	s.Register(2481, handleTeamPKPetFight())
}

func handleTeamPKSign() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		buf.Write(make([]byte, 24))
		binary.Write(buf, binary.BigEndian, uint32(0x7F000001))
		binary.Write(buf, binary.BigEndian, uint16(5100))
		ctx.Server.SendResponse(ctx.Conn, 4001, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKRegister() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4002, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKJoin() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4003, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKShot() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4004, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKRefreshDistance() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4005, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKWin() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4006, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKNote() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4007, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKFreeze() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4008, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKUnfreeze() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4009, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKBeShot() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4010, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKGetBuildingInfo() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4011, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKSituation() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4012, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKResult() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4013, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKUseShield() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4014, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKWeekyScore() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4017, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKHistory() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4018, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKSomeoneJoinInfo() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4019, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKNoPet() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4020, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKActive() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4022, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKActiveNoteGetItem() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4023, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKActiveGetAttack() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4024, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKActiveGetStone() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4025, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKTeamCharts() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4101, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKSeerCharts() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4102, ctx.UserID, buf.Bytes())
	}
}

func handleTeamPKPetFight() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2481, ctx.UserID, buf.Bytes())
	}
}
