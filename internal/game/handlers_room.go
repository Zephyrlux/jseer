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
	s.Register(10004, handleBuyFitment(deps, state))
	s.Register(10005, handleBetrayFitment(deps, state))
	s.Register(10006, handleFitmentUsing(state))
	s.Register(10007, handleFitmentAll(state))
	s.Register(10008, handleSetFitment(deps, state))
	s.Register(10009, handleAddEnergy(state))
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
		if user.RoomID == 0 {
			user.RoomID = ctx.UserID
		}
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

func handleBuyFitment(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		itemID := int(reader.ReadUint32BE())
		count := int(reader.ReadUint32BE())
		if count <= 0 {
			count = 1
		}
		user := state.GetOrCreateUser(ctx.UserID)
		price := uint32(100 * count)
		if user.Coins >= price {
			user.Coins -= price
		}
		if user.Items == nil {
			user.Items = make(map[int]*ItemInfo)
		}
		info := user.Items[itemID]
		if info == nil {
			info = &ItemInfo{Count: 0, ExpireTime: defaultItemExpire}
			user.Items[itemID] = info
		}
		info.Count += count
		upsertItem(deps, user, itemID)
		savePlayer(deps, ctx.UserID, user)

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.Coins)
		binary.Write(buf, binary.BigEndian, uint32(itemID))
		binary.Write(buf, binary.BigEndian, uint32(count))
		ctx.Server.SendResponse(ctx.Conn, 10004, ctx.UserID, buf.Bytes())
	}
}

func handleBetrayFitment(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		itemID := int(reader.ReadUint32BE())
		count := int(reader.ReadUint32BE())
		if count <= 0 {
			count = 1
		}
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Items != nil {
			if info := user.Items[itemID]; info != nil {
				if info.Count < count {
					count = info.Count
				}
				info.Count -= count
				if info.Count <= 0 {
					delete(user.Items, itemID)
				}
				upsertItem(deps, user, itemID)
			}
		}
		user.Coins += uint32(50 * count)
		savePlayer(deps, ctx.UserID, user)

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.Coins)
		binary.Write(buf, binary.BigEndian, uint32(itemID))
		binary.Write(buf, binary.BigEndian, uint32(count))
		ctx.Server.SendResponse(ctx.Conn, 10005, ctx.UserID, buf.Bytes())
	}
}

func handleFitmentUsing(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := ctx.UserID
		if reader.Remaining() >= 4 {
			targetID = reader.ReadUint32BE()
		}
		user := state.GetOrCreateUser(targetID)
		roomID := user.RoomID
		if roomID == 0 {
			roomID = targetID
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, targetID)
		binary.Write(buf, binary.BigEndian, roomID)
		binary.Write(buf, binary.BigEndian, uint32(len(user.Fitments)))
		for _, f := range user.Fitments {
			binary.Write(buf, binary.BigEndian, f.ID)
			binary.Write(buf, binary.BigEndian, f.X)
			binary.Write(buf, binary.BigEndian, f.Y)
			binary.Write(buf, binary.BigEndian, f.Dir)
			binary.Write(buf, binary.BigEndian, f.Status)
		}
		ctx.Server.SendResponse(ctx.Conn, 10006, ctx.UserID, buf.Bytes())
	}
}

func handleFitmentAll(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		stats := map[int]struct {
			Used int
			Bag  int
		}{}
		for _, f := range user.Fitments {
			id := int(f.ID)
			if id == 0 {
				continue
			}
			s := stats[id]
			s.Used++
			stats[id] = s
		}
		for itemID, info := range user.Items {
			if itemID < 500000 {
				continue
			}
			if info == nil || info.Count <= 0 {
				continue
			}
			s := stats[itemID]
			s.Bag += info.Count
			stats[itemID] = s
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(len(stats)))
		for id, s := range stats {
			binary.Write(buf, binary.BigEndian, uint32(id))
			binary.Write(buf, binary.BigEndian, uint32(s.Used))
			binary.Write(buf, binary.BigEndian, uint32(s.Used+s.Bag))
		}
		ctx.Server.SendResponse(ctx.Conn, 10007, ctx.UserID, buf.Bytes())
	}
}

func handleSetFitment(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		roomID := reader.ReadUint32BE()
		count := int(reader.ReadUint32BE())
		newFitments := make([]Fitment, 0, count)
		for i := 0; i < count; i++ {
			if reader.Remaining() < 20 {
				break
			}
			newFitments = append(newFitments, Fitment{
				ID:     reader.ReadUint32BE(),
				X:      reader.ReadUint32BE(),
				Y:      reader.ReadUint32BE(),
				Dir:    reader.ReadUint32BE(),
				Status: reader.ReadUint32BE(),
			})
		}
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Items == nil {
			user.Items = make(map[int]*ItemInfo)
		}
		oldCounts := map[uint32]int{}
		for _, f := range user.Fitments {
			oldCounts[f.ID]++
		}
		newCounts := map[uint32]int{}
		for _, f := range newFitments {
			newCounts[f.ID]++
		}
		for id, old := range oldCounts {
			newc := newCounts[id]
			delta := newc - old
			if delta == 0 {
				continue
			}
			itemID := int(id)
			info := user.Items[itemID]
			if info == nil {
				info = &ItemInfo{Count: 0, ExpireTime: defaultItemExpire}
				user.Items[itemID] = info
			}
			info.Count -= delta
			if info.Count < 0 {
				info.Count = 0
			}
			upsertItem(deps, user, itemID)
		}
		user.Fitments = newFitments
		user.RoomID = roomID
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 10008, ctx.UserID, []byte{})
	}
}

func handleAddEnergy(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, user.Energy)
		ctx.Server.SendResponse(ctx.Conn, 10009, ctx.UserID, buf.Bytes())
	}
}
