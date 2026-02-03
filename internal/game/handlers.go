package game

import (
	"jseer/internal/gateway"
	"jseer/internal/storage"

	"go.uber.org/zap"
)

type Deps struct {
	Logger   *zap.Logger
	State    *State
	GameIP   string
	GamePort int
	Store    storage.Store
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
		1003,
		9757,
		2148, 2149,
		2317,
		2054, 2055,
		2393,
		2801, 2821, 2851, 2852,
		2915, 2916,
		2920, 2921, 2922, 2923, 2924, 2925, 2926, 2927,
		2930, 2931, 2932, 2933, 2934, 2935, 2936,
		2941, 2942, 2943, 2944,
		2951, 2952, 2953, 2954,
		2961, 2964, 2965, 2966, 2967, 2968, 2969, 2970,
		3301,
		3401, 3402, 3403, 3406, 3407,
		6001, 6003,
		7001, 7002, 7003,
		7501, 7502,
		8005, 8006, 8007, 8008, 8009, 8010,
		30000,
		50001, 50003, 50005, 50006, 50007, 50009, 50010, 50011, 50012, 50013, 50014, 50015,
		500000,
		52102,
		70000, 70002,
		80002, 80003, 80004, 80005, 80006, 80007, 80008,
	}
	for _, cmd := range stub4Zero {
		s.RegisterIfAbsent(cmd, handleStub4Zero())
	}

	// 8-byte zero responses.
	// (handled by specific handler where applicable)

	// Empty responses.
	stubEmpty := []int32{
		1011, 1016, 2108, 2289, 2192, 2196, 2361, 3405, 4359, 4364, 4501, 5005,
		9112, 9677, 41006, 41249, 41253, 4148, 4178, 4181, 43706, 45512, 45524,
		45773, 45793, 45798, 45824, 47309, 45071, 40006, 40007,
		8002,
	}
	for _, cmd := range stubEmpty {
		s.RegisterIfAbsent(cmd, handleStubEmpty())
	}
}
