package game

import (
	"bytes"
	"encoding/binary"
	"time"

	"jseer/internal/gateway"
)

func registerCompatHandlers(s *gateway.Server) {
	s.Register(1022, handleCheckFightCode())
	s.Register(10301, handleServerTime())
	s.Register(2231, handleAcceptDailyTask())
	s.Register(41080, handleGetForeverValue())
	s.Register(4475, handleGetItemListLegacy())
	s.Register(47334, handleLegacyFriendList())
	s.Register(47335, handleLegacyBlacklist())
	s.Register(9049, handleOpenBagGet())
	s.Register(11003, handleGetPetInfoLegacy())
	s.Register(11007, handleGetPetByCatchTimeLegacy())
	s.Register(11022, handleGetSecondBag())
	s.Register(40001, handleGetSuperValue())
	s.Register(40002, handleGetSuperValueByIDs())
	s.Register(41983, handleRelogin())
	s.Register(42023, handleBatchGetBitset())
	s.Register(46046, handleGetMultiForever())
	s.Register(46057, handleGetMultiForeverByDB())
}

func handleCheckFightCode() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 1022, ctx.UserID, []byte{})
	}
}

func handleServerTime() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(time.Now().Unix()))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 10301, ctx.UserID, buf.Bytes())
	}
}

func handleAcceptDailyTask() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2231, ctx.UserID, buf.Bytes())
	}
}

func handleGetForeverValue() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 41080, ctx.UserID, buf.Bytes())
	}
}

func handleGetItemListLegacy() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 4475, ctx.UserID, buf.Bytes())
	}
}

func handleLegacyFriendList() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 47334, ctx.UserID, buf.Bytes())
	}
}

func handleLegacyBlacklist() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 47335, ctx.UserID, buf.Bytes())
	}
}

func handleOpenBagGet() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 9049, ctx.UserID, buf.Bytes())
	}
}

func handleGetPetInfoLegacy() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 11003, ctx.UserID, buf.Bytes())
	}
}

func handleGetPetByCatchTimeLegacy() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 11007, ctx.UserID, []byte{})
	}
}

func handleGetSecondBag() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 11022, ctx.UserID, buf.Bytes())
	}
}

func handleGetSuperValue() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 40001, ctx.UserID, buf.Bytes())
	}
}

func handleGetSuperValueByIDs() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 40002, ctx.UserID, buf.Bytes())
	}
}

func handleRelogin() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 41983, ctx.UserID, []byte{})
	}
}

func handleBatchGetBitset() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 42023, ctx.UserID, buf.Bytes())
	}
}

func handleGetMultiForever() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(5))
		for i := 0; i < 5; i++ {
			binary.Write(buf, binary.BigEndian, uint32(0))
		}
		ctx.Server.SendResponse(ctx.Conn, 46046, ctx.UserID, buf.Bytes())
	}
}

func handleGetMultiForeverByDB() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 46057, ctx.UserID, buf.Bytes())
	}
}
