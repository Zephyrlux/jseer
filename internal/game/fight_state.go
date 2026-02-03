package game

type FightState struct {
	UserID        uint32
	PlayerPetID   uint32
	PlayerLevel   uint32
	PlayerHP      int
	PlayerMaxHP   int
	PlayerCatch   uint32
	PlayerSkills  []int
	EnemyPetID    uint32
	EnemyLevel    uint32
	EnemyHP       int
	EnemyMaxHP    int
	EnemyCatch    uint32
	EnemySkills   []int
	Turn          int
}
