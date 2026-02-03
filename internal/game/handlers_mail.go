package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerMailHandlers(s *gateway.Server) {
	s.Register(2751, handleMailGetList())
	s.Register(2752, handleMailSend())
	s.Register(2753, handleMailGetContent())
	s.Register(2754, handleMailSetRead())
	s.Register(2755, handleMailDelete())
	s.Register(2756, handleMailDeleteAll())
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

func handleMailSend() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2752, ctx.UserID, buf.Bytes())
	}
}

func handleMailGetContent() gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		mailID := reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, mailID)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2753, ctx.UserID, buf.Bytes())
	}
}

func handleMailSetRead() gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		mailID := reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, mailID)
		ctx.Server.SendResponse(ctx.Conn, 2754, ctx.UserID, buf.Bytes())
	}
}

func handleMailDelete() gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		mailID := reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, mailID)
		ctx.Server.SendResponse(ctx.Conn, 2755, ctx.UserID, buf.Bytes())
	}
}

func handleMailDeleteAll() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2756, ctx.UserID, buf.Bytes())
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
