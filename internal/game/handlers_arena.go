package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerArenaHandlers(s *gateway.Server) {
	s.Register(2414, handleChoiceFightLevel())
	s.Register(2415, handleStartFightLevel())
	s.Register(2416, handleLeaveFightLevel())
	s.Register(2417, handleArenaSetOwner())
	s.Register(2418, handleArenaFightOwner())
	s.Register(2419, handleArenaGetInfo())
	s.Register(2420, handleArenaUpfight())
	s.Register(2421, handleFightSpecialPet())
	s.Register(2422, handleArenaOwnerAcce())
	s.Register(2423, handleArenaOwnerOut())
	s.Register(2424, handleOpenDarkportal())
	s.Register(2425, handleFightDarkportal())
	s.Register(2426, handleLeaveDarkportal())
	s.Register(2428, handleFreshChoiceFightLevel())
	s.Register(2429, handleFreshStartFightLevel())
	s.Register(2430, handleFreshLeaveFightLevel())
}

func handleChoiceFightLevel() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2414, ctx.UserID, buf.Bytes())
	}
}

func handleStartFightLevel() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2415, ctx.UserID, buf.Bytes())
	}
}

func handleLeaveFightLevel() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2416, ctx.UserID, []byte{})
	}
}

func handleArenaSetOwner() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2417, ctx.UserID, buf.Bytes())
	}
}

func handleArenaFightOwner() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2418, ctx.UserID, buf.Bytes())
	}
}

func handleArenaGetInfo() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		protocol.WriteFixedString(buf, "", 16)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2419, ctx.UserID, buf.Bytes())
	}
}

func handleArenaUpfight() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2420, ctx.UserID, buf.Bytes())
	}
}

func handleFightSpecialPet() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2421, ctx.UserID, buf.Bytes())
	}
}

func handleArenaOwnerAcce() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2422, ctx.UserID, buf.Bytes())
	}
}

func handleArenaOwnerOut() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2423, ctx.UserID, buf.Bytes())
	}
}

func handleOpenDarkportal() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2424, ctx.UserID, buf.Bytes())
	}
}

func handleFightDarkportal() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2425, ctx.UserID, buf.Bytes())
	}
}

func handleLeaveDarkportal() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2426, ctx.UserID, []byte{})
	}
}

func handleFreshChoiceFightLevel() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2428, ctx.UserID, buf.Bytes())
	}
}

func handleFreshStartFightLevel() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2429, ctx.UserID, buf.Bytes())
	}
}

func handleFreshLeaveFightLevel() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2430, ctx.UserID, []byte{})
	}
}
