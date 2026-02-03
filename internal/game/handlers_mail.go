package game

import (
	"bytes"
	"encoding/binary"
	"sync/atomic"
	"time"

	"jseer/internal/gateway"
	"jseer/internal/protocol"
)

func registerMailHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2751, handleMailGetList(state))
	s.Register(2752, handleMailSend(deps, state))
	s.Register(2753, handleMailGetContent(state))
	s.Register(2754, handleMailSetRead(deps, state))
	s.Register(2755, handleMailDelete(deps, state))
	s.Register(2756, handleMailDeleteAll(deps, state))
	s.Register(2757, handleMailGetUnread(state))
	s.Register(8001, handleInform())
	s.Register(8004, handleGetBossMonster())
}

func handleMailGetList(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		buf := new(bytes.Buffer)
		total := uint32(len(user.Mailbox))
		binary.Write(buf, binary.BigEndian, total)
		binary.Write(buf, binary.BigEndian, total)
		for _, m := range user.Mailbox {
			binary.Write(buf, binary.BigEndian, m.ID)
			binary.Write(buf, binary.BigEndian, m.SenderID)
			protocol.WriteFixedString(buf, m.SenderName, 16)
			protocol.WriteFixedString(buf, m.Title, 64)
			binary.Write(buf, binary.BigEndian, m.CreatedAt)
			if m.Read {
				binary.Write(buf, binary.BigEndian, uint32(1))
			} else {
				binary.Write(buf, binary.BigEndian, uint32(0))
			}
		}
		ctx.Server.SendResponse(ctx.Conn, 2751, ctx.UserID, buf.Bytes())
	}
}

func handleMailGetUnread(state *State) gateway.Handler {
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
		ctx.Server.SendResponse(ctx.Conn, 2757, ctx.UserID, buf.Bytes())
	}
}

var mailSeq uint32 = uint32(time.Now().Unix())

func handleMailSend(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		targetID := reader.ReadUint32BE()
		title := ""
		content := ""
		if reader.Remaining() >= 4 {
			titleLen := int(reader.ReadUint32BE())
			if titleLen > 0 && reader.Remaining() >= titleLen {
				title = string(reader.ReadBytes(titleLen))
			}
		}
		if reader.Remaining() >= 4 {
			contentLen := int(reader.ReadUint32BE())
			if contentLen > 0 && reader.Remaining() >= contentLen {
				content = string(reader.ReadBytes(contentLen))
			}
		}
		if targetID == 0 {
			targetID = ctx.UserID
		}
		sender := state.GetOrCreateUser(ctx.UserID)
		recipient := state.GetOrCreateUser(targetID)
		mail := Mail{
			ID:         nextMailID(),
			SenderID:   ctx.UserID,
			SenderName: pickNick(sender, ctx.UserID),
			Title:      title,
			Content:    content,
			CreatedAt:  uint32(time.Now().Unix()),
			Read:       false,
		}
		recipient.Mailbox = append([]Mail{mail}, recipient.Mailbox...)
		savePlayer(deps, targetID, recipient)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2752, ctx.UserID, buf.Bytes())
	}
}

func handleMailGetContent(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		mailID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		var found *Mail
		for i := range user.Mailbox {
			if user.Mailbox[i].ID == mailID {
				found = &user.Mailbox[i]
				break
			}
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, mailID)
		if found == nil {
			binary.Write(buf, binary.BigEndian, uint32(0))
			ctx.Server.SendResponse(ctx.Conn, 2753, ctx.UserID, buf.Bytes())
			return
		}
		protocol.WriteFixedString(buf, found.SenderName, 16)
		protocol.WriteFixedString(buf, found.Title, 64)
		binary.Write(buf, binary.BigEndian, uint32(len(found.Content)))
		if found.Content != "" {
			buf.Write([]byte(found.Content))
		}
		ctx.Server.SendResponse(ctx.Conn, 2753, ctx.UserID, buf.Bytes())
	}
}

func handleMailSetRead(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		mailID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		updated := false
		for i := range user.Mailbox {
			if user.Mailbox[i].ID == mailID {
				user.Mailbox[i].Read = true
				updated = true
				break
			}
		}
		if updated {
			savePlayer(deps, ctx.UserID, user)
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, mailID)
		ctx.Server.SendResponse(ctx.Conn, 2754, ctx.UserID, buf.Bytes())
	}
}

func handleMailDelete(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		mailID := reader.ReadUint32BE()
		user := state.GetOrCreateUser(ctx.UserID)
		if len(user.Mailbox) > 0 {
			next := user.Mailbox[:0]
			for _, m := range user.Mailbox {
				if m.ID != mailID {
					next = append(next, m)
				}
			}
			user.Mailbox = next
			savePlayer(deps, ctx.UserID, user)
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, mailID)
		ctx.Server.SendResponse(ctx.Conn, 2755, ctx.UserID, buf.Bytes())
	}
}

func handleMailDeleteAll(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		user := state.GetOrCreateUser(ctx.UserID)
		if len(user.Mailbox) > 0 {
			user.Mailbox = nil
			savePlayer(deps, ctx.UserID, user)
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2756, ctx.UserID, buf.Bytes())
	}
}

func nextMailID() uint32 {
	return atomic.AddUint32(&mailSeq, 1)
}

func handleInform() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, ctx.UserID)
		protocol.WriteFixedString(buf, "", 16)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(1))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(301))
		protocol.WriteFixedString(buf, "", 64)
		ctx.Server.SendResponse(ctx.Conn, 8001, ctx.UserID, buf.Bytes())
	}
}

func handleGetBossMonster() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		for i := 0; i < 4; i++ {
			binary.Write(buf, binary.BigEndian, uint32(0))
		}
		ctx.Server.SendResponse(ctx.Conn, 8004, ctx.UserID, buf.Bytes())
	}
}
