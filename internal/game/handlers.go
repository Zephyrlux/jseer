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
	registerMailHandlers(s)
	registerMiscHandlers(s)
	registerArenaHandlers(s)
	registerFightHandlers(s, deps, state)
	registerGameHandlers(s)
	registerCompatHandlers(s)
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
		9757,
		2054, 2055,
		2393,
		3401, 3402, 3403, 3406, 3407,
	}
	for _, cmd := range stub4Zero {
		s.Register(cmd, handleStub4Zero())
	}

	// 8-byte zero responses.
	// (handled by specific handler where applicable)

	// Empty responses.
	stubEmpty := []int32{
		1011, 1016, 2289, 2192, 2196, 2361, 3405, 4359, 4364, 4501, 5005,
		9112, 9677, 41006, 41249, 41253, 4148, 4178, 4181, 43706, 45512, 45524,
		45773, 45793, 45798, 45824, 47309, 45071, 40006, 40007,
	}
	for _, cmd := range stubEmpty {
		s.Register(cmd, handleStubEmpty())
	}
}
