package game

import "sync"

type SPTBoss struct {
	Name         string `json:"name"`
	PetID        int    `json:"petId"`
	Level        int    `json:"level"`
	MaxHP        int    `json:"maxHp"`
	RewardName   string `json:"rewardName"`
	RewardItemID int    `json:"rewardItemId"`
	RewardCount  int    `json:"rewardCount"`
}

type sptBossFile struct {
	Bosses []*SPTBoss `json:"bosses"`
}

var defaultSptBosses = []*SPTBoss{
	{Name: "蘑菇怪", Level: 10, MaxHP: 100, RewardName: "小蘑菇"},
	{Name: "钢牙鲨", Level: 25, MaxHP: 200, RewardName: "黑晶矿"},
	{Name: "里奥斯", Level: 35, MaxHP: 338, RewardName: "里奥斯精元"},
	{Name: "阿克希亚", Level: 65, MaxHP: 1000, RewardName: "阿克希亚精元"},
	{Name: "提亚斯", Level: 50, MaxHP: 500, RewardName: "提亚斯精元"},
	{Name: "雷伊", Level: 70, MaxHP: 800, RewardName: "雷伊精元"},
	{Name: "纳多雷", Level: 70, MaxHP: 1400, RewardName: "纳多雷精元"},
	{Name: "雷纳多", Level: 75, MaxHP: 1500, RewardName: "雷纳多精元"},
	{Name: "尤纳斯", Level: 70, MaxHP: 2800, RewardName: "尤纳斯精元"},
	{Name: "魔狮迪露", Level: 50, MaxHP: 3000000, RewardName: "魔狮迪露精元"},
	{Name: "哈默雷特", Level: 80, MaxHP: 10000, RewardName: "哈默雷特精元"},
	{Name: "奈尼芬多", Level: 60, MaxHP: 2500, RewardName: "奈尼芬多精元"},
	{Name: "厄尔塞拉", Level: 80, MaxHP: 3000, RewardName: "厄尔塞拉精元"},
	{Name: "盖亚", Level: 80, MaxHP: 2000, RewardName: "盖亚精元"},
	{Name: "塔克林", Level: 80, MaxHP: 13000, RewardName: "塔克林精元"},
	{Name: "塔西亚", Level: 70, MaxHP: 10000, RewardName: "塔西亚精元"},
	{Name: "远古鱼龙"},
	{Name: "上古炎兽"},
	{Name: "普尼真身", Level: 100},
	{Name: "拂晓兔", Level: 80},
}

var (
	sptBossOnce sync.Once
	sptBossByID map[int]*SPTBoss
	sptBosses   []*SPTBoss
)

func GetSPTBossByID(petID int) *SPTBoss {
	sptBossOnce.Do(initSPTBoss)
	if petID <= 0 {
		return nil
	}
	return sptBossByID[petID]
}

func initSPTBoss() {
	sptBosses = defaultSptBosses
	var cfg sptBossFile
	if readConfigJSON("spt-boss.json", &cfg) && len(cfg.Bosses) > 0 {
		sptBosses = cfg.Bosses
	}
	sptBossByID = make(map[int]*SPTBoss)
	for _, boss := range sptBosses {
		if boss.PetID == 0 && boss.Name != "" {
			boss.PetID = findPetIDByName(boss.Name)
		}
		if boss.RewardItemID == 0 && boss.RewardName != "" {
			boss.RewardItemID = findItemIDByName(boss.RewardName)
		}
		if boss.RewardCount == 0 && boss.RewardItemID > 0 {
			boss.RewardCount = 1
		}
		if boss.PetID > 0 {
			sptBossByID[boss.PetID] = boss
		}
	}
}
