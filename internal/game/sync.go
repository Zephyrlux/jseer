package game

import (
	"context"

	"jseer/internal/storage"
)

func syncUserFromPlayer(userID uint32, u *User, p *storage.Player) *User {
	if u == nil || p == nil {
		return u
	}
	u.PlayerID = p.ID
	u.Nick = p.Nick
	u.Level = int(p.Level)
	u.Coins = uint32(p.Coins)
	u.Gold = uint32(p.Gold)
	u.MapID = uint32(p.MapID)
	u.MapType = uint32(p.MapType)
	u.PosX = uint32(p.PosX)
	u.PosY = uint32(p.PosY)
	u.LastMapID = uint32(p.LastMapID)
	u.Color = uint32(p.Color)
	u.Texture = uint32(p.Texture)
	u.Energy = uint32(p.Energy)
	u.FightBadge = uint32(p.FightBadge)
	u.TimeToday = uint32(p.TimeToday)
	u.TimeLimit = uint32(p.TimeLimit)
	u.TeacherID = uint32(p.TeacherID)
	u.StudentID = uint32(p.StudentID)
	u.CurTitle = uint32(p.CurTitle)
	if p.TaskStatus != "" {
		u.TaskStatus = decodeTaskStatus(p.TaskStatus)
	} else if u.TaskStatus == nil {
		u.TaskStatus = make(map[int]byte)
	}
	if p.TaskBufs != "" {
		u.TaskBufs = decodeTaskBufs(p.TaskBufs)
	} else if u.TaskBufs == nil {
		u.TaskBufs = make(map[int]map[int]uint32)
	}
	if p.Friends != "" {
		u.Friends = decodeFriends(p.Friends)
	} else if u.Friends == nil {
		u.Friends = make([]FriendInfo, 0)
	}
	if p.Blacklist != "" {
		u.Blacklist = decodeBlacklist(p.Blacklist)
	} else if u.Blacklist == nil {
		u.Blacklist = make([]uint32, 0)
	}
	if p.TeamInfo != "" {
		u.Team = decodeTeamInfo(p.TeamInfo)
	}
	if p.StudentIDs != "" {
		u.StudentIDs = decodeStudentIDs(p.StudentIDs)
	} else if u.StudentIDs == nil {
		u.StudentIDs = make([]uint32, 0)
	}
	if p.RoomID > 0 {
		u.RoomID = uint32(p.RoomID)
	}
	if p.Fitments != "" {
		u.Fitments = decodeFitments(p.Fitments)
	}
	u.CurrentPetID = uint32(p.CurrentPetID)
	u.CatchID = uint32(p.CurrentPetCatchTime)
	u.PetDV = uint32(p.CurrentPetDV)
	return u
}

func buildPlayerUpdate(u *User, accountID int64) *storage.Player {
	if u == nil {
		return nil
	}
	return &storage.Player{
		ID:                  u.PlayerID,
		Account:             accountID,
		Nick:                u.Nick,
		Level:               u.Level,
		Coins:               int64(u.Coins),
		Gold:                int64(u.Gold),
		MapID:               int(u.MapID),
		MapType:             int(u.MapType),
		PosX:                int(u.PosX),
		PosY:                int(u.PosY),
		LastMapID:           int(u.LastMapID),
		Color:               int64(u.Color),
		Texture:             int64(u.Texture),
		Energy:              int64(u.Energy),
		FightBadge:          int64(u.FightBadge),
		TimeToday:           int64(u.TimeToday),
		TimeLimit:           int64(u.TimeLimit),
		TeacherID:           int64(u.TeacherID),
		StudentID:           int64(u.StudentID),
		CurTitle:            int64(u.CurTitle),
		TaskStatus:          encodeTaskStatus(u.TaskStatus),
		TaskBufs:            encodeTaskBufs(u.TaskBufs),
		Friends:             encodeFriends(u.Friends),
		Blacklist:           encodeBlacklist(u.Blacklist),
		TeamInfo:            encodeTeamInfo(u.Team),
		StudentIDs:          encodeStudentIDs(u.StudentIDs),
		RoomID:              int64(u.RoomID),
		Fitments:            encodeFitments(u.Fitments),
		CurrentPetID:        int64(u.CurrentPetID),
		CurrentPetCatchTime: int64(u.CatchID),
		CurrentPetDV:        int64(u.PetDV),
	}
}

func savePlayer(deps *Deps, userID uint32, u *User) {
	if deps == nil || deps.Store == nil || u == nil {
		return
	}
	player := buildPlayerUpdate(u, int64(userID))
	if player == nil {
		return
	}
	if u.PlayerID == 0 {
		p, err := deps.Store.CreatePlayer(context.Background(), player)
		if err == nil && p != nil {
			u.PlayerID = p.ID
		}
		return
	}
	_, _ = deps.Store.UpdatePlayer(context.Background(), player)
}
