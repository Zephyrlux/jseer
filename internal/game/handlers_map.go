package game

import (
	"bytes"
	"encoding/binary"
	"sort"
	"time"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerMapHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2000, handleOnMapSwitch())
	s.Register(2001, handleEnterMap(deps, state))
	s.Register(2002, handleLeaveMap(deps, state))
	s.Register(2003, handleListMapPlayer(state))
	s.Register(2004, handleMapOgreList(state))
	s.Register(2021, handleMapBoss(state))
	s.Register(2022, handleSpecialPetNote())
	s.Register(2023, handleOfflineExp(state))
	s.Register(2051, handleGetSimUserInfo(state))
	s.Register(2052, handleGetMoreUserInfo(state))
	s.Register(2053, handleRequestCount(state))
	s.Register(2062, handleChangeDoodle(deps, state))
	s.Register(2064, handleGetRequestAward(deps, state))
	s.Register(2061, handleChangeNickName(deps, state))
	s.Register(2063, handleChangeColor(deps, state))
	s.Register(2101, handlePeopleWalk(state))
	s.Register(2102, handleChat(state))
	s.Register(2103, handleDanceAction(state))
	s.Register(2104, handleAimat(state))
	s.Register(2105, handleHitStone())
	s.Register(2106, handlePrizeOfAtresiaSpace())
	s.Register(2107, handleTransformUser(state))
	s.Register(2109, handleAttackBailuen())
	s.Register(2110, handleGetTimePoke())
	s.Register(2111, handlePeopleTransform(state))
	s.Register(2112, handleOnOrOffFlying(deps, state))
	s.Register(2113, handleRemoveCoins(deps, state))
}

func handleOnMapSwitch() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2000, ctx.UserID, []byte{})
	}
}

func handleMapBoss(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		mapID := int(user.MapID)
		reader := NewReader(ctx.Body)
		if reader.Remaining() >= 4 {
			mapID = int(reader.ReadUint32BE())
		}
		body := buildMapBossList(mapID)
		ctx.Server.SendResponse(ctx.Conn, 2021, ctx.UserID, body)
	}
}

func handleSpecialPetNote() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2022, ctx.UserID, buf.Bytes())
	}
}

func handleOfflineExp(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.ExpPool)
		ctx.Server.SendResponse(ctx.Conn, 2023, ctx.UserID, buf.Bytes())
	}
}

func buildMapBossList(mapID int) []byte {
	entries := getMapBossEntries(mapID)
	if len(entries) == 0 {
		return make([]byte, 4)
	}
	regions := make([]int, 0, len(entries))
	for region := range entries {
		regions = append(regions, int(region))
	}
	sort.Ints(regions)

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(len(regions)))
	for _, r := range regions {
		region := uint32(r)
		entry := entries[region]
		hp := uint32(0)
		if entry.HasShield {
			if boss := GetSPTBossByID(entry.BossPetID); boss != nil && boss.MaxHP > 0 {
				hp = uint32(boss.MaxHP)
			}
		}
		binary.Write(buf, binary.BigEndian, uint32(entry.BossPetID))
		binary.Write(buf, binary.BigEndian, region)
		binary.Write(buf, binary.BigEndian, hp)
		binary.Write(buf, binary.BigEndian, uint32(0))
	}
	return buf.Bytes()
}

func handleChangeDoodle(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		doodleID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		user.Texture = doodleID
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, user.Color)
		binary.Write(buf, binary.BigEndian, user.Texture)
		binary.Write(buf, binary.BigEndian, user.Coins)
		ctx.Server.SendResponse(ctx.Conn, 2062, ctx.UserID, buf.Bytes())
	}
}

func handleGetRequestAward(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Items == nil {
			user.Items = make(map[int]*ItemInfo)
		}
		rewardIDs := []int{100073, 100074, 100075}
		for _, itemID := range rewardIDs {
			if isUniqueItem(itemID) && user.Items[itemID] != nil {
				continue
			}
			info := user.Items[itemID]
			if info == nil {
				info = &ItemInfo{Count: 0, ExpireTime: defaultItemExpire}
				user.Items[itemID] = info
			}
			info.Count++
			upsertItem(deps, user, itemID)
		}
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 2064, ctx.UserID, []byte{})
	}
}

func handleRequestCount(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		if targetID == 0 {
			targetID = ctx.UserID
		}
		var count uint32
		if state != nil {
			_ = state.GetOrCreateUser(targetID)
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, targetID)
		binary.Write(buf, binary.BigEndian, count)
		ctx.Server.SendResponse(ctx.Conn, 2053, ctx.UserID, buf.Bytes())
	}
}

func handleEnterMap(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		mapType := reader.ReadUint32BE()
		mapID := reader.ReadUint32BE()
		x := reader.ReadUint32BE()
		y := reader.ReadUint32BE()
		if x == 0 && y == 0 {
			x, y = 300, 270
		}

		user := state.GetOrCreateUser(ctx.UserID)
		if mapID == 0 {
			mapID = user.MapID
		}
		user.MapType = mapType
		user.PosX = x
		user.PosY = y
		user.LastMapID = user.MapID
		state.UpdatePlayerMap(ctx.UserID, mapID)
		savePlayer(deps, ctx.UserID, user)

		body := buildPeopleInfo(ctx.UserID, user, uint32(time.Now().Unix()))
		ctx.Server.SendResponse(ctx.Conn, 2001, ctx.UserID, body)

		// Send map player list (including self)
		listBuf := new(bytes.Buffer)
		players := state.GetPlayersInMap(mapID)
		binary.Write(listBuf, binary.BigEndian, uint32(len(players)))
		for _, pid := range players {
			pUser := state.GetOrCreateUser(pid)
			info := buildPeopleInfo(pid, pUser, uint32(time.Now().Unix()))
			listBuf.Write(info)
		}
		ctx.Server.SendResponse(ctx.Conn, 2003, ctx.UserID, listBuf.Bytes())
	}
}

func handleLeaveMap(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		state.UpdatePlayerMap(ctx.UserID, 0)
		savePlayer(deps, ctx.UserID, state.GetOrCreateUser(ctx.UserID))
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		ctx.Server.SendResponse(ctx.Conn, 2002, ctx.UserID, buf.Bytes())
	}
}

func handleListMapPlayer(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		mapID := user.MapID
		players := state.GetPlayersInMap(mapID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(len(players)))
		for _, pid := range players {
			pUser := state.GetOrCreateUser(pid)
			info := buildPeopleInfo(pid, pUser, uint32(time.Now().Unix()))
			buf.Write(info)
		}
		ctx.Server.SendResponse(ctx.Conn, 2003, ctx.UserID, buf.Bytes())
	}
}

func handleMapOgreList(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		mapID := user.MapID

		// mapID -> slot -> [petID, shiny]
		ogres := map[uint32]map[int][2]uint32{
			8: {
				0: {10, 0},
				1: {58, 0},
			},
			301: {
				0: {1, 0},
				1: {4, 0},
				2: {7, 0},
				3: {10, 0},
			},
		}
		buf := new(bytes.Buffer)
		for i := 0; i <= 8; i++ {
			if data, ok := ogres[mapID][i]; ok {
				binary.Write(buf, binary.BigEndian, data[0])
				binary.Write(buf, binary.BigEndian, data[1])
			} else {
				binary.Write(buf, binary.BigEndian, uint32(0))
				binary.Write(buf, binary.BigEndian, uint32(0))
			}
		}
		ctx.Server.SendResponse(ctx.Conn, 2004, ctx.UserID, buf.Bytes())
	}
}

func handleGetSimUserInfo(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := ctx.UserID
		if reader.Remaining() >= 4 {
			targetID = reader.ReadUint32BE()
		}
		user := state.GetOrCreateUser(targetID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, targetID)
		protocol.WriteFixedString(buf, pickNick(user, targetID), 16)
		binary.Write(buf, binary.BigEndian, user.Color)
		binary.Write(buf, binary.BigEndian, user.Texture)
		binary.Write(buf, binary.BigEndian, uint32(0)) // vip
		binary.Write(buf, binary.BigEndian, uint32(0)) // status
		binary.Write(buf, binary.BigEndian, user.MapType)
		binary.Write(buf, binary.BigEndian, user.MapID)
		binary.Write(buf, binary.BigEndian, uint32(0)) // canBeTeacher
		binary.Write(buf, binary.BigEndian, user.TeacherID)
		binary.Write(buf, binary.BigEndian, user.StudentID)
		binary.Write(buf, binary.BigEndian, user.GraduationCount)
		binary.Write(buf, binary.BigEndian, user.Nono.VipLevel)
		binary.Write(buf, binary.BigEndian, user.Team.ID)
		if user.Team.IsShow {
			binary.Write(buf, binary.BigEndian, uint32(1))
		} else {
			binary.Write(buf, binary.BigEndian, uint32(0))
		}
		binary.Write(buf, binary.BigEndian, uint32(len(user.Clothes)))
		for _, c := range user.Clothes {
			binary.Write(buf, binary.BigEndian, c.ID)
			binary.Write(buf, binary.BigEndian, c.Level)
		}
		ctx.Server.SendResponse(ctx.Conn, 2051, ctx.UserID, buf.Bytes())
	}
}

func handleGetMoreUserInfo(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := ctx.UserID
		if reader.Remaining() >= 4 {
			targetID = reader.ReadUint32BE()
		}
		user := state.GetOrCreateUser(targetID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, targetID)
		protocol.WriteFixedString(buf, pickNick(user, targetID), 16)
		binary.Write(buf, binary.BigEndian, user.RegTime)
		binary.Write(buf, binary.BigEndian, user.PetAllNum)
		binary.Write(buf, binary.BigEndian, pickNonZero(user.PetMaxLev, 100))
		protocol.WriteFixedString(buf, "", 200)
		binary.Write(buf, binary.BigEndian, user.GraduationCount)
		binary.Write(buf, binary.BigEndian, user.MonKingWin)
		binary.Write(buf, binary.BigEndian, uint32(0)) // messWin
		binary.Write(buf, binary.BigEndian, user.MaxStage)
		binary.Write(buf, binary.BigEndian, user.MaxArenaWins)
		binary.Write(buf, binary.BigEndian, user.CurTitle)
		ctx.Server.SendResponse(ctx.Conn, 2052, ctx.UserID, buf.Bytes())
	}
}

func handleChangeNickName(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		newNick := reader.ReadFixedString(16)
		user := state.GetOrCreateUser(ctx.UserID)
		if newNick != "" {
			user.Nick = newNick
		}
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		protocol.WriteFixedString(buf, pickNick(user, ctx.UserID), 16)
		ctx.Server.SendResponse(ctx.Conn, 2061, ctx.UserID, buf.Bytes())
	}
}

func handleChangeColor(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		newColor := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if newColor != 0 {
			user.Color = newColor
		}
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, user.Color)
		binary.Write(buf, binary.BigEndian, uint32(0)) // cost
		binary.Write(buf, binary.BigEndian, user.Coins)
		ctx.Server.SendResponse(ctx.Conn, 2063, ctx.UserID, buf.Bytes())
	}
}

func handlePeopleWalk(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		walkType := reader.ReadUint32BE()
		x := reader.ReadUint32BE()
		y := reader.ReadUint32BE()
		amfLen := reader.ReadUint32BE()
		amfData := reader.ReadBytes(int(amfLen))

		user := state.GetOrCreateUser(ctx.UserID)
		user.PosX = x
		user.PosY = y

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, walkType)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, x)
		binary.Write(buf, binary.BigEndian, y)
		binary.Write(buf, binary.BigEndian, amfLen)
		if len(amfData) > 0 {
			buf.Write(amfData)
		}
		resp := protocol.BuildResponse(2101, ctx.UserID, 0, buf.Bytes())

		if user.MapID > 0 {
			state.BroadcastToMap(user.MapID, resp)
		} else {
			ctx.Server.SendResponse(ctx.Conn, 2101, ctx.UserID, buf.Bytes())
		}
	}
}

func handleChat(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		_ = reader.ReadUint32BE() // chatType
		msgLen := reader.ReadUint32BE()
		msg := reader.ReadBytes(int(msgLen))

		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		protocol.WriteFixedString(buf, pickNick(user, ctx.UserID), 16)
		binary.Write(buf, binary.BigEndian, uint32(0)) // toID
		binary.Write(buf, binary.BigEndian, uint32(len(msg)))
		if len(msg) > 0 {
			buf.Write(msg)
		}
		resp := protocol.BuildResponse(2102, ctx.UserID, 0, buf.Bytes())

		if user.MapID > 0 {
			state.BroadcastToMap(user.MapID, resp)
		} else {
			ctx.Server.SendResponse(ctx.Conn, 2102, ctx.UserID, buf.Bytes())
		}
	}
}

func handlePeopleTransform(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		transID := reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, transID)
		ctx.Server.SendResponse(ctx.Conn, 2111, ctx.UserID, buf.Bytes())
	}
}

func handleOnOrOffFlying(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		flyMode := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		user.FlyMode = flyMode
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, flyMode)
		resp := protocol.BuildResponse(2112, ctx.UserID, 0, buf.Bytes())
		if user.MapID > 0 {
			state.BroadcastToMap(user.MapID, resp)
		} else {
			ctx.Server.SendResponse(ctx.Conn, 2112, ctx.UserID, buf.Bytes())
		}
	}
}

func handleHitStone() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0)) // bonusID
		binary.Write(buf, binary.BigEndian, uint32(0)) // petID
		binary.Write(buf, binary.BigEndian, uint32(0)) // captureTm
		binary.Write(buf, binary.BigEndian, uint32(0)) // item count
		ctx.Server.SendResponse(ctx.Conn, 2105, ctx.UserID, buf.Bytes())
	}
}

func handlePrizeOfAtresiaSpace() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0)) // bonusID
		binary.Write(buf, binary.BigEndian, uint32(0)) // petID
		binary.Write(buf, binary.BigEndian, uint32(0)) // captureTm
		binary.Write(buf, binary.BigEndian, uint32(0)) // item count
		ctx.Server.SendResponse(ctx.Conn, 2106, ctx.UserID, buf.Bytes())
	}
}

func handleTransformUser(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		tranID := reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, targetID)
		binary.Write(buf, binary.BigEndian, tranID)
		binary.Write(buf, binary.BigEndian, uint32(0))
		resp := protocol.BuildResponse(2108, ctx.UserID, 0, buf.Bytes())
		user := state.GetOrCreateUser(ctx.UserID)
		if user.MapID > 0 {
			state.BroadcastToMap(user.MapID, resp)
		} else {
			ctx.Server.SendResponse(ctx.Conn, 2108, ctx.UserID, buf.Bytes())
		}
		ctx.Server.SendResponse(ctx.Conn, 2107, ctx.UserID, []byte{})
	}
}

func handleAttackBailuen() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2109, ctx.UserID, []byte{})
	}
}

func handleGetTimePoke() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2110, ctx.UserID, buf.Bytes())
	}
}

func handleRemoveCoins(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		delta := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if delta > 0 && user.Coins >= delta {
			user.Coins -= delta
		}
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.Coins)
		ctx.Server.SendResponse(ctx.Conn, 2113, ctx.UserID, buf.Bytes())
	}
}

func handleDanceAction(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		aid := reader.ReadUint32BE()
		atype := reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, aid)
		binary.Write(buf, binary.BigEndian, atype)
		ctx.Server.SendResponse(ctx.Conn, 2103, ctx.UserID, buf.Bytes())
	}
}

func handleAimat(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		tType := reader.ReadUint32BE()
		tID := reader.ReadUint32BE()
		x := reader.ReadUint32BE()
		y := reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, tType)
		binary.Write(buf, binary.BigEndian, tID)
		binary.Write(buf, binary.BigEndian, x)
		binary.Write(buf, binary.BigEndian, y)
		ctx.Server.SendResponse(ctx.Conn, 2104, ctx.UserID, buf.Bytes())
	}
}
