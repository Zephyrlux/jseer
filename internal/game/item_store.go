package game

import (
	"context"
	"encoding/json"
)

const defaultItemExpire uint32 = 0x057E40

type itemMeta struct {
	ExpireTime uint32 `json:"expireTime"`
}

func decodeItemMeta(meta string) uint32 {
	if meta == "" {
		return defaultItemExpire
	}
	var m itemMeta
	if err := json.Unmarshal([]byte(meta), &m); err != nil {
		return defaultItemExpire
	}
	if m.ExpireTime == 0 {
		return defaultItemExpire
	}
	return m.ExpireTime
}

func encodeItemMeta(expire uint32) string {
	if expire == 0 {
		expire = defaultItemExpire
	}
	data, err := json.Marshal(itemMeta{ExpireTime: expire})
	if err != nil {
		return ""
	}
	return string(data)
}

func upsertItem(deps *Deps, user *User, itemID int) {
	if deps == nil || deps.Store == nil || user == nil || user.PlayerID == 0 {
		return
	}
	info := user.Items[itemID]
	if info == nil || info.Count <= 0 {
		_ = deps.Store.DeleteItem(context.Background(), user.PlayerID, itemID)
		return
	}
	_, _ = deps.Store.UpsertItem(context.Background(), user.PlayerID, itemID, info.Count, encodeItemMeta(info.ExpireTime))
}
