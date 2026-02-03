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
	GetConfigVersion(ctx context.Context, key string, version int64) (*ConfigVersion, error)
	RollbackConfig(ctx context.Context, key string, version int64, operator string) (*ConfigVersion, error)

	// GM users & RBAC
	GetGMUserByUsername(ctx context.Context, username string) (*GMUser, error)
	GetGMUserByID(ctx context.Context, id int64) (*GMUser, error)
	ListGMUsers(ctx context.Context, filter GMUserFilter) ([]*GMUser, error)
	CreateGMUser(ctx context.Context, in *GMUser, roleIDs []int64) (*GMUser, error)
	UpdateGMUser(ctx context.Context, in *GMUser, roleIDs []int64) (*GMUser, error)
	SetGMUserPassword(ctx context.Context, id int64, passwordHash string) error
	SetGMUserStatus(ctx context.Context, id int64, status string) error
	ListGMRolesByUser(ctx context.Context, userID int64) ([]*GMRole, error)
	SetGMUserRoles(ctx context.Context, userID int64, roleIDs []int64) error

	GetGMRoleByName(ctx context.Context, name string) (*GMRole, error)
	GetGMRoleByID(ctx context.Context, id int64) (*GMRole, error)
	ListGMRoles(ctx context.Context, filter GMRoleFilter) ([]*GMRole, error)
	CreateGMRole(ctx context.Context, in *GMRole, permIDs []int64) (*GMRole, error)
	UpdateGMRole(ctx context.Context, in *GMRole, permIDs []int64) (*GMRole, error)
	DeleteGMRole(ctx context.Context, id int64) error
	ListPermissionsByRole(ctx context.Context, roleID int64) ([]*GMPermission, error)
	SetRolePermissions(ctx context.Context, roleID int64, permIDs []int64) error

	ListGMPermissions(ctx context.Context, filter GMPermissionFilter) ([]*GMPermission, error)
	GetGMPermissionByCode(ctx context.Context, code string) (*GMPermission, error)
	CreateGMPermission(ctx context.Context, in *GMPermission) (*GMPermission, error)
	UpdateGMPermission(ctx context.Context, in *GMPermission) (*GMPermission, error)
	DeleteGMPermission(ctx context.Context, id int64) error

	// Audit logs
	CreateAuditLog(ctx context.Context, in *AuditLog) (*AuditLog, error)
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
	ID                  int64
	Account             int64
	Nick                string
	Level               int
	Coins               int64
	Gold                int64
	MapID               int
	PosX                int
	PosY                int
	MapType             int
	Color               int64
	Texture             int64
	Energy              int64
	FightBadge          int64
	TimeToday           int64
	TimeLimit           int64
	TeacherID           int64
	StudentID           int64
	CurTitle            int64
	LastMapID           int
	CurrentPetID        int64
	CurrentPetCatchTime int64
	CurrentPetDV        int64
	TaskStatus          string
	TaskBufs            string
	Friends             string
	Blacklist           string
	Achievements        string
	Titles              string
	TeamInfo            string
	StudentIDs          string
	RoomID              int64
	Fitments            string
	NonoInfo            string
	Mailbox             string
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
	ID        int64  `json:"id"`
	Key       string `json:"key"`
	Version   int64  `json:"version"`
	Value     []byte `json:"value"`
	Checksum  string `json:"checksum"`
	Operator  string `json:"operator"`
	CreatedAt int64  `json:"created_at"`
}

type AuditLog struct {
	ID         int64  `json:"id"`
	Operator   string `json:"operator"`
	Action     string `json:"action"`
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
	Detail     string `json:"detail"`
	CreatedAt  int64  `json:"created_at"`
}

type GMUser struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Status       string `json:"status"`
	LastLoginAt  int64  `json:"last_login_at"`
	CreatedAt    int64  `json:"created_at"`
}

type GMRole struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
}

type GMPermission struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"`
}

type GMUserFilter struct {
	Limit  int
	Offset int
	Search string
}

type GMRoleFilter struct {
	Limit  int
	Offset int
	Search string
}

type GMPermissionFilter struct {
	Limit  int
	Offset int
	Search string
}
