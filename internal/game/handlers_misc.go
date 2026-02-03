package game

import (
	"bytes"
	"encoding/binary"
	"time"

	"jseer/internal/gateway"
)

func registerMiscHandlers(s *gateway.Server) {
	s.Register(50004, handleXinCheck())
	s.Register(50008, handleXinGetQuadrupleExeTime())
	s.Register(8005, handleSyncTime())
	s.Register(70001, handleGetExchangeInfo())
}

func handleXinCheck() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 50004, ctx.UserID, []byte{})
	}
}

func handleXinGetQuadrupleExeTime() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 50008, ctx.UserID, buf.Bytes())
	}
}

func handleGetExchangeInfo() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 70001, ctx.UserID, buf.Bytes())
	}
}

func handleSyncTime() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(time.Now().Unix()))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 8005, ctx.UserID, buf.Bytes())
	}
}
