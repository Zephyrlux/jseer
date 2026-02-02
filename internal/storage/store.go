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

	// Configs & versions
	ListConfigKeys(ctx context.Context) ([]string, error)
	GetConfig(ctx context.Context, key string) (*ConfigEntry, error)
	SaveConfig(ctx context.Context, entry *ConfigEntry, operator string) (*ConfigVersion, error)
	ListConfigVersions(ctx context.Context, key string, limit int) ([]*ConfigVersion, error)
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
