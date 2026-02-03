package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerTeamHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2910, handleTeamCreate(deps, state))
	s.Register(2911, handleTeamAdd())
	s.Register(2912, handleTeamAnswer())
	s.Register(2913, handleTeamInform())
	s.Register(2914, handleTeamQuit(deps, state))
	s.Register(2915, handleTeamStub4Zero())
	s.Register(2916, handleTeamStub4Zero())
	s.Register(2917, handleTeamGetInfo(state))
	s.Register(2918, handleTeamGetMemberList())
	s.Register(2920, handleTeamStub4Zero())
	s.Register(2921, handleTeamStub4Zero())
	s.Register(2922, handleTeamStub4Zero())
	s.Register(2923, handleTeamStub4Zero())
	s.Register(2924, handleTeamStub4Zero())
	s.Register(2925, handleTeamStub4Zero())
	s.Register(2926, handleTeamStub4Zero())
	s.Register(2927, handleTeamStub4Zero())
	s.Register(2928, handleTeamGetLogoInfo())
	s.Register(2929, handleTeamChat())
	s.Register(2930, handleTeamStub4Zero())
	s.Register(2931, handleTeamStub4Zero())
	s.Register(2932, handleTeamStub4Zero())
	s.Register(2933, handleTeamStub4Zero())
	s.Register(2934, handleTeamStub4Zero())
	s.Register(2935, handleTeamStub4Zero())
	s.Register(2936, handleTeamStub4Zero())
	s.Register(2941, handleTeamStub4Zero())
	s.Register(2942, handleTeamStub4Zero())
	s.Register(2943, handleTeamStub4Zero())
	s.Register(2944, handleTeamStub4Zero())
	s.Register(2951, handleTeamStub4Zero())
	s.Register(2952, handleTeamStub4Zero())
	s.Register(2953, handleTeamStub4Zero())
	s.Register(2954, handleTeamStub4Zero())
	s.Register(2961, handleTeamStub4Zero())
	s.Register(2962, handleArmUpWork())
	s.Register(2963, handleArmUpDonate())
	s.Register(2964, handleTeamStub4Zero())
	s.Register(2965, handleTeamStub4Zero())
	s.Register(2966, handleTeamStub4Zero())
	s.Register(2967, handleTeamStub4Zero())
	s.Register(2968, handleTeamStub4Zero())
	s.Register(2969, handleTeamStub4Zero())
	s.Register(2970, handleTeamStub4Zero())
}

func handleTeamCreate(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		teamID := uint32(10000) + ctx.UserID
		user.Team = TeamInfo{
			ID:                teamID,
			Priv:              4,
			IsShow:            true,
			LogoBg:            1,
			LogoIcon:          1,
			LogoColor:         0xFFFF,
			TxtColor:          0xFFFF,
			LogoWord:          "",
			AllContribution:   0,
			CanExContribution: 0,
			CoreCount:         0,
			MemberCount:       1,
		}
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, teamID)
		ctx.Server.SendResponse(ctx.Conn, 2910, ctx.UserID, buf.Bytes())
	}
}

func handleTeamAdd() gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		teamID := reader.ReadUint32BE()
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, teamID)
		ctx.Server.SendResponse(ctx.Conn, 2911, ctx.UserID, buf.Bytes())
	}
}

func handleTeamAnswer() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2912, ctx.UserID, []byte{})
	}
}

func handleTeamInform() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2913, ctx.UserID, []byte{})
	}
}

func handleTeamQuit(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		user.Team = TeamInfo{}
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 2914, ctx.UserID, []byte{})
	}
}

func handleTeamGetInfo(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		team := user.Team
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, team.ID)
		binary.Write(buf, binary.BigEndian, uint32(0)) // leader
		binary.Write(buf, binary.BigEndian, uint32(0)) // superCoreNum
		memberCount := team.MemberCount
		if memberCount == 0 && team.ID != 0 {
			memberCount = 1
		}
		binary.Write(buf, binary.BigEndian, memberCount)
		binary.Write(buf, binary.BigEndian, team.Interest)
		binary.Write(buf, binary.BigEndian, team.JoinFlag)
		if team.IsShow {
			binary.Write(buf, binary.BigEndian, uint32(1))
		} else {
			binary.Write(buf, binary.BigEndian, uint32(0))
		}
		binary.Write(buf, binary.BigEndian, team.Exp)
		binary.Write(buf, binary.BigEndian, team.Score)
		protocol.WriteFixedString(buf, team.Name, 16)
		protocol.WriteFixedString(buf, team.Slogan, 60)
		protocol.WriteFixedString(buf, team.Notice, 60)
		protocol.WriteUint16BE(buf, team.LogoBg)
		protocol.WriteUint16BE(buf, team.LogoIcon)
		protocol.WriteUint16BE(buf, team.LogoColor)
		protocol.WriteUint16BE(buf, team.TxtColor)
		protocol.WriteFixedString(buf, team.LogoWord, 4)
		ctx.Server.SendResponse(ctx.Conn, 2917, ctx.UserID, buf.Bytes())
	}
}

func handleTeamGetMemberList() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2918, ctx.UserID, buf.Bytes())
	}
}

func handleTeamGetLogoInfo() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		protocol.WriteUint16BE(buf, 0)
		protocol.WriteUint16BE(buf, 0)
		protocol.WriteUint16BE(buf, 0)
		protocol.WriteUint16BE(buf, 0)
		protocol.WriteFixedString(buf, "", 4)
		ctx.Server.SendResponse(ctx.Conn, 2928, ctx.UserID, buf.Bytes())
	}
}

func handleTeamChat() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 2929, ctx.UserID, []byte{})
	}
}

func handleTeamStub4Zero() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, ctx.CmdID, ctx.UserID, buf.Bytes())
	}
}

func handleArmUpWork() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2962, ctx.UserID, buf.Bytes())
	}
}

func handleArmUpDonate() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2963, ctx.UserID, buf.Bytes())
	}
}
