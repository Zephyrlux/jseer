package game

import (
	"sync"
	"time"
)

type defaultPlayerConfig struct {
	Player defaultPlayerFields `json:"player"`
	Nono   defaultNonoFields   `json:"nono"`
}

type defaultPlayerFields struct {
	Energy          uint32 `json:"energy"`
	Coins           uint32 `json:"coins"`
	FightBadge      uint32 `json:"fightBadge"`
	AllocatableExp  uint32 `json:"allocatableExp"`
	MapID           uint32 `json:"mapId"`
	PosX            uint32 `json:"posX"`
	PosY            uint32 `json:"posY"`
	TimeToday       uint32 `json:"timeToday"`
	TimeLimit       uint32 `json:"timeLimit"`
	LoginCnt        uint32 `json:"loginCnt"`
	Inviter         uint32 `json:"inviter"`
	VipLevel        uint32 `json:"vipLevel"`
	VipValue        uint32 `json:"vipValue"`
	VipStage        uint32 `json:"vipStage"`
	VipEndTime      uint32 `json:"vipEndTime"`
	TeacherID       uint32 `json:"teacherId"`
	StudentID       uint32 `json:"studentId"`
	GraduationCount uint32 `json:"graduationCount"`
	PetMaxLev       uint32 `json:"petMaxLev"`
	PetAllNum       uint32 `json:"petAllNum"`
	MonKingWin      uint32 `json:"monKingWin"`
	MessWin         uint32 `json:"messWin"`
	CurStage        uint32 `json:"curStage"`
	MaxStage        uint32 `json:"maxStage"`
	CurFreshStage   uint32 `json:"curFreshStage"`
	MaxFreshStage   uint32 `json:"maxFreshStage"`
	MaxArenaWins    uint32 `json:"maxArenaWins"`
}

type defaultNonoFields struct {
	HasNono         uint32 `json:"hasNono"`
	SuperNono       uint32 `json:"superNono"`
	NonoState       uint32 `json:"nonoState"`
	NonoColor       uint32 `json:"nonoColor"`
	NonoNick        string `json:"nonoNick"`
	NonoFlag        uint32 `json:"nonoFlag"`
	NonoPower       uint32 `json:"nonoPower"`
	NonoMate        uint32 `json:"nonoMate"`
	NonoIq          uint32 `json:"nonoIq"`
	NonoAi          uint16 `json:"nonoAi"`
	NonoBirth       uint32 `json:"nonoBirth"`
	NonoChargeTime  uint32 `json:"nonoChargeTime"`
	NonoSuperEnergy uint32 `json:"nonoSuperEnergy"`
	NonoSuperLevel  uint32 `json:"nonoSuperLevel"`
	NonoSuperStage  uint32 `json:"nonoSuperStage"`
}

var (
	defaultPlayerOnce sync.Once
	defaultPlayerCfg  defaultPlayerConfig
)

func loadDefaultPlayerConfig() defaultPlayerConfig {
	defaultPlayerOnce.Do(func() {
		defaultPlayerCfg = defaultPlayerConfig{
			Player: defaultPlayerFields{
				Energy:       100,
				Coins:        2000,
				MapID:        1,
				PosX:         300,
				PosY:         270,
				TimeLimit:    86400,
				PetMaxLev:    100,
				PetAllNum:    0,
				MaxStage:     0,
				MaxArenaWins: 0,
			},
			Nono: defaultNonoFields{
				HasNono:         1,
				NonoColor:       0xFFFFFF,
				NonoFlag:        1,
				NonoNick:        "NoNo",
				NonoPower:       10000,
				NonoMate:        10000,
				NonoIq:          0,
				NonoAi:          0,
				NonoBirth:       0,
				NonoChargeTime:  0,
				NonoSuperEnergy: 0,
				NonoSuperLevel:  0,
				NonoSuperStage:  1,
			},
		}
		_ = readConfigJSON("default-player.json", &defaultPlayerCfg)
	})
	return defaultPlayerCfg
}

func newDefaultUser(userID uint32) *User {
	cfg := loadDefaultPlayerConfig()
	now := uint32(time.Now().Unix())
	u := &User{
		ID:              userID,
		Nick:            "Seer" + itoa(userID),
		RegTime:         now - 86400*365,
		Level:           1,
		Color:           0x66CCFF,
		Texture:         1,
		Energy:          cfg.Player.Energy,
		Coins:           cfg.Player.Coins,
		FightBadge:      cfg.Player.FightBadge,
		MapID:           cfg.Player.MapID,
		MapType:         0,
		PosX:            cfg.Player.PosX,
		PosY:            cfg.Player.PosY,
		TimeToday:       cfg.Player.TimeToday,
		TimeLimit:       cfg.Player.TimeLimit,
		LoginCnt:        cfg.Player.LoginCnt,
		TeacherID:       cfg.Player.TeacherID,
		StudentID:       cfg.Player.StudentID,
		GraduationCount: cfg.Player.GraduationCount,
		PetMaxLev:       cfg.Player.PetMaxLev,
		PetAllNum:       cfg.Player.PetAllNum,
		MonKingWin:      cfg.Player.MonKingWin,
		CurStage:        cfg.Player.CurStage,
		MaxStage:        cfg.Player.MaxStage,
		CurFreshStage:   cfg.Player.CurFreshStage,
		MaxFreshStage:   cfg.Player.MaxFreshStage,
		MaxArenaWins:    cfg.Player.MaxArenaWins,
		ExpPool:         cfg.Player.AllocatableExp,
		PetDV:           31,
		TaskStatus:      make(map[int]byte),
		TaskBufs:        make(map[int]map[int]uint32),
		Items:           make(map[int]*ItemInfo),
		Friends:         make([]FriendInfo, 0),
		Blacklist:       make([]uint32, 0),
		Achievements:    make([]uint32, 0),
		Titles:          make([]uint32, 0),
		Mailbox:         make([]Mail, 0),
		BossShield:      make(map[uint64]uint32),
		StudentIDs:      make([]uint32, 0),
		Nono: NonoInfo{
			HasNono:     cfg.Nono.HasNono > 0,
			SuperNono:   cfg.Nono.SuperNono,
			State:       cfg.Nono.NonoState,
			Color:       cfg.Nono.NonoColor,
			Nick:        cfg.Nono.NonoNick,
			Flag:        cfg.Nono.NonoFlag,
			Power:       cfg.Nono.NonoPower,
			Mate:        cfg.Nono.NonoMate,
			IQ:          cfg.Nono.NonoIq,
			AI:          cfg.Nono.NonoAi,
			Birth:       cfg.Nono.NonoBirth,
			ChargeTime:  cfg.Nono.NonoChargeTime,
			SuperEnergy: cfg.Nono.NonoSuperEnergy,
			SuperLevel:  cfg.Nono.NonoSuperLevel,
			SuperStage:  cfg.Nono.NonoSuperStage,
			VipLevel:    cfg.Player.VipLevel,
			VipValue:    cfg.Player.VipValue,
			VipStage:    cfg.Player.VipStage,
			VipEndTime:  cfg.Player.VipEndTime,
		},
	}
	if u.MapID == 0 {
		u.MapID = 1
	}
	u.LastMapID = u.MapID
	return u
}
