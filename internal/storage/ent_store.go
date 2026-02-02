package storage

import (
	"context"
	"errors"

	"jseer/ent"
	"jseer/ent/account"
	"jseer/ent/configentry"
	"jseer/ent/configversion"
	"jseer/ent/player"

	"jseer/internal/config"
)

type EntStore struct {
	client *ent.Client
}

func newEntStore(cfg config.DatabaseConfig) (Store, error) {
	client, err := openEntClient(cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &EntStore{client: client}, nil
}

func (s *EntStore) Ping(ctx context.Context) error {
	return nil
}

func (s *EntStore) Close() error {
	return s.client.Close()
}

func (s *EntStore) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	row, err := s.client.Account.Query().Where(account.EmailEQ(email)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return &Account{ID: int64(row.ID), Email: row.Email, Password: row.PasswordHash, Salt: row.Salt, Status: row.Status}, nil
}

func (s *EntStore) CreateAccount(ctx context.Context, in *Account) (*Account, error) {
	row, err := s.client.Account.Create().SetEmail(in.Email).SetPasswordHash(in.Password).SetSalt(in.Salt).SetStatus(in.Status).Save(ctx)
	if err != nil {
		return nil, err
	}
	return &Account{ID: int64(row.ID), Email: row.Email, Password: row.PasswordHash, Salt: row.Salt, Status: row.Status}, nil
}

func (s *EntStore) GetPlayerByID(ctx context.Context, id int64) (*Player, error) {
	row, err := s.client.Player.Get(ctx, int(id))
	if err != nil {
		return nil, err
	}
	return &Player{ID: int64(row.ID), Account: int64(row.AccountID), Nick: row.Nick, Level: row.Level, Coins: row.Coins, Gold: row.Gold, MapID: row.MapID, PosX: row.PosX, PosY: row.PosY}, nil
}

func (s *EntStore) GetPlayerByAccount(ctx context.Context, accountID int64) (*Player, error) {
	row, err := s.client.Player.Query().Where(player.AccountIDEQ(int(accountID))).Only(ctx)
	if err != nil {
		return nil, err
	}
	return &Player{ID: int64(row.ID), Account: int64(row.AccountID), Nick: row.Nick, Level: row.Level, Coins: row.Coins, Gold: row.Gold, MapID: row.MapID, PosX: row.PosX, PosY: row.PosY}, nil
}

func (s *EntStore) CreatePlayer(ctx context.Context, in *Player) (*Player, error) {
	row, err := s.client.Player.Create().SetAccountID(int(in.Account)).SetNick(in.Nick).SetLevel(in.Level).SetCoins(in.Coins).SetGold(in.Gold).SetMapID(in.MapID).SetPosX(in.PosX).SetPosY(in.PosY).Save(ctx)
	if err != nil {
		return nil, err
	}
	return &Player{ID: int64(row.ID), Account: int64(row.AccountID), Nick: row.Nick, Level: row.Level, Coins: row.Coins, Gold: row.Gold, MapID: row.MapID, PosX: row.PosX, PosY: row.PosY}, nil
}

func (s *EntStore) ListConfigKeys(ctx context.Context) ([]string, error) {
	return s.client.ConfigEntry.Query().Select(configentry.FieldKey).Strings(ctx)
}

func (s *EntStore) GetConfig(ctx context.Context, key string) (*ConfigEntry, error) {
	row, err := s.client.ConfigEntry.Query().Where(configentry.KeyEQ(key)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return &ConfigEntry{Key: row.Key, Value: row.Value, Version: row.Version, Checksum: row.Checksum}, nil
}

func (s *EntStore) SaveConfig(ctx context.Context, entry *ConfigEntry, operator string) (*ConfigVersion, error) {
	_ = configversion.FieldKey
	return nil, errors.New("SaveConfig not implemented")
}

func (s *EntStore) ListConfigVersions(ctx context.Context, key string, limit int) ([]*ConfigVersion, error) {
	return nil, errors.New("ListConfigVersions not implemented")
}

var _ Store = (*EntStore)(nil)
