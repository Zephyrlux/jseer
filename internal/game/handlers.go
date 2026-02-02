package game

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"time"

	"jseer/internal/gateway"
	"jseer/internal/protocol"

	"go.uber.org/zap"
)

type Deps struct {
	Logger *zap.Logger
}

func RegisterHandlers(s *gateway.Server, deps *Deps) {
	s.Register(1001, handleLogin(deps))
	s.Register(1002, handleSystemTime())
	s.Register(2001, handleEnterMap())

	registerStubHandlers(s)

	s.SetDefault(handleStubEmpty())
}

func handleLogin(deps *Deps) gateway.Handler {
	return func(ctx *gateway.Context) {
		body := buildLoginBody(ctx.UserID)
		ctx.Server.SendResponse(ctx.Conn, 1001, ctx.UserID, body)
		if deps != nil && deps.Logger != nil {
			deps.Logger.Info("login placeholder sent", zap.Uint32("uid", ctx.UserID))
		}
	}
}

func handleSystemTime() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint32(buf[0:4], uint32(time.Now().Unix()))
		binary.BigEndian.PutUint32(buf[4:8], 0)
		ctx.Server.SendResponse(ctx.Conn, 1002, ctx.UserID, buf)
	}
}

func handleEnterMap() gateway.Handler {
	return func(ctx *gateway.Context) {
		// Minimal map entry response placeholder. Full protocol should be aligned with Lua client.
		buf := new(bytes.Buffer)
		protocol.WriteUint32BE(buf, 0) // ret
		ctx.Server.SendResponse(ctx.Conn, 2001, ctx.UserID, buf.Bytes())
	}
}

func handleStubEmpty() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, ctx.CmdID, ctx.UserID, []byte{})
	}
}

func handleStub4Zero() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, ctx.CmdID, ctx.UserID, make([]byte, 4))
	}
}

func handleStub8Zero() gateway.Handler {
	return func(ctx *gateway.Context) {
		ctx.Server.SendResponse(ctx.Conn, ctx.CmdID, ctx.UserID, make([]byte, 8))
	}
}

func registerStubHandlers(s *gateway.Server) {
	// 4-byte zero responses (aligned with Lua emptyResponse(4)).
	stub4Zero := []int32{
		5001, 5002, 3201, 9757, 2442, 2444, 2445, 2446,
		2053, 2054, 2055,
		2302, 2306, 2307, 2308, 2309, 2310, 2311, 2312, 2313, 2314, 2315, 2316,
		2320, 2321, 2322, 2323, 2324, 2327, 2328, 2329, 2330, 2331, 2332,
		2343, 2351, 2352, 2353, 2356, 2357, 2358, 2393,
		3401, 3402, 3403, 3406, 3407,
		2414, 2415, 2416, 2417, 2418, 2419, 2420, 2421, 2422, 2423, 2424, 2425, 2426, 2428, 2429, 2430,
		2910, 2911, 2912, 2913, 2914, 2917, 2918, 2928, 2929, 2962, 2963,
		3001, 3002, 3003, 3004, 3005, 3006, 3007, 3008, 3009, 3010, 3011,
		4001, 4002, 4003, 4004, 4005, 4006, 4007, 4008, 4009, 4010, 4011, 4012, 4013, 4014,
		4017, 4018, 4019, 4020, 4022, 4023, 4024, 4025, 4101, 4102, 2481,
		10001, 10002, 10003, 10004, 10005, 10007, 10008, 10009,
	}
	for _, cmd := range stub4Zero {
		s.Register(cmd, handleStub4Zero())
	}

	// 8-byte zero responses.
	s.Register(5052, handleStub8Zero())

	// Empty responses.
	stubEmpty := []int32{
		5003, 1011, 1016, 2289, 2192, 2196, 2361, 3405, 4359, 4364, 4501, 5005,
		9112, 9677, 41006, 41249, 41253, 4148, 4178, 4181, 43706, 45512, 45524,
		45773, 45793, 45798, 45824, 47309, 45071, 40006, 40007,
	}
	for _, cmd := range stubEmpty {
		s.Register(cmd, handleStubEmpty())
	}
}

func buildLoginBody(userID uint32) []byte {
	buf := new(bytes.Buffer)
	// session 16 bytes
	session := make([]byte, 16)
	_, _ = rand.Read(session)
	buf.Write(session)
	// keySeed 4 bytes
	binary.Write(buf, binary.BigEndian, uint32(12345))
	// userID
	binary.Write(buf, binary.BigEndian, userID)
	// regTime
	binary.Write(buf, binary.BigEndian, uint32(time.Now().Unix()))
	// nick 16 bytes
	protocol.WriteFixedString(buf, "Seer", 16)
	// decorate list 5 * uint32
	for i := 0; i < 5; i++ {
		binary.Write(buf, binary.BigEndian, uint32(0))
	}
	// placeholder fields (aligned with Lua structure, expanded later)
	for i := 0; i < 20; i++ {
		binary.Write(buf, binary.BigEndian, uint32(0))
	}
	return buf.Bytes()
}
