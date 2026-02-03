package loginserver

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"jseer/internal/config"
	"jseer/internal/storage"
)

// RegisterHandlers wires login commands.
func RegisterHandlers(s *Server) {
	s.Register(1, handleVerify())
	s.Register(2, handleRegister())
	s.Register(3, handleSendEmailCode())
	s.Register(103, handleLegacyLogin())
	s.Register(104, handleMainLogin())
	s.Register(105, handleGoodServerList())
	s.Register(106, handleServerList())
	s.Register(108, handleCreateRole())
	s.Register(109, handleSysRole())
	s.Register(111, handleFenghaoTime())
}

func handleVerify() Handler {
	return func(ctx *Context) {
		ctx.Server.SendResponse(ctx.Conn, 1, ctx.UserID, 0, nil)
	}
}

func handleRegister() Handler {
	return func(ctx *Context) {
		password := readFixedString(ctx.Body, 0, 32)
		email := readFixedString(ctx.Body, 32, 64)
		if email == "" {
			ctx.Server.SendResponse(ctx.Conn, 2, 0, 1, nil)
			return
		}
		_, err := ctx.Server.store.GetAccountByEmail(ctx.Context(), email)
		if err == nil {
			ctx.Server.SendResponse(ctx.Conn, 2, 0, 1, nil)
			return
		}
		acct, err := ctx.Server.store.CreateAccount(ctx.Context(), &storage.Account{
			Email:    email,
			Password: password,
			Salt:     "",
			Status:   "active",
		})
		if err != nil {
			ctx.Server.SendResponse(ctx.Conn, 2, 0, 1, nil)
			return
		}
		ctx.Server.SendResponse(ctx.Conn, 2, uint32(acct.ID), 0, nil)
	}
}

func handleSendEmailCode() Handler {
	return func(ctx *Context) {
		code := randomHex(32)
		body := make([]byte, 32)
		copy(body, []byte(code))
		ctx.Server.SendResponse(ctx.Conn, 3, ctx.UserID, 0, body)
	}
}

func handleLegacyLogin() Handler {
	return func(ctx *Context) {
		session := randomBytes(16)
		body := append(session, make([]byte, 4)...) // roleCreate=1 by default (0/1)
		body[len(body)-1] = 1
		ctx.Server.SendResponse(ctx.Conn, 103, ctx.UserID, 0, body)
	}
}

func handleMainLogin() Handler {
	return func(ctx *Context) {
		email := readFixedString(ctx.Body, 0, 64)
		passMD5 := readFixedString(ctx.Body, 64, 32)
		if email == "" {
			ctx.Server.SendResponse(ctx.Conn, 104, 0, 1, nil)
			return
		}

		acct, err := ctx.Server.store.GetAccountByEmail(ctx.Context(), email)
		if err != nil {
			acct, err = ctx.Server.store.CreateAccount(ctx.Context(), &storage.Account{
				Email:    email,
				Password: passMD5,
				Salt:     "",
				Status:   "active",
			})
			if err != nil {
				ctx.Server.SendResponse(ctx.Conn, 104, 0, 1, nil)
				return
			}
		}

		if acct.Password != "" {
			stored := strings.ToLower(acct.Password)
			if stored != strings.ToLower(passMD5) {
				// allow if stored is plaintext by hashing it
				h := md5.Sum([]byte(acct.Password))
				if hex.EncodeToString(h[:]) != strings.ToLower(passMD5) {
					ctx.Server.SendResponse(ctx.Conn, 104, uint32(acct.ID), 5003, makeLoginBody(randomBytes(16), false))
					return
				}
			}
		}

		player, _ := ctx.Server.store.GetPlayerByAccount(ctx.Context(), acct.ID)
		roleCreated := player != nil

		session := randomBytes(16)
		body := makeLoginBody(session, roleCreated)
		ctx.Server.SendResponse(ctx.Conn, 104, uint32(acct.ID), 0, body)
	}
}

func handleCreateRole() Handler {
	return func(ctx *Context) {
		if len(ctx.Body) < 24 {
			ctx.Server.SendResponse(ctx.Conn, 108, ctx.UserID, 1, nil)
			return
		}
		nickname := readFixedString(ctx.Body, 4, 16)
		if nickname == "" {
			nickname = fmtUserID(ctx.UserID)
		}
		color := int(readUint32(ctx.Body, 20))

		_, err := ctx.Server.store.CreatePlayer(ctx.Context(), &storage.Player{
			Account: int64(ctx.UserID),
			Nick:    nickname,
			Level:   1,
			Coins:   10000,
			Gold:    0,
			MapID:   1,
			PosX:    300,
			PosY:    300,
		})
		if err != nil {
			ctx.Server.SendResponse(ctx.Conn, 108, ctx.UserID, 1, nil)
			return
		}

		_ = color // reserved for later persistence
		session := randomBytes(16)
		ctx.Server.SendResponse(ctx.Conn, 108, ctx.UserID, 0, session)
	}
}

func handleSysRole() Handler {
	return func(ctx *Context) {
		ctx.Server.SendResponse(ctx.Conn, 109, ctx.UserID, 0, nil)
	}
}

func handleFenghaoTime() Handler {
	return func(ctx *Context) {
		body := make([]byte, 4)
		ctx.Server.SendResponse(ctx.Conn, 111, ctx.UserID, 0, body)
	}
}

func handleGoodServerList() Handler {
	return func(ctx *Context) {
		body := buildGoodServerList(ctx.Server.gameCfg, 0)
		ctx.Server.SendResponse(ctx.Conn, 105, ctx.UserID, 0, body)
	}
}

func handleServerList() Handler {
	return func(ctx *Context) {
		body := buildServerList(ctx.Server.gameCfg)
		ctx.Server.SendResponse(ctx.Conn, 106, ctx.UserID, 0, body)
	}
}

func readFixedString(data []byte, offset, length int) string {
	if offset >= len(data) {
		return ""
	}
	end := offset + length
	if end > len(data) {
		end = len(data)
	}
	segment := data[offset:end]
	if idx := strings.IndexByte(string(segment), 0); idx >= 0 {
		segment = segment[:idx]
	}
	return strings.TrimRight(string(segment), "\x00")
}

func readUint32(data []byte, offset int) uint32 {
	if offset+4 > len(data) {
		return 0
	}
	return binary.BigEndian.Uint32(data[offset:])
}

func makeLoginBody(session []byte, roleCreated bool) []byte {
	body := make([]byte, 20)
	copy(body[:16], session)
	if roleCreated {
		binary.BigEndian.PutUint32(body[16:20], 1)
	} else {
		binary.BigEndian.PutUint32(body[16:20], 0)
	}
	return body
}

func buildGoodServerList(cfg config.GameConfig, onlineCount uint32) []byte {
	// maxOnlineID(4) + isVIP(4) + onlineCnt(4) + serverInfo(30) + friendData(8)
	body := make([]byte, 12+30+8)
	binary.BigEndian.PutUint32(body[0:4], uint32(cfg.ServerID))
	binary.BigEndian.PutUint32(body[4:8], 0)
	binary.BigEndian.PutUint32(body[8:12], 1)

	off := 12
	binary.BigEndian.PutUint32(body[off:off+4], uint32(cfg.ServerID))
	binary.BigEndian.PutUint32(body[off+4:off+8], onlineCount)
	writeFixedString(body, off+8, 16, cfg.PublicIP)
	binary.BigEndian.PutUint16(body[off+24:off+26], uint16(cfg.Port))
	binary.BigEndian.PutUint32(body[off+26:off+30], 1) // friends

	off += 30
	binary.BigEndian.PutUint32(body[off:off+4], 0)   // friend count
	binary.BigEndian.PutUint32(body[off+4:off+8], 0) // black count
	return body
}

func buildServerList(cfg config.GameConfig) []byte {
	body := make([]byte, 4+30)
	binary.BigEndian.PutUint32(body[0:4], 1)
	off := 4
	binary.BigEndian.PutUint32(body[off:off+4], uint32(cfg.ServerID))
	binary.BigEndian.PutUint32(body[off+4:off+8], 0)
	writeFixedString(body, off+8, 16, cfg.PublicIP)
	binary.BigEndian.PutUint16(body[off+24:off+26], uint16(cfg.Port))
	binary.BigEndian.PutUint32(body[off+26:off+30], 1)
	return body
}

func writeFixedString(buf []byte, offset, length int, s string) {
	for i := 0; i < length; i++ {
		if i < len(s) {
			buf[offset+i] = s[i]
		} else {
			buf[offset+i] = 0
		}
	}
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(256))
	}
	return b
}

func randomHex(n int) string {
	b := randomBytes(n / 2)
	return hex.EncodeToString(b)[:n]
}

func fmtUserID(id uint32) string {
	return strconv.FormatUint(uint64(id), 10)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
