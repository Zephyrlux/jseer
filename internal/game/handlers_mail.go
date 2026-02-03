package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerMailHandlers(s *gateway.Server) {
	s.Register(2751, handleMailGetList())
	s.Register(2757, handleMailGetUnread())
	s.Register(8001, handleInform())
	s.Register(8004, handleGetBossMonster())
}

func handleMailGetList() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2751, ctx.UserID, buf.Bytes())
	}
}

func handleMailGetUnread() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2757, ctx.UserID, buf.Bytes())
	}
}

func handleInform() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		protocol.WriteFixedString(buf, "", 16)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(1))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(301))
		protocol.WriteFixedString(buf, "", 64)
		ctx.Server.SendResponse(ctx.Conn, 8001, ctx.UserID, buf.Bytes())
	}
}

func handleGetBossMonster() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		for i := 0; i < 4; i++ {
			binary.Write(buf, binary.BigEndian, uint32(0))
		}
		ctx.Server.SendResponse(ctx.Conn, 8004, ctx.UserID, buf.Bytes())
	}
}
