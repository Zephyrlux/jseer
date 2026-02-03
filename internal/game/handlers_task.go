package game

import (
	"bytes"
	"encoding/binary"

	"jseer/internal/gateway"
)

func registerTaskHandlers(s *gateway.Server, deps *Deps, state *State) {
	s.Register(2201, handleAcceptTask(deps, state))
	s.Register(2202, handleCompleteTask(deps, state))
	s.Register(2203, handleGetTaskBuf(state))
	s.Register(2204, handleAddTaskBuf(deps, state))
	s.Register(2205, handleDeleteTask(deps, state))
	s.Register(2206, handleChangeTaskStatus(deps, state))
	s.Register(2232, handleDeleteDailyTask(deps, state))
	s.Register(2233, handleCompleteDailyTask(deps, state))
	s.Register(2234, handleGetDailyTaskBuf())
	s.Register(2235, handleAddDailyTaskBuf())
}

func handleAcceptTask(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		taskID := int(reader.ReadUint32BE())
		user := state.GetOrCreateUser(ctx.UserID)
		if user.TaskStatus == nil {
			user.TaskStatus = make(map[int]byte)
		}
		user.TaskStatus[taskID] = 1
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(taskID))
		ctx.Server.SendResponse(ctx.Conn, 2201, ctx.UserID, buf.Bytes())
	}
}

func handleCompleteTask(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		taskID := int(reader.ReadUint32BE())
		param := 0
		if reader.Remaining() >= 4 {
			param = int(reader.ReadUint32BE())
		}
		user := state.GetOrCreateUser(ctx.UserID)
		body, _ := buildTaskCompleteResponse(taskID, param, user, deps)
		if user.TaskStatus == nil {
			user.TaskStatus = make(map[int]byte)
		}
		user.TaskStatus[taskID] = 3
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 2202, ctx.UserID, body)
	}
}

func handleGetTaskBuf(state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		taskID := int(reader.ReadUint32BE())
		user := state.GetOrCreateUser(ctx.UserID)
		if user.TaskBufs == nil {
			user.TaskBufs = make(map[int]map[int]uint32)
		}
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(taskID))
		binary.Write(buf, binary.BigEndian, uint32(1))
		taskBuf := user.TaskBufs[taskID]
		for i := 0; i <= 4; i++ {
			val := uint32(0)
			if taskBuf != nil {
				val = taskBuf[i]
			}
			binary.Write(buf, binary.BigEndian, val)
		}
		ctx.Server.SendResponse(ctx.Conn, 2203, ctx.UserID, buf.Bytes())
	}
}

func handleAddTaskBuf(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		taskID := int(reader.ReadUint32BE())
		index := 0
		if reader.Remaining() >= 1 {
			index = int(reader.ReadBytes(1)[0])
		}
		value := uint32(0)
		if reader.Remaining() >= 4 {
			value = reader.ReadUint32BE()
		}
		user := state.GetOrCreateUser(ctx.UserID)
		if user.TaskBufs == nil {
			user.TaskBufs = make(map[int]map[int]uint32)
		}
		if user.TaskBufs[taskID] == nil {
			user.TaskBufs[taskID] = make(map[int]uint32)
		}
		user.TaskBufs[taskID][index] = value
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 2204, ctx.UserID, []byte{})
	}
}

func handleGetDailyTaskBuf() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2234, ctx.UserID, buf.Bytes())
	}
}

func handleDeleteTask(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		taskID := int(reader.ReadUint32BE())
		user := state.GetOrCreateUser(ctx.UserID)
		if user.TaskStatus != nil {
			delete(user.TaskStatus, taskID)
		}
		if user.TaskBufs != nil {
			delete(user.TaskBufs, taskID)
		}
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(taskID))
		ctx.Server.SendResponse(ctx.Conn, 2205, ctx.UserID, buf.Bytes())
	}
}

func handleChangeTaskStatus(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		taskID := int(reader.ReadUint32BE())
		status := byte(0)
		if reader.Remaining() >= 4 {
			status = byte(reader.ReadUint32BE())
		}
		user := state.GetOrCreateUser(ctx.UserID)
		if user.TaskStatus == nil {
			user.TaskStatus = make(map[int]byte)
		}
		user.TaskStatus[taskID] = status
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(taskID))
		binary.Write(buf, binary.BigEndian, uint32(status))
		ctx.Server.SendResponse(ctx.Conn, 2206, ctx.UserID, buf.Bytes())
	}
}

func handleDeleteDailyTask(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		taskID := int(reader.ReadUint32BE())
		user := state.GetOrCreateUser(ctx.UserID)
		if user.TaskStatus != nil {
			delete(user.TaskStatus, taskID)
		}
		savePlayer(deps, ctx.UserID, user)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(taskID))
		ctx.Server.SendResponse(ctx.Conn, 2232, ctx.UserID, buf.Bytes())
	}
}

func handleCompleteDailyTask(deps *Deps, state *State) gateway.Handler {
	return func(ctx *gateway.Context) {
		reader := NewReader(ctx.Body)
		taskID := int(reader.ReadUint32BE())
		param := 0
		if reader.Remaining() >= 4 {
			param = int(reader.ReadUint32BE())
		}
		user := state.GetOrCreateUser(ctx.UserID)
		body, _ := buildTaskCompleteResponse(taskID, param, user, deps)
		if user.TaskStatus == nil {
			user.TaskStatus = make(map[int]byte)
		}
		user.TaskStatus[taskID] = 3
		savePlayer(deps, ctx.UserID, user)
		ctx.Server.SendResponse(ctx.Conn, 2233, ctx.UserID, body)
	}
}

func handleAddDailyTaskBuf() gateway.Handler {
	return func(ctx *gateway.Context) {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(0))
		ctx.Server.SendResponse(ctx.Conn, 2235, ctx.UserID, buf.Bytes())
	}
}

func buildTaskCompleteResponse(taskID int, param int, user *User, deps *Deps) ([]byte, int) {
	cfg := GetTaskConfig(taskID)
	if cfg == nil {
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.BigEndian, uint32(taskID))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		binary.Write(buf, binary.BigEndian, uint32(0))
		return buf.Bytes(), 0
	}

	responseItems := make([]TaskRewardItem, 0, 8)
	petID := 0
	captureTm := 0

	if cfg.Type == "select_pet" {
		petID = cfg.ParamMap[param]
		if petID == 0 {
			if param > 0 {
				petID = param
			} else {
				petID = 1
			}
		}
		captureTm = 0x69686700 + petID
		user.CurrentPetID = uint32(petID)
		user.CatchID = uint32(captureTm)
		pet := createStarterPet(petID, 5)
		if pet != nil {
			newPet := Pet{
				ID:        uint32(petID),
				CatchTime: uint32(captureTm),
				Level:     uint32(pet.Level),
				DV:        uint32(pet.DV),
				Exp:       pet.Exp,
				HP:        pet.HP,
				Skills:    pet.Skills,
			}
			user.Pets = append(user.Pets, newPet)
			upsertPet(deps, user, newPet)
			user.PetDV = uint32(pet.DV)
		}
	} else if cfg.Rewards.PetID > 0 {
		petID = cfg.Rewards.PetID
	}

	if user.Items == nil {
		user.Items = make(map[int]*ItemInfo)
	}
	for _, it := range cfg.Rewards.Items {
		responseItems = append(responseItems, it)
		info := user.Items[it.ID]
		if info == nil {
			info = &ItemInfo{Count: 0, ExpireTime: defaultItemExpire}
			user.Items[it.ID] = info
		}
		info.Count += it.Count
		upsertItem(deps, user, it.ID)
	}

	for _, spec := range cfg.Rewards.Special {
		responseItems = append(responseItems, TaskRewardItem{ID: spec.Type, Count: spec.Value})
		if spec.Type == 1 {
			user.Coins += uint32(spec.Value)
		}
	}

	if cfg.Rewards.Coins > 0 {
		responseItems = append(responseItems, TaskRewardItem{ID: 1, Count: cfg.Rewards.Coins})
		user.Coins += uint32(cfg.Rewards.Coins)
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(taskID))
	binary.Write(buf, binary.BigEndian, uint32(petID))
	binary.Write(buf, binary.BigEndian, uint32(captureTm))
	binary.Write(buf, binary.BigEndian, uint32(len(responseItems)))
	for _, item := range responseItems {
		binary.Write(buf, binary.BigEndian, uint32(item.ID))
		binary.Write(buf, binary.BigEndian, uint32(item.Count))
	}
	return buf.Bytes(), petID
}
