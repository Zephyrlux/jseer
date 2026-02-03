package game

import "sync"

type TaskRewardItem struct {
	ID    int
	Count int
}

type TaskSpecialReward struct {
	Type  int
	Value int
}

type TaskRewards struct {
	Items   []TaskRewardItem
	PetID   int
	Special []TaskSpecialReward
	Coins   int
}

type TaskConfig struct {
	ID           int
	Name         string
	Type         string
	ParamMap     map[int]int
	Rewards      TaskRewards
	TargetItemID int
}

var defaultTaskConfigs = map[int]*TaskConfig{
	85: {
		ID:   85,
		Name: "novice_gift",
		Rewards: TaskRewards{
			Items: []TaskRewardItem{
				{ID: 100027, Count: 1},
				{ID: 100028, Count: 1},
				{ID: 500001, Count: 1},
				{ID: 300650, Count: 3},
				{ID: 300025, Count: 3},
				{ID: 300035, Count: 3},
				{ID: 500502, Count: 1},
				{ID: 500503, Count: 1},
			},
		},
	},
	86: {
		ID:   86,
		Name: "novice_pet",
		Type: "select_pet",
		ParamMap: map[int]int{
			1: 1,
			2: 7,
			3: 4,
		},
	},
	87: {
		ID:   87,
		Name: "novice_battle",
		Rewards: TaskRewards{
			Items: []TaskRewardItem{
				{ID: 300001, Count: 5},
				{ID: 300011, Count: 3},
			},
		},
	},
	88: {
		ID:   88,
		Name: "novice_item",
		Rewards: TaskRewards{
			Coins: 50000,
			Special: []TaskSpecialReward{
				{Type: 1, Value: 50000},
				{Type: 3, Value: 250000},
				{Type: 5, Value: 20},
			},
		},
	},
	94: {
		ID:           94,
		Name:         "get_drill",
		Type:         "get_item",
		TargetItemID: 100014,
	},
}

var (
	taskConfigOnce sync.Once
	taskConfigMap  map[int]*TaskConfig
)

type taskConfigFile struct {
	Tasks map[string]taskConfigJSON `json:"tasks"`
}

type taskConfigJSON struct {
	Name         string                `json:"name"`
	Type         string                `json:"type"`
	ParamMap     map[string]int        `json:"paramMap"`
	TargetItemID int                   `json:"targetItemId"`
	Rewards      taskConfigRewardsJSON `json:"rewards"`
}

type taskConfigRewardsJSON struct {
	Items   []TaskRewardItem    `json:"items"`
	PetID   int                 `json:"petId"`
	Special []TaskSpecialReward `json:"special"`
	Coins   int                 `json:"coins"`
}

func GetTaskConfig(taskID int) *TaskConfig {
	taskConfigOnce.Do(loadTaskConfig)
	return taskConfigMap[taskID]
}

func loadTaskConfig() {
	taskConfigMap = make(map[int]*TaskConfig)
	for id, cfg := range defaultTaskConfigs {
		taskConfigMap[id] = cfg
	}

	var cfgFile taskConfigFile
	if !readConfigJSON("tasks.json", &cfgFile) {
		return
	}
	if len(cfgFile.Tasks) == 0 {
		return
	}
	taskConfigMap = make(map[int]*TaskConfig)
	for key, item := range cfgFile.Tasks {
		id := int(parseUint32(key))
		if id == 0 {
			continue
		}
		paramMap := make(map[int]int)
		for k, v := range item.ParamMap {
			ki := int(parseUint32(k))
			if ki == 0 {
				continue
			}
			paramMap[ki] = v
		}
		taskConfigMap[id] = &TaskConfig{
			ID:           id,
			Name:         item.Name,
			Type:         item.Type,
			ParamMap:     paramMap,
			Rewards:      TaskRewards{Items: item.Rewards.Items, PetID: item.Rewards.PetID, Special: item.Rewards.Special, Coins: item.Rewards.Coins},
			TargetItemID: item.TargetItemID,
		}
	}
}
