package game

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"jseer/internal/gateway"
	"jseer/internal/protocol"

	"go.uber.org/zap"
)

func registerSystemHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(1001, handleLoginIn(deps, state))
	s.Register(1002, handleSystemTime())
	s.Register(1004, handleMapHot(state))
	s.Register(1005, handleGetImageAddress(deps))
	s.Register(1006, handleGetSessionKey())
	s.Register(1007, handleReadCount())
	s.Register(1101, handleMoneyCheckPassword())
	s.Register(1102, handleMoneyBuyProduct(state))
	s.Register(1103, handleMoneyCheckRemain(state))
	s.Register(1104, handleGoldBuyProduct(state))
	s.Register(1105, handleGoldCheckRemain(state))
	s.Register(1106, handleGoldOnlineCheckRemain(state))
	s.Register(1108, handleNewYearRedpackets())
	s.Register(1110, handleGetYuanxiaoGift())
	s.Register(1111, handleNameplateExchangePet())
	s.Register(1112, handleGetNameplate())
}

func handleLoginIn(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		if state != nil {
			state.RegisterConn(ctx.UserID, ctx.Conn)
		}
		user := state.GetOrCreateUser(ctx.UserID)
		if deps != nil && deps.Store != nil {
			p, err := deps.Store.GetPlayerByAccount(context.Background(), int64(ctx.UserID))
			if err != nil {
				player := buildPlayerUpdate(user, int64(ctx.UserID))
				if player != nil {
					player.Nick = pickNick(user, ctx.UserID)
					if player.LastMapID == 0 {
						player.LastMapID = player.MapID
					}
					p, err = deps.Store.CreatePlayer(context.Background(), player)
				}
				if err == nil {
					syncUserFromPlayer(ctx.UserID, user, p)
				}
			} else {
				syncUserFromPlayer(ctx.UserID, user, p)
			}

			if user.PlayerID > 0 {
				if pets, err := deps.Store.ListPetsByPlayer(context.Background(), user.PlayerID); err == nil {
					user.Pets = user.Pets[:0]
					for _, p := range pets {
						user.Pets = append(user.Pets, Pet{
							ID:        uint32(p.SpeciesID),
							CatchTime: uint32(p.CatchTime),
							Level:     uint32(p.Level),
							DV:        uint32(p.DV),
							Exp:       p.Exp,
							HP:        p.HP,
							Skills:    decodePetSkills(p.Skills),
						})
					}
				}
				if items, err := deps.Store.ListItemsByPlayer(context.Background(), user.PlayerID); err == nil {
					if user.Items == nil {
						user.Items = make(map[int]*ItemInfo)
					} else {
						for k := range user.Items {
							delete(user.Items, k)
						}
					}
					for _, it := range items {
						user.Items[it.ItemID] = &ItemInfo{
							Count:      it.Count,
							ExpireTime: decodeItemMeta(it.Meta),
						}
					}
				}
			}
		}
		applySpawnOverride(deps, user, user.LoginCnt == 0)
		if user.LoginCnt == 0 {
			user.LoginCnt = 1
		} else {
			user.LoginCnt++
		}
		if user.MapID > 0 {
			state.UpdatePlayerMap(ctx.UserID, user.MapID)
		}
		body := buildLoginResponse(user)
		ctx.Server.SendResponse(ctx.Conn, 1001, ctx.UserID, body)
		pushInitialMapEnter(deps, state, ctx)
		if user.Nono.SuperNono > 0 {
			vipBuf := new(bytes.Buffer)
			binary.Write(vipBuf, binary.BigEndian, ctx.UserID)
			binary.Write(vipBuf, binary.BigEndian, uint32(2))
			binary.Write(vipBuf, binary.BigEndian, user.Nono.AutoCharge)
			endTime := user.Nono.VipEndTime
			if endTime == 0 {
				endTime = 0x7FFFFFFF
			}
			binary.Write(vipBuf, binary.BigEndian, endTime)
			ctx.Server.SendResponse(ctx.Conn, 8006, ctx.UserID, vipBuf.Bytes())
		}
		if deps != nil && deps.Logger != nil {
			deps.Logger.Info("LOGIN_IN response", zap.Uint32("uid", ctx.UserID))
		}
	}
}

func applySpawnOverride(deps *Deps, user *User, isFirstLogin bool) {
	if deps == nil || user == nil {
		return
	}
	if deps.SpawnMap == 0 {
		deps.SpawnMap = 1
	}
	force := deps.ForceSpawn || isFirstLogin
	if force || user.MapID == 0 {
		user.MapID = deps.SpawnMap
		if deps.SpawnX > 0 {
			user.PosX = deps.SpawnX
		}
		if deps.SpawnY > 0 {
			user.PosY = deps.SpawnY
		}
		if user.PosX == 0 {
			user.PosX = 300
		}
		if user.PosY == 0 {
			user.PosY = 270
		}
	}
}

func pushInitialMapEnter(deps *Deps, state *State, ctx *gateway.Context) {
	if state == nil {
		return
	}
	user := state.GetOrCreateUser(ctx.UserID)
	cfg := loadDefaultPlayerConfig()
	mapID := user.MapID
	if mapID == 0 {
		mapID = cfg.Player.MapID
	}
	if deps != nil && deps.SpawnMap != 0 {
		if deps.ForceSpawn || user.LoginCnt <= 1 {
			mapID = deps.SpawnMap
		}
	}
	if mapID == 0 {
		mapID = 1
	}
	x := user.PosX
	y := user.PosY
	if deps != nil && (deps.ForceSpawn || user.LoginCnt <= 1) {
		if deps.SpawnX > 0 {
			x = deps.SpawnX
		}
		if deps.SpawnY > 0 {
			y = deps.SpawnY
		}
	}
	if x == 0 && y == 0 {
		x = cfg.Player.PosX
		y = cfg.Player.PosY
	}

	user.MapType = 0
	user.PosX = x
	user.PosY = y
	user.LastMapID = mapID
	state.UpdatePlayerMap(ctx.UserID, mapID)
	savePlayer(deps, ctx.UserID, user)

	body := buildPeopleInfo(ctx.UserID, user, uint32(time.Now().Unix()))
	ctx.Server.SendResponse(ctx.Conn, 2001, ctx.UserID, body)

	listBuf := new(bytes.Buffer)
	players := state.GetPlayersInMap(mapID)
	binary.Write(listBuf, binary.BigEndian, uint32(len(players)))
	for _, pid := range players {
		pUser := state.GetOrCreateUser(pid)
		info := buildPeopleInfo(pid, pUser, uint32(time.Now().Unix()))
		listBuf.Write(info)
	}
	ctx.Server.SendResponse(ctx.Conn, 2003, ctx.UserID, listBuf.Bytes())

	ctx.Server.SendResponse(ctx.Conn, 2004, ctx.UserID, buildMapOgreBody(mapID))

	if bossBody := buildMapBossListForUser(user, int(mapID)); len(bossBody) > 0 {
		ctx.Server.SendResponse(ctx.Conn, 2021, ctx.UserID, bossBody)
	}

	if mapID > 10000 || mapID == ctx.UserID {
		handleNonoInfo(state)(ctx)
	}

	if deps != nil && deps.Logger != nil {
		deps.Logger.Info(
			"push initial map enter",
			zap.Uint32("uid", ctx.UserID),
			zap.Uint32("map_id", mapID),
			zap.Uint32("pos_x", x),
			zap.Uint32("pos_y", y),
		)
	}
}

func handleSystemTime() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint32(buf[0:4], uint32(time.Now().Unix()))
		binary.BigEndian.PutUint32(buf[4:8], 0)
		ctx.Server.SendResponse(ctx.Conn, 1002, ctx.UserID, buf)
	}
}

func handleMapHot(state *State) gateway.Handler {
	officialMaps := []uint32{
		1, 4, 5, 325, 6, 7, 8, 328, 9, 10,
		333, 15, 17, 338, 19, 20, 25, 30,
		101, 102, 103, 40, 107, 47, 51, 54, 57, 314, 60,
	}
	return func(ctx *gateway.Context) {
		mapCounts := map[uint32]int{}
		if state != nil {
			mapCounts = state.GetMapCounts()
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(len(officialMaps)))
		for _, mapID := range officialMaps {
			binary.Write(buf, binary.BigEndian, mapID)
			binary.Write(buf, binary.BigEndian, uint32(mapCounts[mapID]))
		}
		ctx.Server.SendResponse(ctx.Conn, 1004, ctx.UserID, buf.Bytes())
	}
}

func handleGetImageAddress(deps *Deps) gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		ip := deps.GameIP
		if ip == "" {
			ip = "127.0.0.1"
		}
		protocol.WriteFixedString(buf, ip, 16)
		binary.Write(buf, binary.BigEndian, uint16(80))
		protocol.WriteFixedString(buf, "", 16)
		ctx.Server.SendResponse(ctx.Conn, 1005, ctx.UserID, buf.Bytes())
	}
}

func handleGetSessionKey() gateway.Handler {
	return func(ctx *gateway.Context) {
		now := uint32(time.Now().Unix())
		sum := md5.Sum([]byte(fmt.Sprintf("%d:%d", ctx.UserID, now)))
		key := hex.EncodeToString(sum[:])
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, now)
		buf.WriteString(key)
		ctx.Server.SendResponse(ctx.Conn, 1006, ctx.UserID, buf.Bytes())
	}
}

func handleReadCount() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 1007, ctx.UserID, buf.Bytes())
	}
}

func handleMoneyCheckPassword() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(1))
		ctx.Server.SendResponse(ctx.Conn, 1101, ctx.UserID, buf.Bytes())
	}
}

func handleMoneyBuyProduct(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, user.Coins*100)
		ctx.Server.SendResponse(ctx.Conn, 1102, ctx.UserID, buf.Bytes())
	}
}

func handleMoneyCheckRemain(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.Coins*100)
		ctx.Server.SendResponse(ctx.Conn, 1103, ctx.UserID, buf.Bytes())
	}
}

func handleGoldBuyProduct(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, user.Gold*100)
		ctx.Server.SendResponse(ctx.Conn, 1104, ctx.UserID, buf.Bytes())
	}
}

func handleGoldCheckRemain(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.Gold*100)
		ctx.Server.SendResponse(ctx.Conn, 1105, ctx.UserID, buf.Bytes())
	}
}

func handleGoldOnlineCheckRemain(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.Gold)
		ctx.Server.SendResponse(ctx.Conn, 1106, ctx.UserID, buf.Bytes())
	}
}

func handleNewYearRedpackets() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 1108, ctx.UserID, buf.Bytes())
	}
}

func handleGetYuanxiaoGift() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 1110, ctx.UserID, buf.Bytes())
	}
}

func handleNameplateExchangePet() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 1111, ctx.UserID, []byte{})
	}
}

func handleGetNameplate() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 1112, ctx.UserID, []byte{})
	}
}
