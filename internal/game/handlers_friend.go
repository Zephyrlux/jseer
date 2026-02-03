package game

import (
	"bytes"
	"encoding/binary"
	"time"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerFriendHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2150, handleGetRelationList(state))
	s.Register(2151, handleFriendAdd(deps, state))
	s.Register(2152, handleFriendAnswer(deps, state))
	s.Register(2153, handleFriendRemove(deps, state))
	s.Register(2154, handleBlackAdd(deps, state))
	s.Register(2155, handleBlackRemove(deps, state))
	s.Register(2157, handleSeeOnline(state))
	s.Register(2158, handleRequestOut())
	s.Register(2159, handleRequestAnswer())
}

func handleFriendAdd(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		if targetID == 0 || targetID == ctx.UserID {
			resp := protocol.BuildResponse(2151, ctx.UserID, 1, []byte{})
			_, _ = ctx.Conn.Write(resp)
			return
		}
		user := state.GetOrCreateUser(ctx.UserID)
		found := false
		for _, f := range user.Friends {
			if f.UserID == targetID {
				found = true
				break
			}
		}
		if !found {
			user.Friends = append(user.Friends, FriendInfo{
				UserID:   targetID,
				TimePoke: uint32(time.Now().Unix()),
			})
			savePlayer(deps, ctx.UserID, user)
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, targetID)
		ctx.Server.SendResponse(ctx.Conn, 2151, ctx.UserID, buf.Bytes())
	}
}

func handleFriendAnswer(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		accept := reader.ReadUint32BE()
		if accept == 1 && targetID > 0 {
			user := state.GetOrCreateUser(ctx.UserID)
			found := false
			for _, f := range user.Friends {
				if f.UserID == targetID {
					found = true
					break
				}
			}
			if !found {
				user.Friends = append(user.Friends, FriendInfo{
					UserID:   targetID,
					TimePoke: uint32(time.Now().Unix()),
				})
				savePlayer(deps, ctx.UserID, user)
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, accept)
		ctx.Server.SendResponse(ctx.Conn, 2152, ctx.UserID, buf.Bytes())
	}
}

func handleFriendRemove(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		next := user.Friends[:0]
		for _, f := range user.Friends {
			if f.UserID != targetID {
				next = append(next, f)
			}
		}
		user.Friends = next
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 2153, ctx.UserID, []byte{})
	}
}

func handleBlackAdd(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		if targetID == 0 || targetID == ctx.UserID {
			resp := protocol.BuildResponse(2154, ctx.UserID, 1, []byte{})
			_, _ = ctx.Conn.Write(resp)
			return
		}
		user := state.GetOrCreateUser(ctx.UserID)
		// remove from friends
		friends := user.Friends[:0]
		for _, f := range user.Friends {
			if f.UserID != targetID {
				friends = append(friends, f)
			}
		}
		user.Friends = friends
		// add to blacklist
		found := false
		for _, id := range user.Blacklist {
			if id == targetID {
				found = true
				break
			}
		}
		if !found {
			user.Blacklist = append(user.Blacklist, targetID)
		}
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, targetID)
		ctx.Server.SendResponse(ctx.Conn, 2154, ctx.UserID, buf.Bytes())
	}
}

func handleBlackRemove(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		next := user.Blacklist[:0]
		for _, id := range user.Blacklist {
			if id != targetID {
				next = append(next, id)
			}
		}
		user.Blacklist = next
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 2155, ctx.UserID, []byte{})
	}
}

func handleSeeOnline(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		count := int(reader.ReadUint32BE())
		ids := make([]uint32, 0, count)
		for i := 0; i < count; i++ {
			ids = append(ids, reader.ReadUint32BE())
		}
		buf := new(bytes.Buffer)
		online := make([]uint32, 0, len(ids))
		for _, id := range ids {
			if _, ok := state.GetConn(id); !ok {
				continue
			}
			online = append(online, id)
		}
		binary.Write(buf, binary.BigEndian, uint32(len(online)))
		for _, id := range online {
			user := state.GetOrCreateUser(id)
			binary.Write(buf, binary.BigEndian, id)
			binary.Write(buf, binary.BigEndian, uint32(1))
			binary.Write(buf, binary.BigEndian, user.MapType)
			binary.Write(buf, binary.BigEndian, user.MapID)
		}
		ctx.Server.SendResponse(ctx.Conn, 2157, ctx.UserID, buf.Bytes())
	}
}

func handleRequestOut() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2158, ctx.UserID, []byte{})
	}
}

func handleRequestAnswer() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2159, ctx.UserID, []byte{})
	}
}
