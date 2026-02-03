package game

import (
	"jseer/internal/gateway"
	"jseer/internal/storage"

	"go.uber.org/zap"
)

type Deps struct {
	Logger     *zap.Logger
	State      *State
	GameIP     string
	GamePort   int
	Store      storage.Store
	SpawnMap   uint32
	SpawnX     uint32
	SpawnY     uint32
	ForceSpawn bool
}

func RegisterHandlers(s *gateway.Server, deps *Deps) {
	state := deps.State
	if state == nil {
		state = NewState()
	}
	registerSystemHandlers(s, deps, state)
	registerNonoHandlers(s, deps, state)
	registerPetHandlers(s, deps, state)
	registerPetAdvancedHandlers(s, deps, state)
	registerMapHandlers(s, deps, state)
	registerRoomHandlers(s, deps, state)
	registerTaskHandlers(s, deps, state)
	registerItemHandlers(s, deps, state)
	registerFriendHandlers(s, deps, state)
	registerExchangeHandlers(s)
	registerTeamHandlers(s, deps, state)
	registerTeamPKHandlers(s)
	registerTeacherHandlers(s, deps, state)
	registerMailHandlers(s, deps, state)
	registerAchievementHandlers(s, deps, state)
	registerMiscHandlers(s, deps, state)
	registerArenaHandlers(s)
	registerFightHandlers(s, deps, state)
	registerGameHandlers(s)
	registerCompatHandlers(s, deps, state)
	registerStubHandlers(s)

	s.SetDefault(handleStubEmpty())
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
		1003, 2051, 2052, 2053, 2054, 2055, 2148, 2149, 2151, 2159,
		2302, 2306, 2307, 2308, 2309, 2310, 2311, 2312, 2313, 2314,
		2315, 2316, 2317, 2320, 2321, 2322, 2323, 2324, 2327, 2328,
		2329, 2330, 2331, 2332, 2343, 2351, 2352, 2353, 2356, 2357,
		2358, 2393, 2414, 2415, 2416, 2417, 2418, 2419, 2420, 2421,
		2422, 2423, 2424, 2425, 2426, 2428, 2429, 2430, 2442, 2444,
		2445, 2446, 2481, 2801, 2821, 2851, 2852, 2910, 2911, 2912,
		2913, 2914, 2915, 2916, 2917, 2918, 2920, 2921, 2922, 2923,
		2924, 2925, 2926, 2927, 2928, 2929, 2930, 2931, 2932, 2933,
		2934, 2935, 2936, 2941, 2942, 2943, 2944, 2951, 2952, 2953,
		2954, 2961, 2962, 2963, 2964, 2965, 2966, 2967, 2968, 2969,
		2970, 3001, 3002, 3003, 3004, 3005, 3006, 3007, 3008, 3009,
		3010, 3011, 3201, 3301, 3401, 3402, 3403, 3404, 3405, 3406,
		3407, 4001, 4002, 4003, 4004, 4005, 4006, 4007, 4008, 4009,
		4010, 4011, 4012, 4013, 4014, 4017, 4018, 4019, 4020, 4022,
		4023, 4024, 4025, 4101, 4102, 5001, 5002, 6001, 6003, 7001,
		7002, 7003, 7501, 7502, 8005, 8006, 8007, 8008, 8009, 8010,
		9757, 10001, 10002, 10003, 10004, 10005, 10007, 10008, 10009, 30000,
		50001, 50003, 50005, 50006, 50007, 50009, 50010, 50011, 50012, 50013,
		50014, 50015, 52102, 70000, 70002, 80002, 80003, 80004, 80005, 80006,
		80007, 80008, 500000,
	}
	for _, cmd := range stub4Zero {
		s.RegisterIfAbsent(cmd, handleStub4Zero())
	}

	// 8-byte zero responses.
	// (handled by specific handler where applicable)

	// Empty responses.
	stubEmpty := []int32{
		1011, 1016, 2108, 2192, 2196, 2289, 2361, 3405, 4148, 4178,
		4181, 4359, 4364, 4501, 5003, 5005, 8002, 9112, 9677, 40006,
		40007, 41006, 41249, 41253, 43706, 45071, 45512, 45524, 45773, 45793,
		45798, 45824, 47309,
	}
	for _, cmd := range stubEmpty {
		s.RegisterIfAbsent(cmd, handleStubEmpty())
	}
}
