package storage

import (
	"context"

	"jseer/ent"
	"jseer/ent/account"
	"jseer/ent/auditlog"
	"jseer/ent/configentry"
	"jseer/ent/configversion"
	"jseer/ent/item"
	"jseer/ent/pet"
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
	return mapPlayer(row), nil
}

func (s *EntStore) GetPlayerByAccount(ctx context.Context, accountID int64) (*Player, error) {
	row, err := s.client.Player.Query().Where(player.AccountIDEQ(int(accountID))).Only(ctx)
	if err != nil {
		return nil, err
	}
	return mapPlayer(row), nil
}

func (s *EntStore) CreatePlayer(ctx context.Context, in *Player) (*Player, error) {
	row, err := s.client.Player.Create().
		SetAccountID(int(in.Account)).
		SetNick(in.Nick).
		SetLevel(in.Level).
		SetCoins(in.Coins).
		SetGold(in.Gold).
		SetMapID(in.MapID).
		SetMapType(in.MapType).
		SetPosX(in.PosX).
		SetPosY(in.PosY).
		SetLastMapID(in.LastMapID).
		SetColor(in.Color).
		SetTexture(in.Texture).
		SetEnergy(in.Energy).
		SetFightBadge(in.FightBadge).
		SetTimeToday(in.TimeToday).
		SetTimeLimit(in.TimeLimit).
		SetTeacherID(in.TeacherID).
		SetStudentID(in.StudentID).
		SetCurTitle(in.CurTitle).
		SetTaskStatus(normalizeJSON(in.TaskStatus)).
		SetTaskBufs(normalizeJSON(in.TaskBufs)).
		SetCurrentPetID(in.CurrentPetID).
		SetCurrentPetCatchTime(in.CurrentPetCatchTime).
		SetCurrentPetDV(in.CurrentPetDV).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapPlayer(row), nil
}

func (s *EntStore) UpdatePlayer(ctx context.Context, in *Player) (*Player, error) {
	row, err := s.client.Player.UpdateOneID(int(in.ID)).
		SetNick(in.Nick).
		SetLevel(in.Level).
		SetCoins(in.Coins).
		SetGold(in.Gold).
		SetMapID(in.MapID).
		SetMapType(in.MapType).
		SetPosX(in.PosX).
		SetPosY(in.PosY).
		SetLastMapID(in.LastMapID).
		SetColor(in.Color).
		SetTexture(in.Texture).
		SetEnergy(in.Energy).
		SetFightBadge(in.FightBadge).
		SetTimeToday(in.TimeToday).
		SetTimeLimit(in.TimeLimit).
		SetTeacherID(in.TeacherID).
		SetStudentID(in.StudentID).
		SetCurTitle(in.CurTitle).
		SetTaskStatus(normalizeJSON(in.TaskStatus)).
		SetTaskBufs(normalizeJSON(in.TaskBufs)).
		SetCurrentPetID(in.CurrentPetID).
		SetCurrentPetCatchTime(in.CurrentPetCatchTime).
		SetCurrentPetDV(in.CurrentPetDV).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapPlayer(row), nil
}

func (s *EntStore) ListItemsByPlayer(ctx context.Context, playerID int64) ([]*Item, error) {
	rows, err := s.client.Item.Query().Where(item.PlayerIDEQ(int(playerID))).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, &Item{
			ID:       int64(row.ID),
			PlayerID: int64(row.PlayerID),
			ItemID:   row.ItemID,
			Count:    row.Count,
			Meta:     row.Meta,
		})
	}
	return out, nil
}

func (s *EntStore) UpsertItem(ctx context.Context, playerID int64, itemID int, count int, meta string) (*Item, error) {
	row, err := s.client.Item.Query().
		Where(item.PlayerIDEQ(int(playerID)), item.ItemIDEQ(itemID)).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
		row, err = s.client.Item.Create().
			SetPlayerID(int(playerID)).
			SetItemID(itemID).
			SetCount(count).
			SetMeta(meta).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		row, err = row.Update().
			SetCount(count).
			SetMeta(meta).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	}
	return &Item{
		ID:       int64(row.ID),
		PlayerID: int64(row.PlayerID),
		ItemID:   row.ItemID,
		Count:    row.Count,
		Meta:     row.Meta,
	}, nil
}

func (s *EntStore) DeleteItem(ctx context.Context, playerID int64, itemID int) error {
	_, err := s.client.Item.Delete().
		Where(item.PlayerIDEQ(int(playerID)), item.ItemIDEQ(itemID)).
		Exec(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	return err
}

func (s *EntStore) ListPetsByPlayer(ctx context.Context, playerID int64) ([]*Pet, error) {
	rows, err := s.client.Pet.Query().Where(pet.PlayerIDEQ(int(playerID))).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*Pet, 0, len(rows))
	for _, row := range rows {
		out = append(out, &Pet{
			ID:        int64(row.ID),
			PlayerID:  int64(row.PlayerID),
			SpeciesID: row.SpeciesID,
			Level:     row.Level,
			Exp:       row.Exp,
			HP:        row.HP,
			Nature:    row.Nature,
			Skills:    row.Skills,
			CatchTime: row.CatchTime,
			DV:        row.Dv,
		})
	}
	return out, nil
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
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	existing, err := tx.ConfigEntry.Query().Where(configentry.KeyEQ(entry.Key)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		_ = tx.Rollback()
		return nil, err
	}

	var ver int64 = 1
	var e *ent.ConfigEntry
	if existing == nil {
		e, err = tx.ConfigEntry.Create().
			SetKey(entry.Key).
			SetValue(entry.Value).
			SetChecksum(entry.Checksum).
			SetVersion(ver).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	} else {
		ver = existing.Version + 1
		e, err = existing.Update().
			SetValue(entry.Value).
			SetChecksum(entry.Checksum).
			SetVersion(ver).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	cv, err := tx.ConfigVersion.Create().
		SetKey(entry.Key).
		SetVersion(ver).
		SetValue(entry.Value).
		SetChecksum(entry.Checksum).
		SetOperator(operator).
		SetEntry(e).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	_, _ = tx.AuditLog.Create().
		SetOperator(operator).
		SetAction("config.save").
		SetResource("config").
		SetResourceID(entry.Key).
		SetDetail("config updated").
		Save(ctx)

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &ConfigVersion{ID: int64(cv.ID), Key: cv.Key, Version: cv.Version, Operator: cv.Operator, CreatedAt: cv.CreatedAt.Unix()}, nil
}

func (s *EntStore) ListConfigVersions(ctx context.Context, key string, limit int) ([]*ConfigVersion, error) {
	query := s.client.ConfigVersion.Query().Where(configversion.KeyEQ(key)).Order(ent.Desc(configversion.FieldCreatedAt))
	if limit > 0 {
		query = query.Limit(limit)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*ConfigVersion, 0, len(rows))
	for _, row := range rows {
		out = append(out, &ConfigVersion{
			ID:        int64(row.ID),
			Key:       row.Key,
			Version:   row.Version,
			Operator:  row.Operator,
			CreatedAt: row.CreatedAt.Unix(),
		})
	}
	return out, nil
}

func (s *EntStore) ListAuditLogs(ctx context.Context, limit int) ([]*AuditLog, error) {
	query := s.client.AuditLog.Query().Order(ent.Desc(auditlog.FieldCreatedAt))
	if limit > 0 {
		query = query.Limit(limit)
	}
	rows, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*AuditLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, &AuditLog{
			ID:         int64(row.ID),
			Operator:   row.Operator,
			Action:     row.Action,
			Resource:   row.Resource,
			ResourceID: row.ResourceID,
			Detail:     row.Detail,
			CreatedAt:  row.CreatedAt.Unix(),
		})
	}
	return out, nil
}

var _ Store = (*EntStore)(nil)

func normalizeJSON(s string) string {
	if s == "" {
		return "{}"
	}
	return s
}

func mapPlayer(row *ent.Player) *Player {
	if row == nil {
		return nil
	}
	return &Player{
		ID:                  int64(row.ID),
		Account:             int64(row.AccountID),
		Nick:                row.Nick,
		Level:               row.Level,
		Coins:               row.Coins,
		Gold:                row.Gold,
		MapID:               row.MapID,
		MapType:             row.MapType,
		PosX:                row.PosX,
		PosY:                row.PosY,
		LastMapID:           row.LastMapID,
		Color:               row.Color,
		Texture:             row.Texture,
		Energy:              row.Energy,
		FightBadge:          row.FightBadge,
		TimeToday:           row.TimeToday,
		TimeLimit:           row.TimeLimit,
		TeacherID:           row.TeacherID,
		StudentID:           row.StudentID,
		CurTitle:            row.CurTitle,
		TaskStatus:          row.TaskStatus,
		TaskBufs:            row.TaskBufs,
		CurrentPetID:        row.CurrentPetID,
		CurrentPetCatchTime: row.CurrentPetCatchTime,
		CurrentPetDV:        row.CurrentPetDV,
	}
}
