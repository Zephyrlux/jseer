package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
)

func registerGameHandlers(s *gateway.Server) {
	s.Register(5001, handleJoinGame())
	s.Register(5002, handleGameOver())
	s.Register(5003, handleLeaveGame())
	s.Register(5052, handleFBGameOver())
	s.Register(3201, handleEggGamePlay())
	s.Register(2442, handleMLFigBoss())
	s.Register(2444, handleMLStateBoss())
	s.Register(2445, handleMLStepPos())
	s.Register(2446, handleMLGetPrize())
}

func handleJoinGame() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 5001, ctx.UserID, buf.Bytes())
	}
}

func handleGameOver() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 5002, ctx.UserID, buf.Bytes())
	}
}

func handleLeaveGame() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 5003, ctx.UserID, []byte{})
	}
}

func handleFBGameOver() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 5052, ctx.UserID, buf.Bytes())
	}
}

func handleEggGamePlay() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 3201, ctx.UserID, buf.Bytes())
	}
}

func handleMLFigBoss() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2442, ctx.UserID, buf.Bytes())
	}
}

func handleMLStateBoss() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2444, ctx.UserID, buf.Bytes())
	}
}

func handleMLStepPos() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2445, ctx.UserID, buf.Bytes())
	}
}

func handleMLGetPrize() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2446, ctx.UserID, buf.Bytes())
	}
}
