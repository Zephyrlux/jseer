package game

import (
	"bytes"
	"encoding/binary"
	"time"

	"jseer/internal/protocol"
)

func buildPeopleInfo(userID uint32, u *User, now uint32) []byte {
	if u == nil {
		return nil
	}
	if now == 0 {
		now = uint32(time.Now().Unix())
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, now)
	binary.Write(buf, binary.BigEndian, userID)
	protocol.WriteFixedString(buf, pickNick(u, userID), 16)
	binary.Write(buf, binary.BigEndian, u.Color)
	binary.Write(buf, binary.BigEndian, u.Texture)

	vipFlags := u.VipFlags
	if u.Nono.SuperNono > 0 {
		vipFlags = 3
	}
	binary.Write(buf, binary.BigEndian, vipFlags)
	binary.Write(buf, binary.BigEndian, u.Nono.VipStage)

	actionType := uint32(0)
	if u.FlyMode > 0 {
		actionType = 1
	}
	binary.Write(buf, binary.BigEndian, actionType)
	binary.Write(buf, binary.BigEndian, u.PosX)
	binary.Write(buf, binary.BigEndian, u.PosY)
	binary.Write(buf, binary.BigEndian, uint32(0)) // action
	binary.Write(buf, binary.BigEndian, uint32(0)) // direction
	binary.Write(buf, binary.BigEndian, uint32(0)) // changeShape

	binary.Write(buf, binary.BigEndian, u.CatchID)
	binary.Write(buf, binary.BigEndian, u.CurrentPetID)
	petDV := u.PetDV
	if petDV == 0 {
		petDV = 31
	}
	binary.Write(buf, binary.BigEndian, petDV)
	binary.Write(buf, binary.BigEndian, uint32(0)) // petSkin
	binary.Write(buf, binary.BigEndian, uint32(0)) // fightFlag

	binary.Write(buf, binary.BigEndian, u.TeacherID)
	binary.Write(buf, binary.BigEndian, u.StudentID)

	nonoState := u.Nono.Flag
	binary.Write(buf, binary.BigEndian, nonoState)
	binary.Write(buf, binary.BigEndian, u.Nono.Color)
	super := uint32(0)
	if u.Nono.SuperNono > 0 {
		super = 1
	}
	binary.Write(buf, binary.BigEndian, super)
	binary.Write(buf, binary.BigEndian, uint32(0)) // playerForm
	binary.Write(buf, binary.BigEndian, uint32(0)) // transTime

	team := u.Team
	binary.Write(buf, binary.BigEndian, team.ID)
	binary.Write(buf, binary.BigEndian, team.CoreCount)
	if team.IsShow {
		binary.Write(buf, binary.BigEndian, uint32(1))
	} else {
		binary.Write(buf, binary.BigEndian, uint32(0))
	}
	protocol.WriteUint16BE(buf, team.LogoBg)
	protocol.WriteUint16BE(buf, team.LogoIcon)
	protocol.WriteUint16BE(buf, team.LogoColor)
	protocol.WriteUint16BE(buf, team.TxtColor)
	protocol.WriteFixedString(buf, team.LogoWord, 4)

	binary.Write(buf, binary.BigEndian, uint32(len(u.Clothes)))
	for _, c := range u.Clothes {
		binary.Write(buf, binary.BigEndian, c.ID)
		binary.Write(buf, binary.BigEndian, c.Level)
	}

	binary.Write(buf, binary.BigEndian, u.CurTitle)
	return buf.Bytes()
}

func buildLoginResponse(u *User) []byte {
	if u == nil {
		return nil
	}
	if u.RegTime == 0 {
		u.RegTime = uint32(time.Now().Unix()) - 86400*365
	}
	buf := new(bytes.Buffer)

	// 1. Account basics
	binary.Write(buf, binary.BigEndian, u.ID)
	binary.Write(buf, binary.BigEndian, u.RegTime)
	protocol.WriteFixedString(buf, pickNick(u, u.ID), 16)

	vipFlags := u.VipFlags
	if u.Nono.SuperNono > 0 {
		vipFlags = 3
	}
	binary.Write(buf, binary.BigEndian, vipFlags)

	// 3. Basic attributes
	binary.Write(buf, binary.BigEndian, uint32(0)) // dsFlag
	binary.Write(buf, binary.BigEndian, u.Color)
	binary.Write(buf, binary.BigEndian, u.Texture)
	binary.Write(buf, binary.BigEndian, u.Energy)
	binary.Write(buf, binary.BigEndian, u.Coins)
	binary.Write(buf, binary.BigEndian, u.FightBadge)
	binary.Write(buf, binary.BigEndian, u.MapID)
	binary.Write(buf, binary.BigEndian, u.PosX)
	binary.Write(buf, binary.BigEndian, u.PosY)
	binary.Write(buf, binary.BigEndian, u.TimeToday)
	if u.TimeLimit == 0 {
		u.TimeLimit = 86400
	}
	binary.Write(buf, binary.BigEndian, u.TimeLimit)

	// 4. Flags (4 bytes)
	buf.WriteByte(0)
	buf.WriteByte(0)
	buf.WriteByte(0)
	buf.WriteByte(0)

	// 5. Stats
	binary.Write(buf, binary.BigEndian, u.LoginCnt)
	binary.Write(buf, binary.BigEndian, uint32(0)) // inviter
	binary.Write(buf, binary.BigEndian, uint32(0)) // newInviteeCnt
	binary.Write(buf, binary.BigEndian, u.Nono.VipLevel)
	binary.Write(buf, binary.BigEndian, u.Nono.VipValue)
	binary.Write(buf, binary.BigEndian, u.Nono.VipStage)
	binary.Write(buf, binary.BigEndian, u.Nono.AutoCharge)
	endTime := u.Nono.VipEndTime
	if u.Nono.SuperNono > 0 && endTime == 0 {
		endTime = 0x7FFFFFFF
	}
	binary.Write(buf, binary.BigEndian, endTime)
	binary.Write(buf, binary.BigEndian, uint32(0)) // freshManBonus

	// 6. Lists
	protocol.WriteFixedString(buf, "", 80)
	protocol.WriteFixedString(buf, "", 50)

	// 7. More stats
	binary.Write(buf, binary.BigEndian, u.TeacherID)
	binary.Write(buf, binary.BigEndian, u.StudentID)
	binary.Write(buf, binary.BigEndian, u.GraduationCount)
	binary.Write(buf, binary.BigEndian, pickNonZero(u.MaxPuniLv, 100))
	binary.Write(buf, binary.BigEndian, pickNonZero(u.PetMaxLev, 100))
	binary.Write(buf, binary.BigEndian, u.PetAllNum)
	binary.Write(buf, binary.BigEndian, u.MonKingWin)
	binary.Write(buf, binary.BigEndian, u.CurStage)
	binary.Write(buf, binary.BigEndian, u.MaxStage)
	binary.Write(buf, binary.BigEndian, u.CurFreshStage)
	binary.Write(buf, binary.BigEndian, u.MaxFreshStage)
	binary.Write(buf, binary.BigEndian, u.MaxArenaWins)
	binary.Write(buf, binary.BigEndian, u.TwoTimes)
	binary.Write(buf, binary.BigEndian, u.ThreeTimes)
	binary.Write(buf, binary.BigEndian, u.AutoFight)
	binary.Write(buf, binary.BigEndian, u.AutoFightTimes)
	binary.Write(buf, binary.BigEndian, u.EnergyTimes)
	binary.Write(buf, binary.BigEndian, u.LearnTimes)
	binary.Write(buf, binary.BigEndian, u.MonBtlMedal)
	binary.Write(buf, binary.BigEndian, u.RecordCnt)
	binary.Write(buf, binary.BigEndian, u.ObtainTm)
	binary.Write(buf, binary.BigEndian, u.SoulBeadItemID)
	binary.Write(buf, binary.BigEndian, u.ExpireTm)
	binary.Write(buf, binary.BigEndian, u.FuseTimes)

	// 8. NoNo
	hasNono := uint32(0)
	if u.Nono.HasNono {
		hasNono = 1
	}
	binary.Write(buf, binary.BigEndian, hasNono)
	super := uint32(0)
	if u.Nono.SuperNono > 0 {
		super = 1
	}
	binary.Write(buf, binary.BigEndian, super)
	binary.Write(buf, binary.BigEndian, u.Nono.Flag)
	binary.Write(buf, binary.BigEndian, u.Nono.Color)
	protocol.WriteFixedString(buf, u.Nono.Nick, 16)

	// 9. TeamInfo
	team := u.Team
	binary.Write(buf, binary.BigEndian, team.ID)
	binary.Write(buf, binary.BigEndian, team.Priv)
	binary.Write(buf, binary.BigEndian, team.SuperCore)
	if team.IsShow {
		binary.Write(buf, binary.BigEndian, uint32(1))
	} else {
		binary.Write(buf, binary.BigEndian, uint32(0))
	}
	binary.Write(buf, binary.BigEndian, team.AllContribution)
	binary.Write(buf, binary.BigEndian, team.CanExContribution)

	// 10. TeamPKInfo
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(0))

	// 11. Badge & reserved
	buf.WriteByte(0)
	binary.Write(buf, binary.BigEndian, uint32(0))
	protocol.WriteFixedString(buf, "", 27)

	// 12. Task list (500 bytes)
	for i := 1; i <= 500; i++ {
		if u.TaskStatus != nil {
			buf.WriteByte(u.TaskStatus[i])
		} else {
			buf.WriteByte(0)
		}
	}

	// 13. Pet list
	binary.Write(buf, binary.BigEndian, uint32(len(u.Pets)))
	for _, p := range u.Pets {
		petBody := buildFullPetInfo(int(p.ID), int(p.CatchTime), int(p.Level), int(p.DV), p.Exp, p.Skills)
		buf.Write(petBody)
	}

	// 14. Clothes
	binary.Write(buf, binary.BigEndian, uint32(len(u.Clothes)))
	for _, c := range u.Clothes {
		binary.Write(buf, binary.BigEndian, c.ID)
		binary.Write(buf, binary.BigEndian, c.Level)
	}

	// 15. Title & achievements
	binary.Write(buf, binary.BigEndian, u.CurTitle)
	protocol.WriteFixedString(buf, "", 200)

	return buf.Bytes()
}

func pickNick(u *User, userID uint32) string {
	if u.Nick != "" {
		return u.Nick
	}
	return "Seer" + itoa(userID)
}

func pickNonZero(v uint32, fallback uint32) uint32 {
	if v == 0 {
		return fallback
	}
	return v
}
