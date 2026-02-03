package game

import (
	"bytes"
	"encoding/binary"
	"time"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerRoomHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(10001, handleRoomLogin(deps, state))
	s.Register(10002, handleGetRoomAddress(deps))
	s.Register(10003, handleLeaveRoom(deps, state))
}

func handleRoomLogin(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		_ = reader.ReadBytes(24) // session
		_ = reader.ReadUint32BE() // catchTime
		_ = reader.ReadUint32BE() // flag
		targetID := ctx.UserID
		if reader.Remaining() >= 4 {
			targetID = reader.ReadUint32BE()
		}
		x := uint32(300)
		y := uint32(300)
		if reader.Remaining() >= 8 {
			x = reader.ReadUint32BE()
			y = reader.ReadUint32BE()
		}
		if x == 0 && y == 0 {
			x, y = 300, 300
		}

		_ = targetID
		user := state.GetOrCreateUser(ctx.UserID)
		user.MapType = 1
		user.PosX = x
		user.PosY = y
		user.LastMapID = user.MapID
		state.UpdatePlayerMap(ctx.UserID, 500001)
		savePlayer(deps, ctx.UserID, user)

		body := buildPeopleInfo(ctx.UserID, user, uint32(time.Now().Unix()))
		ctx.Server.SendResponse(ctx.Conn, 2001, ctx.UserID, body)
	}
}

func handleGetRoomAddress(deps *Deps) gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		protocol.WriteFixedString(buf, "", 24)
		ip := deps.GameIP
		if ip == "" {
			ip = "127.0.0.1"
		}
		writeIP(buf, ip)
		port := uint16(5000)
		if deps.GamePort > 0 {
			port = uint16(deps.GamePort)
		}
		binary.Write(buf, binary.BigEndian, port)
		ctx.Server.SendResponse(ctx.Conn, 10002, ctx.UserID, buf.Bytes())
	}
}

func handleLeaveRoom(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		mapID := user.LastMapID
		if mapID == 0 {
			mapID = 1
		}
		state.UpdatePlayerMap(ctx.UserID, mapID)
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 10003, ctx.UserID, []byte{})
	}
}
