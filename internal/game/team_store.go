package game

import (
	"encoding/json"
)

type teamInfoPayload struct {
	ID                uint32 `json:"id"`
	Priv              uint32 `json:"priv"`
	SuperCore         uint32 `json:"superCore"`
	IsShow            bool   `json:"isShow"`
	AllContribution   uint32 `json:"allContribution"`
	CanExContribution uint32 `json:"canExContribution"`
	CoreCount         uint32 `json:"coreCount"`
	Name              string `json:"name"`
	MemberCount       uint32 `json:"memberCount"`
	Interest          uint32 `json:"interest"`
	JoinFlag          uint32 `json:"joinFlag"`
	Exp               uint32 `json:"exp"`
	Score             uint32 `json:"score"`
	Slogan            string `json:"slogan"`
	Notice            string `json:"notice"`
	LogoBg            uint16 `json:"logoBg"`
	LogoIcon          uint16 `json:"logoIcon"`
	LogoColor         uint16 `json:"logoColor"`
	TxtColor          uint16 `json:"txtColor"`
	LogoWord          string `json:"logoWord"`
}

func decodeTeamInfo(raw string) TeamInfo {
	if raw == "" {
		return TeamInfo{}
	}
	var p teamInfoPayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return TeamInfo{}
	}
	return TeamInfo{
		ID:                p.ID,
		Priv:              p.Priv,
		SuperCore:         p.SuperCore,
		IsShow:            p.IsShow,
		AllContribution:   p.AllContribution,
		CanExContribution: p.CanExContribution,
		CoreCount:         p.CoreCount,
		Name:              p.Name,
		MemberCount:       p.MemberCount,
		Interest:          p.Interest,
		JoinFlag:          p.JoinFlag,
		Exp:               p.Exp,
		Score:             p.Score,
		Slogan:            p.Slogan,
		Notice:            p.Notice,
		LogoBg:            p.LogoBg,
		LogoIcon:          p.LogoIcon,
		LogoColor:         p.LogoColor,
		TxtColor:          p.TxtColor,
		LogoWord:          p.LogoWord,
	}
}

func encodeTeamInfo(info TeamInfo) string {
	p := teamInfoPayload{
		ID:                info.ID,
		Priv:              info.Priv,
		SuperCore:         info.SuperCore,
		IsShow:            info.IsShow,
		AllContribution:   info.AllContribution,
		CanExContribution: info.CanExContribution,
		CoreCount:         info.CoreCount,
		Name:              info.Name,
		MemberCount:       info.MemberCount,
		Interest:          info.Interest,
		JoinFlag:          info.JoinFlag,
		Exp:               info.Exp,
		Score:             info.Score,
		Slogan:            info.Slogan,
		Notice:            info.Notice,
		LogoBg:            info.LogoBg,
		LogoIcon:          info.LogoIcon,
		LogoColor:         info.LogoColor,
		TxtColor:          info.TxtColor,
		LogoWord:          info.LogoWord,
	}
	data, err := json.Marshal(p)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func decodeStudentIDs(raw string) []uint32 {
	return decodeBlacklist(raw)
}

func encodeStudentIDs(list []uint32) string {
	return encodeBlacklist(list)
}
