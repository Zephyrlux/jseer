package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerItemHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2601, handleItemBuy(deps, state))
	s.Register(2602, handleItemSale(deps, state))
	s.Register(2603, handleItemRepair())
	s.Register(2604, handleChangeCloth(deps, state))
	s.Register(2605, handleItemList(state))
	s.Register(2606, handleMultiItemBuy(deps, state))
	s.Register(2607, handleItemExpend(deps, state))
	s.Register(2608, handleGetLastEgg())
	s.Register(2609, handleEquipUpdate())
	s.Register(2610, handleEatSpecialMedicine())
	s.Register(2901, handleExchangeClothComplete())
}

func handleItemBuy(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		itemID := int(reader.ReadUint32BE())
		count := int(reader.ReadUint32BE())
		if count <= 0 {
			count = 1
		}
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Items == nil {
			user.Items = make(map[int]*ItemInfo)
		}
		if isUniqueItem(itemID) && user.Items[itemID] != nil {
			resp := protocol.BuildResponse(2601, ctx.UserID, 103203, []byte{})
			_, _ = ctx.Conn.Write(resp)
			return
		}
		unitPrice := getItemPrice(itemID)
		totalCost := unitPrice * count
		if totalCost > 0 {
			if int(user.Coins) < totalCost {
				return
			}
			user.Coins -= uint32(totalCost)
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
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2601, ctx.UserID, buf.Bytes())
	}
}

func handleItemSale(deps *Deps, state *State) gateway.Handler {
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
				info.Count -= count
				if info.Count <= 0 {
					delete(user.Items, itemID)
				}
				upsertItem(deps, user, itemID)
			}
		}
		ctx.Server.SendResponse(ctx.Conn, 2602, ctx.UserID, []byte{})
	}
}

func handleChangeCloth(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		count := int(reader.ReadUint32BE())
		clothes := make([]Cloth, 0, count)
		for i := 0; i < count; i++ {
			clothID := reader.ReadUint32BE()
			clothes = append(clothes, Cloth{ID: clothID, Level: 0})
		}
		user := state.GetOrCreateUser(ctx.UserID)
		user.Clothes = clothes
		savePlayer(deps, ctx.UserID, user)

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		binary.Write(buf, binary.BigEndian, uint32(len(clothes)))
		for _, c := range clothes {
			binary.Write(buf, binary.BigEndian, c.ID)
			binary.Write(buf, binary.BigEndian, uint32(0))
		}
		resp := protocol.BuildResponse(2604, ctx.UserID, 0, buf.Bytes())
		if user.MapID > 0 {
			state.BroadcastToMap(user.MapID, resp)
		} else {
			ctx.Server.SendResponse(ctx.Conn, 2604, ctx.UserID, buf.Bytes())
		}
	}
}

func handleItemExpend(deps *Deps, state *State) gateway.Handler {
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
				info.Count -= count
				if info.Count <= 0 {
					delete(user.Items, itemID)
				}
				upsertItem(deps, user, itemID)
			}
		}
		ctx.Server.SendResponse(ctx.Conn, 2607, ctx.UserID, []byte{})
	}
}

func handleItemList(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		itemType1 := int(reader.ReadUint32BE())
		itemType2 := int(reader.ReadUint32BE())
		itemType3 := int(reader.ReadUint32BE())
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Items == nil {
			user.Items = make(map[int]*ItemInfo)
		}

		buf := new(bytes.Buffer)
		count := 0
		itemsBuf := new(bytes.Buffer)
		for itemID, info := range user.Items {
			if itemID < itemType1 || itemID > itemType2 {
				if itemID != itemType3 {
					continue
				}
			}
			binary.Write(itemsBuf, binary.BigEndian, uint32(itemID))
			binary.Write(itemsBuf, binary.BigEndian, uint32(info.Count))
			expire := info.ExpireTime
			if expire == 0 {
				expire = defaultItemExpire
			}
			binary.Write(itemsBuf, binary.BigEndian, expire)
			binary.Write(itemsBuf, binary.BigEndian, uint32(0))
			count++
		}
		binary.Write(buf, binary.BigEndian, uint32(count))
		buf.Write(itemsBuf.Bytes())
		ctx.Server.SendResponse(ctx.Conn, 2605, ctx.UserID, buf.Bytes())
	}
}

func handleMultiItemBuy(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		itemCount := int(reader.ReadUint32BE())
		itemIDs := make([]int, 0, itemCount)
		for i := 0; i < itemCount; i++ {
			itemIDs = append(itemIDs, int(reader.ReadUint32BE()))
		}
		user := state.GetOrCreateUser(ctx.UserID)
		if user.Items == nil {
			user.Items = make(map[int]*ItemInfo)
		}

		totalCost := 0
		validItems := make([]int, 0, len(itemIDs))
		for _, itemID := range itemIDs {
			if isUniqueItem(itemID) && user.Items[itemID] != nil {
				continue
			}
			totalCost += getItemPrice(itemID)
			validItems = append(validItems, itemID)
		}
		if int(user.Coins) < totalCost {
			buf := new(bytes.Buffer)
			binary.Write(buf, binary.BigEndian, uint32(10016))
			binary.Write(buf, binary.BigEndian, user.Coins)
			ctx.Server.SendResponse(ctx.Conn, 2606, ctx.UserID, buf.Bytes())
			return
		}
		if totalCost > 0 {
			user.Coins -= uint32(totalCost)
		}
		for _, itemID := range validItems {
			info := user.Items[itemID]
			if info == nil {
				info = &ItemInfo{Count: 0, ExpireTime: defaultItemExpire}
				user.Items[itemID] = info
			}
			info.Count++
			upsertItem(deps, user, itemID)
		}
		savePlayer(deps, ctx.UserID, user)

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, user.Coins)
		ctx.Server.SendResponse(ctx.Conn, 2606, ctx.UserID, buf.Bytes())
	}
}

func handleEquipUpdate() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2609, ctx.UserID, []byte{})
	}
}

func handleItemRepair() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2603, ctx.UserID, buf.Bytes())
	}
}

func handleGetLastEgg() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2608, ctx.UserID, buf.Bytes())
	}
}

func handleEatSpecialMedicine() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2610, ctx.UserID, buf.Bytes())
	}
}

func handleExchangeClothComplete() gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		_ = reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(1))
		ctx.Server.SendResponse(ctx.Conn, 2901, ctx.UserID, buf.Bytes())
	}
}
