package game

type FightState struct {
	UserID        uint32
	PlayerPetID   uint32
	PlayerLevel   uint32
	PlayerHP      int
	PlayerMaxHP   int
	PlayerCatch   uint32
	PlayerSkills  []int
	PlayerStats   petStats
	PlayerType    int
	PlayerStage   stageModifiers
	EnemyPetID    uint32
	EnemyLevel    uint32
	EnemyHP       int
	EnemyMaxHP    int
	EnemyCatch    uint32
	EnemySkills   []int
	EnemyStats    petStats
	EnemyType     int
	EnemyStage    stageModifiers
	EnemyRewardID int
	EnemyRewardNm string
	EnemyRewardCt int
	PlayerFatigue int
	EnemyFatigue  int
	PlayerStatus  map[int]int
	EnemyStatus   map[int]int
	Turn          int
}
