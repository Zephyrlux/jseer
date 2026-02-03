package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
)

func registerAchievementHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(3401, handleAchievementList(state))
	s.Register(3402, handleAchievementInfo())
	s.Register(3403, handleAchievementTitleList(state))
	s.Register(3404, handleSetTitle(deps, state))
	s.Register(3406, handleConferAchievement(deps, state))
	s.Register(3407, handleAchieveAndTitle(state))
}

func handleAchievementList(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(len(user.Achievements)))
		for _, id := range user.Achievements {
			binary.Write(buf, binary.BigEndian, id)
		}
		ctx.Server.SendResponse(ctx.Conn, 3401, ctx.UserID, buf.Bytes())
	}
}

func handleAchievementInfo() gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		achieveID := reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, achieveID)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 3402, ctx.UserID, buf.Bytes())
	}
}

func handleAchievementTitleList(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(len(user.Titles)))
		for _, id := range user.Titles {
			binary.Write(buf, binary.BigEndian, id)
		}
		ctx.Server.SendResponse(ctx.Conn, 3403, ctx.UserID, buf.Bytes())
	}
}

func handleSetTitle(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		titleID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		user.CurTitle = titleID
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 3404, ctx.UserID, buf.Bytes())
	}
}

func handleConferAchievement(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		achieveID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if achieveID > 0 && !containsUint32(user.Achievements, achieveID) {
			user.Achievements = append(user.Achievements, achieveID)
			savePlayer(deps, ctx.UserID, user)
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 3406, ctx.UserID, buf.Bytes())
	}
}

func handleAchieveAndTitle(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(len(user.Achievements)))
		for _, id := range user.Achievements {
			binary.Write(buf, binary.BigEndian, id)
		}
		binary.Write(buf, binary.BigEndian, uint32(len(user.Titles)))
		for _, id := range user.Titles {
			binary.Write(buf, binary.BigEndian, id)
		}
		ctx.Server.SendResponse(ctx.Conn, 3407, ctx.UserID, buf.Bytes())
	}
}

func containsUint32(list []uint32, val uint32) bool {
	for _, v := range list {
		if v == val {
			return true
		}
	}
	return false
}
