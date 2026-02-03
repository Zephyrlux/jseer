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
	"jseer/internal/storage"

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
				p, err = deps.Store.CreatePlayer(context.Background(), &storage.Player{
					Account:      int64(ctx.UserID),
					Nick:         pickNick(user, ctx.UserID),
					Level:        1,
					Coins:        2000,
					Gold:         0,
					MapID:        1,
					MapType:      0,
					PosX:         300,
					PosY:         270,
					LastMapID:    1,
					Color:        0x66CCFF,
					Texture:      1,
					Energy:       100,
					TimeLimit:    86400,
					CurrentPetDV: 31,
				})
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
