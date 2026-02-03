package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
)

func registerExchangeHandlers(s *gateway.Server) {
	s.Register(2902, handleExchangePetComplete())
	s.Register(2251, handleExchangeOre())
	s.Register(2065, handleExchangeNewYear())
	s.Register(2701, handleTalkCount())
	s.Register(2702, handleTalkCate())
}

func handleExchangePetComplete() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2902, ctx.UserID, buf.Bytes())
	}
}

func handleExchangeOre() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2251, ctx.UserID, buf.Bytes())
	}
}

func handleExchangeNewYear() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2065, ctx.UserID, buf.Bytes())
	}
}

func handleTalkCount() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2701, ctx.UserID, buf.Bytes())
	}
}

func handleTalkCate() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2702, ctx.UserID, buf.Bytes())
	}
}
