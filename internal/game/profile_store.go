package game

import (
	"encoding/json"
)

type nonoStore struct {
	HasNono       bool     `json:"hasNono"`
	Flag          uint32   `json:"flag"`
	State         uint32   `json:"state"`
	Color         uint32   `json:"color"`
	SuperNono     uint32   `json:"superNono"`
	VipStage      uint32   `json:"vipStage"`
	VipLevel      uint32   `json:"vipLevel"`
	VipValue      uint32   `json:"vipValue"`
	AutoCharge    uint32   `json:"autoCharge"`
	VipEndTime    uint32   `json:"vipEndTime"`
	Nick          string   `json:"nick"`
	FreshManBonus uint32   `json:"freshManBonus"`
	SuperEnergy   uint32   `json:"superEnergy"`
	SuperLevel    uint32   `json:"superLevel"`
	SuperStage    uint32   `json:"superStage"`
	Power         uint32   `json:"power"`
	Mate          uint32   `json:"mate"`
	IQ            uint32   `json:"iq"`
	AI            uint16   `json:"ai"`
	HP            uint32   `json:"hp"`
	MaxHP         uint32   `json:"maxHP"`
	Energy        uint32   `json:"energy"`
	Birth         uint32   `json:"birth"`
	ChargeTime    uint32   `json:"chargeTime"`
	Expire        uint32   `json:"expire"`
	Chip          uint32   `json:"chip"`
	Grow          uint32   `json:"grow"`
	Func          [20]byte `json:"func"`
}

func encodeNonoInfo(n NonoInfo) string {
	store := nonoStore{
		HasNono:       n.HasNono,
		Flag:          n.Flag,
		State:         n.State,
		Color:         n.Color,
		SuperNono:     n.SuperNono,
		VipStage:      n.VipStage,
		VipLevel:      n.VipLevel,
		VipValue:      n.VipValue,
		AutoCharge:    n.AutoCharge,
		VipEndTime:    n.VipEndTime,
		Nick:          n.Nick,
		FreshManBonus: n.FreshManBonus,
		SuperEnergy:   n.SuperEnergy,
		SuperLevel:    n.SuperLevel,
		SuperStage:    n.SuperStage,
		Power:         n.Power,
		Mate:          n.Mate,
		IQ:            n.IQ,
		AI:            n.AI,
		HP:            n.HP,
		MaxHP:         n.MaxHP,
		Energy:        n.Energy,
		Birth:         n.Birth,
		ChargeTime:    n.ChargeTime,
		Expire:        n.Expire,
		Chip:          n.Chip,
		Grow:          n.Grow,
		Func:          n.Func,
	}
	data, err := json.Marshal(store)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func decodeNonoInfo(raw string) *NonoInfo {
	if raw == "" {
		return nil
	}
	var store nonoStore
	if err := json.Unmarshal([]byte(raw), &store); err != nil {
		return nil
	}
	return &NonoInfo{
		HasNono:       store.HasNono,
		Flag:          store.Flag,
		State:         store.State,
		Color:         store.Color,
		SuperNono:     store.SuperNono,
		VipStage:      store.VipStage,
		VipLevel:      store.VipLevel,
		VipValue:      store.VipValue,
		AutoCharge:    store.AutoCharge,
		VipEndTime:    store.VipEndTime,
		Nick:          store.Nick,
		FreshManBonus: store.FreshManBonus,
		SuperEnergy:   store.SuperEnergy,
		SuperLevel:    store.SuperLevel,
		SuperStage:    store.SuperStage,
		Power:         store.Power,
		Mate:          store.Mate,
		IQ:            store.IQ,
		AI:            store.AI,
		HP:            store.HP,
		MaxHP:         store.MaxHP,
		Energy:        store.Energy,
		Birth:         store.Birth,
		ChargeTime:    store.ChargeTime,
		Expire:        store.Expire,
		Chip:          store.Chip,
		Grow:          store.Grow,
		Func:          store.Func,
	}
}

func encodeUint32List(list []uint32) string {
	if len(list) == 0 {
		return "[]"
	}
	data, err := json.Marshal(list)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeUint32List(raw string) []uint32 {
	if raw == "" {
		return nil
	}
	var list []uint32
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil
	}
	return list
}

type mailEntry struct {
	ID         uint32 `json:"id"`
	SenderID   uint32 `json:"senderId"`
	SenderName string `json:"senderName"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreatedAt  uint32 `json:"createdAt"`
	Read       bool   `json:"read"`
}

func encodeMailbox(list []Mail) string {
	if len(list) == 0 {
		return "[]"
	}
	entries := make([]mailEntry, 0, len(list))
	for _, m := range list {
		entries = append(entries, mailEntry{
			ID:         m.ID,
			SenderID:   m.SenderID,
			SenderName: m.SenderName,
			Title:      m.Title,
			Content:    m.Content,
			CreatedAt:  m.CreatedAt,
			Read:       m.Read,
		})
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeMailbox(raw string) []Mail {
	if raw == "" {
		return nil
	}
	var entries []mailEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil
	}
	out := make([]Mail, 0, len(entries))
	for _, e := range entries {
		out = append(out, Mail{
			ID:         e.ID,
			SenderID:   e.SenderID,
			SenderName: e.SenderName,
			Title:      e.Title,
			Content:    e.Content,
			CreatedAt:  e.CreatedAt,
			Read:       e.Read,
		})
	}
	return out
}
