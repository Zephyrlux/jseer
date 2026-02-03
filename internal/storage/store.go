package storage

import "context"

type Store interface {
	Ping(ctx context.Context) error
	Close() error

	// Accounts & players
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	CreateAccount(ctx context.Context, in *Account) (*Account, error)
	GetPlayerByID(ctx context.Context, id int64) (*Player, error)
	GetPlayerByAccount(ctx context.Context, accountID int64) (*Player, error)
	CreatePlayer(ctx context.Context, in *Player) (*Player, error)
	UpdatePlayer(ctx context.Context, in *Player) (*Player, error)

	// Items & pets
	ListItemsByPlayer(ctx context.Context, playerID int64) ([]*Item, error)
	UpsertItem(ctx context.Context, playerID int64, itemID int, count int, meta string) (*Item, error)
	DeleteItem(ctx context.Context, playerID int64, itemID int) error
	ListPetsByPlayer(ctx context.Context, playerID int64) ([]*Pet, error)
	UpsertPet(ctx context.Context, in *Pet) (*Pet, error)

	// Configs & versions
	ListConfigKeys(ctx context.Context) ([]string, error)
	GetConfig(ctx context.Context, key string) (*ConfigEntry, error)
	SaveConfig(ctx context.Context, entry *ConfigEntry, operator string) (*ConfigVersion, error)
	ListConfigVersions(ctx context.Context, key string, limit int) ([]*ConfigVersion, error)

	// Audit logs
	ListAuditLogs(ctx context.Context, limit int) ([]*AuditLog, error)
}

type Account struct {
	ID       int64
	Email    string
	Password string
	Salt     string
	Status   string
}

type Player struct {
	ID       int64
	Account  int64
	Nick     string
	Level    int
	Coins    int64
	Gold     int64
	MapID    int
	PosX     int
	PosY     int
	MapType  int
	Color    int64
	Texture  int64
	Energy   int64
	FightBadge int64
	TimeToday int64
	TimeLimit int64
	TeacherID int64
	StudentID int64
	CurTitle  int64
	LastMapID int
	CurrentPetID        int64
	CurrentPetCatchTime int64
	CurrentPetDV        int64
	TaskStatus string
	TaskBufs   string
	Friends    string
	Blacklist  string
	TeamInfo   string
	StudentIDs string
	RoomID     int64
	Fitments   string
}

type Item struct {
	ID       int64
	PlayerID int64
	ItemID   int
	Count    int
	Meta     string
}

type Pet struct {
	ID        int64
	PlayerID  int64
	SpeciesID int
	Level     int
	Exp       int
	HP        int
	Nature    string
	Skills    string
	CatchTime int64
	DV        int
}

type ConfigEntry struct {
	Key      string
	Value    []byte
	Version  int64
	Checksum string
}

type ConfigVersion struct {
	ID        int64
	Key       string
	Version   int64
	Operator  string
	CreatedAt int64
}

type AuditLog struct {
	ID        int64
	Operator  string
	Action    string
	Resource  string
	ResourceID string
	Detail    string
	CreatedAt int64
}
