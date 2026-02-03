package game

import (
	"net"
	"sync"
	"time"
)

type Cloth struct {
	ID    uint32
	Level uint32
}

type ItemInfo struct {
	Count      int
	ExpireTime uint32
}

type Pet struct {
	ID        uint32
	CatchTime uint32
	Level     uint32
	DV        uint32
	Exp       int
	HP        int
	Name      string
	Skills    []int
}

type NonoInfo struct {
	HasNono    bool
	Flag       uint32
	State      uint32
	Color      uint32
	SuperNono  uint32
	VipStage   uint32
	VipLevel   uint32
	VipValue   uint32
	AutoCharge uint32
	VipEndTime uint32
	Nick       string
	FreshManBonus uint32
	SuperEnergy   uint32
	SuperLevel    uint32
	SuperStage    uint32
	Power      uint32
	Mate       uint32
	IQ         uint32
	AI         uint16
	HP         uint32
	MaxHP      uint32
	Energy     uint32
	Birth      uint32
	ChargeTime uint32
	Expire     uint32
	Chip       uint32
	Grow       uint32
	Func       [20]byte
}

type TeamInfo struct {
	ID               uint32
	Priv             uint32
	SuperCore        uint32
	IsShow           bool
	AllContribution  uint32
	CanExContribution uint32
	CoreCount        uint32
	LogoBg           uint16
	LogoIcon         uint16
	LogoColor        uint16
	TxtColor         uint16
	LogoWord         string
}

type User struct {
	PlayerID int64
	ID        uint32
	Nick      string
	Level     int
	RegTime   uint32
	Color     uint32
	Texture   uint32
	Energy    uint32
	Coins     uint32
	FightBadge uint32
	MapID     uint32
	MapType   uint32
	PosX      uint32
	PosY      uint32
	TimeToday uint32
	TimeLimit uint32
	LoginCnt  uint32

	VipFlags uint32
	VipStage uint32

	TeacherID       uint32
	StudentID       uint32
	GraduationCount uint32
	MaxPuniLv       uint32
	PetMaxLev       uint32
	PetAllNum       uint32
	MonKingWin      uint32
	CurStage        uint32
	MaxStage        uint32
	CurFreshStage   uint32
	MaxFreshStage   uint32
	MaxArenaWins    uint32
	TwoTimes        uint32
	ThreeTimes      uint32
	AutoFight       uint32
	AutoFightTimes  uint32
	EnergyTimes     uint32
	LearnTimes      uint32
	MonBtlMedal     uint32
	RecordCnt       uint32
	ObtainTm        uint32
	SoulBeadItemID  uint32
	ExpireTm        uint32
	FuseTimes       uint32

	FlyMode       uint32
	CurrentPetID  uint32
	CatchID       uint32
	PetDV         uint32
	Clothes       []Cloth
	Pets          []Pet
	CurTitle      uint32
	TaskStatus    map[int]byte
	TaskBufs      map[int]map[int]uint32
	Nono          NonoInfo
	Team          TeamInfo

	LastMapID uint32
	NonoFollowing bool
	ExpPool   uint32
	Items     map[int]*ItemInfo
}

type State struct {
	mu       sync.RWMutex
	users    map[uint32]*User
	conns    map[uint32]net.Conn
	mapUsers map[uint32]map[uint32]struct{}
}

func NewState() *State {
	return &State{
		users:    make(map[uint32]*User),
		conns:    make(map[uint32]net.Conn),
		mapUsers: make(map[uint32]map[uint32]struct{}),
	}
}

func (s *State) RegisterConn(userID uint32, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conns[userID] = conn
}

func (s *State) GetConn(userID uint32) (net.Conn, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	conn, ok := s.conns[userID]
	return conn, ok
}

func (s *State) GetOrCreateUser(userID uint32) *User {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u, ok := s.users[userID]; ok {
		return u
	}
	now := uint32(time.Now().Unix())
	u := &User{
		ID:        userID,
		Nick:      "Seer" + itoa(userID),
		RegTime:   now - 86400*365,
		Color:     0x66CCFF,
		Texture:   1,
		Energy:    100,
		Coins:     2000,
		MapID:     1,
		PosX:      300,
		PosY:      270,
		TimeLimit: 86400,
		LoginCnt:  0,
		PetDV:     31,
		TaskStatus: make(map[int]byte),
		TaskBufs:  make(map[int]map[int]uint32),
		Items:     make(map[int]*ItemInfo),
		Nono: NonoInfo{
			HasNono: true,
			Color: 0xFFFFFF,
			Flag:  1,
			State: 0,
			Nick:  "NoNo",
			Power: 10000,
			Mate:  10000,
			AI:    0,
			IQ:    0,
			HP:    10000,
			MaxHP: 10000,
			Energy: 100,
			Birth: now,
			SuperStage: 1,
		},
	}
	s.users[userID] = u
	return u
}

func (s *State) UpdatePlayerMap(userID, mapID uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[userID]
	if !ok {
		return
	}
	oldMap := u.MapID
	if oldMap != 0 {
		if set, ok := s.mapUsers[oldMap]; ok {
			delete(set, userID)
			if len(set) == 0 {
				delete(s.mapUsers, oldMap)
			}
		}
	}
	u.MapID = mapID
	if mapID == 0 {
		return
	}
	set := s.mapUsers[mapID]
	if set == nil {
		set = make(map[uint32]struct{})
		s.mapUsers[mapID] = set
	}
	set[userID] = struct{}{}
}

func (s *State) GetPlayersInMap(mapID uint32) []uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	set := s.mapUsers[mapID]
	if len(set) == 0 {
		return nil
	}
	out := make([]uint32, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}

func (s *State) GetMapCounts() map[uint32]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[uint32]int, len(s.mapUsers))
	for id, set := range s.mapUsers {
		out[id] = len(set)
	}
	return out
}

func (s *State) OnlineCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.conns)
}

func (s *State) BroadcastToMap(mapID uint32, payload []byte) {
	ids := s.GetPlayersInMap(mapID)
	for _, id := range ids {
		if conn, ok := s.GetConn(id); ok {
			_, _ = conn.Write(payload)
		}
	}
}

func itoa(v uint32) string {
	if v == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
