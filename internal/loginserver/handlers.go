package loginserver

import (
	"crypto/md5"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"jseer/internal/config"
	"jseer/internal/storage"

	"go.uber.org/zap"
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
			Salt:     makeSalt(),
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
		ctx.Server.logger.Info("email code", zap.String("code", code), zap.String("ip", remoteIP(ctx.Conn)))
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
			email = readFixedString(ctx.Body, 4, 64)
		}
		if passMD5 == "" {
			passMD5 = readFixedString(ctx.Body, 4+64, 32)
		}
		if email == "" {
			email = readFixedString(ctx.Body, 8, 64)
		}
		if passMD5 == "" {
			passMD5 = readFixedString(ctx.Body, 8+64, 32)
		}
		if email == "" {
			email = findEmailInBody(ctx.Body)
		}
		if passMD5 == "" {
			passMD5 = findMD5InBody(ctx.Body)
		}
		if email == "" || passMD5 == "" {
			ctx.Server.logger.Warn(
				"login parse empty",
				zap.Int("body_len", len(ctx.Body)),
				zap.String("email", email),
				zap.Int("pass_len", len(passMD5)),
				zap.String("body_hex", hexDump(ctx.Body, 96)),
				zap.String("body_ascii", asciiPreview(ctx.Body, 96)),
				zap.String("ip", remoteIP(ctx.Conn)),
			)
		} else {
			ctx.Server.logger.Info(
				"login parse ok",
				zap.String("email", email),
				zap.Int("pass_len", len(passMD5)),
				zap.String("ip", remoteIP(ctx.Conn)),
			)
		}
		if email == "" {
			ctx.Server.logger.Warn("login reject empty email", zap.String("ip", remoteIP(ctx.Conn)))
			ctx.Server.SendResponse(ctx.Conn, 104, 0, 1, nil)
			return
		}

		acct, err := ctx.Server.store.GetAccountByEmail(ctx.Context(), email)
		if err != nil {
			acct, err = ctx.Server.store.CreateAccount(ctx.Context(), &storage.Account{
				Email:    email,
				Password: passMD5,
				Salt:     makeSalt(),
				Status:   "active",
			})
			if err != nil {
				ctx.Server.logger.Warn("login create account failed", zap.Error(err), zap.String("ip", remoteIP(ctx.Conn)))
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

		cfg := loadDefaultPlayerConfig()
		level := cfg.Player.Level
		if level == 0 {
			level = 1
		}
		coins := cfg.Player.Coins
		if coins == 0 {
			coins = 2000
		}
		mapID := cfg.Player.MapID
		if mapID == 0 {
			mapID = 1
		}
		posX := cfg.Player.PosX
		posY := cfg.Player.PosY
		if posX == 0 {
			posX = 300
		}
		if posY == 0 {
			posY = 270
		}
		timeLimit := cfg.Player.TimeLimit
		if timeLimit == 0 {
			timeLimit = 86400
		}

		_, err := ctx.Server.store.CreatePlayer(ctx.Context(), &storage.Player{
			Account:   int64(ctx.UserID),
			Nick:      nickname,
			Level:     level,
			Coins:     coins,
			Gold:      cfg.Player.Gold,
			MapID:     mapID,
			PosX:      posX,
			PosY:      posY,
			TimeLimit: timeLimit,
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
	if s := decodeMaybeUTF16(segment); s != "" {
		return s
	}
	raw := string(segment)
	if idx := strings.IndexByte(raw, 0); idx >= 0 {
		raw = raw[:idx]
	}
	return strings.TrimRight(raw, "\x00")
}

func readUint32(data []byte, offset int) uint32 {
	if offset+4 > len(data) {
		return 0
	}
	return binary.BigEndian.Uint32(data[offset:])
}

func decodeMaybeUTF16(segment []byte) string {
	if len(segment) < 2 {
		return ""
	}
	zeroEven := 0
	zeroOdd := 0
	for i := 0; i < len(segment); i++ {
		if segment[i] == 0 {
			if i%2 == 0 {
				zeroEven++
			} else {
				zeroOdd++
			}
		}
	}
	// If most even or odd bytes are zero, treat as UTF-16.
	half := len(segment) / 2
	if zeroEven < half-2 && zeroOdd < half-2 {
		return ""
	}
	be := zeroEven > zeroOdd
	runes := make([]uint16, 0, half)
	for i := 0; i+1 < len(segment); i += 2 {
		var v uint16
		if be {
			v = uint16(segment[i])<<8 | uint16(segment[i+1])
		} else {
			v = uint16(segment[i+1])<<8 | uint16(segment[i])
		}
		if v == 0 {
			break
		}
		runes = append(runes, v)
	}
	if len(runes) == 0 {
		return ""
	}
	return string(utf16.Decode(runes))
}

func stripZeroBytes(data []byte) string {
	buf := make([]byte, 0, len(data))
	for _, b := range data {
		if b != 0 {
			buf = append(buf, b)
		}
	}
	return string(buf)
}

func hexDump(data []byte, max int) string {
	if max <= 0 || max > len(data) {
		max = len(data)
	}
	const hex = "0123456789abcdef"
	out := make([]byte, 0, max*2)
	for i := 0; i < max; i++ {
		b := data[i]
		out = append(out, hex[b>>4], hex[b&0x0f])
	}
	return string(out)
}

func asciiPreview(data []byte, max int) string {
	if max <= 0 || max > len(data) {
		max = len(data)
	}
	out := make([]byte, 0, max)
	for i := 0; i < max; i++ {
		b := data[i]
		if b >= 32 && b <= 126 {
			out = append(out, b)
		} else {
			out = append(out, '.')
		}
	}
	return string(out)
}

var (
	emailRe = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	md5Re   = regexp.MustCompile(`[a-fA-F0-9]{32}`)
)

func findEmailInBody(data []byte) string {
	if s := decodeMaybeUTF16(data); s != "" {
		if m := emailRe.FindString(s); m != "" {
			return m
		}
	}
	if m := emailRe.FindString(stripZeroBytes(data)); m != "" {
		return m
	}
	ascii := make([]byte, 0, len(data))
	for _, b := range data {
		if b >= 32 && b <= 126 {
			ascii = append(ascii, b)
		} else {
			ascii = append(ascii, ' ')
		}
	}
	if m := emailRe.FindString(string(ascii)); m != "" {
		return m
	}
	return ""
}

func findMD5InBody(data []byte) string {
	if s := decodeMaybeUTF16(data); s != "" {
		if m := md5Re.FindString(s); m != "" {
			return m
		}
	}
	if m := md5Re.FindString(stripZeroBytes(data)); m != "" {
		return m
	}
	ascii := make([]byte, 0, len(data))
	for _, b := range data {
		if b >= 32 && b <= 126 {
			ascii = append(ascii, b)
		} else {
			ascii = append(ascii, ' ')
		}
	}
	if m := md5Re.FindString(string(ascii)); m != "" {
		return m
	}
	return ""
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
	if n <= 0 {
		return ""
	}
	if n%2 != 0 {
		n++
	}
	b := make([]byte, n/2)
	if _, err := crand.Read(b); err != nil {
		for i := range b {
			b[i] = byte(rand.Intn(256))
		}
	}
	return hex.EncodeToString(b)[:n]
}

func makeSalt() string {
	salt := randomHex(32)
	if salt == "" {
		salt = "0000000000000000"
	}
	return salt
}

func fmtUserID(id uint32) string {
	return strconv.FormatUint(uint64(id), 10)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
