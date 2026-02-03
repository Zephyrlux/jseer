package game

import (
	"bytes"
	"context"
	"encoding/binary"
	"time"

	"jseer/internal/gateway"
	"jseer/internal/storage"
)

func registerMiscHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(50004, handleXinCheck())
	s.Register(50008, handleXinGetQuadrupleExeTime())
	s.Register(8005, handleSyncTime())
	s.Register(8006, handleVipCo(state))
	s.Register(8007, handleVipLevelUp(state))
	s.Register(8008, handleMailNewNote(state))
	s.Register(8009, handleMedalGetCount())
	s.Register(8010, handleSprintGiftNotice())
	s.Register(6001, handleWorkConnection(state))
	s.Register(6003, handleAllConnection(state))
	s.Register(7001, handleUserReport(deps))
	s.Register(7002, handleUserContribute(deps))
	s.Register(7003, handleUserIndagate(deps))
	s.Register(70001, handleGetExchangeInfo())
}

func handleXinCheck() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, 50004, ctx.UserID, []byte{})
	}
}

func handleXinGetQuadrupleExeTime() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 50008, ctx.UserID, buf.Bytes())
	}
}

func handleGetExchangeInfo() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 70001, ctx.UserID, buf.Bytes())
	}
}

func handleSyncTime() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(time.Now().Unix()))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 8005, ctx.UserID, buf.Bytes())
	}
}

func handleVipCo(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		flag := uint32(0)
		if user.Nono.SuperNono > 0 {
			flag = 2
		}
		binary.Write(buf, binary.BigEndian, flag)
		binary.Write(buf, binary.BigEndian, user.Nono.AutoCharge)
		endTime := user.Nono.VipEndTime
		if user.Nono.SuperNono > 0 && endTime == 0 {
			endTime = 0x7FFFFFFF
		}
		binary.Write(buf, binary.BigEndian, endTime)
		ctx.Server.SendResponse(ctx.Conn, 8006, ctx.UserID, buf.Bytes())
	}
}

func handleVipLevelUp(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, user.Nono.VipLevel)
		ctx.Server.SendResponse(ctx.Conn, 8007, ctx.UserID, buf.Bytes())
	}
}

func handleMailNewNote(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		unread := uint32(0)
		for _, m := range user.Mailbox {
			if !m.Read {
				unread++
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, unread)
		ctx.Server.SendResponse(ctx.Conn, 8008, ctx.UserID, buf.Bytes())
	}
}

func handleMedalGetCount() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 8009, ctx.UserID, buf.Bytes())
	}
}

func handleSprintGiftNotice() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 8010, ctx.UserID, buf.Bytes())
	}
}

func handleWorkConnection(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		count := uint32(state.OnlineCount())
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, count)
		ctx.Server.SendResponse(ctx.Conn, 6001, ctx.UserID, buf.Bytes())
	}
}

func handleAllConnection(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		count := uint32(state.OnlineCount())
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, count)
		ctx.Server.SendResponse(ctx.Conn, 6003, ctx.UserID, buf.Bytes())
	}
}

func handleUserReport(deps *Deps) gateway.Handler {
	return func(ctx *gateway.Context) {
		writeAudit(deps, ctx, "user_report")
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 7001, ctx.UserID, buf.Bytes())
	}
}

func handleUserContribute(deps *Deps) gateway.Handler {
	return func(ctx *gateway.Context) {
		writeAudit(deps, ctx, "user_contribute")
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 7002, ctx.UserID, buf.Bytes())
	}
}

func handleUserIndagate(deps *Deps) gateway.Handler {
	return func(ctx *gateway.Context) {
		writeAudit(deps, ctx, "user_indagate")
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 7003, ctx.UserID, buf.Bytes())
	}
}

func writeAudit(deps *Deps, ctx *gateway.Context, action string) {
	if ctx == nil {
		return
	}
	if deps == nil || deps.Store == nil {
		return
	}
	_, _ = deps.Store.CreateAuditLog(context.Background(), &storage.AuditLog{
		Operator:  "user",
		Action:    action,
		Resource:  "client",
		ResourceID: itoa(ctx.UserID),
		Detail:    string(ctx.Body),
	})
}
