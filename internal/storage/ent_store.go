package storage

import (
	"context"
	"time"

	"jseer/ent"
	"jseer/ent/account"
	"jseer/ent/auditlog"
	"jseer/ent/configentry"
	"jseer/ent/configversion"
	"jseer/ent/gmuser"
	"jseer/ent/item"
	"jseer/ent/permission"
	"jseer/ent/pet"
	"jseer/ent/player"
	"jseer/ent/role"

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
	ensureAccountSalt(in)
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
		SetFriends(normalizeJSONArray(in.Friends)).
		SetBlacklist(normalizeJSONArray(in.Blacklist)).
		SetAchievements(normalizeJSONArray(in.Achievements)).
		SetTitles(normalizeJSONArray(in.Titles)).
		SetTeamInfo(normalizeJSON(in.TeamInfo)).
		SetStudentIds(normalizeJSONArray(in.StudentIDs)).
		SetRoomID(in.RoomID).
		SetFitments(normalizeJSONArray(in.Fitments)).
		SetNonoInfo(normalizeJSON(in.NonoInfo)).
		SetMailbox(normalizeJSONArray(in.Mailbox)).
		SetCurrentPetID(in.CurrentPetID).
		SetCurrentPetCatchTime(in.CurrentPetCatchTime).
		SetCurrentPetDv(in.CurrentPetDV).
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
		SetFriends(normalizeJSONArray(in.Friends)).
		SetBlacklist(normalizeJSONArray(in.Blacklist)).
		SetAchievements(normalizeJSONArray(in.Achievements)).
		SetTitles(normalizeJSONArray(in.Titles)).
		SetTeamInfo(normalizeJSON(in.TeamInfo)).
		SetStudentIds(normalizeJSONArray(in.StudentIDs)).
		SetRoomID(in.RoomID).
		SetFitments(normalizeJSONArray(in.Fitments)).
		SetNonoInfo(normalizeJSON(in.NonoInfo)).
		SetMailbox(normalizeJSONArray(in.Mailbox)).
		SetCurrentPetID(in.CurrentPetID).
		SetCurrentPetCatchTime(in.CurrentPetCatchTime).
		SetCurrentPetDv(in.CurrentPetDV).
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
			HP:        row.Hp,
			Nature:    row.Nature,
			Skills:    row.Skills,
			CatchTime: row.CatchTime,
			DV:        row.Dv,
		})
	}
	return out, nil
}

func (s *EntStore) UpsertPet(ctx context.Context, in *Pet) (*Pet, error) {
	row, err := s.client.Pet.Query().
		Where(pet.PlayerIDEQ(int(in.PlayerID)), pet.CatchTimeEQ(in.CatchTime)).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			return nil, err
		}
		row, err = s.client.Pet.Create().
			SetPlayerID(int(in.PlayerID)).
			SetSpeciesID(in.SpeciesID).
			SetLevel(in.Level).
			SetExp(in.Exp).
			SetHp(in.HP).
			SetCatchTime(in.CatchTime).
			SetDv(in.DV).
			SetSkills(in.Skills).
			SetNature(in.Nature).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		row, err = row.Update().
			SetSpeciesID(in.SpeciesID).
			SetLevel(in.Level).
			SetExp(in.Exp).
			SetHp(in.HP).
			SetDv(in.DV).
			SetSkills(in.Skills).
			SetNature(in.Nature).
			Save(ctx)
		if err != nil {
			return nil, err
		}
	}
	return &Pet{
		ID:        int64(row.ID),
		PlayerID:  int64(row.PlayerID),
		SpeciesID: row.SpeciesID,
		Level:     row.Level,
		Exp:       row.Exp,
		HP:        row.Hp,
		Nature:    row.Nature,
		Skills:    row.Skills,
		CatchTime: row.CatchTime,
		DV:        row.Dv,
	}, nil
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
	return &ConfigVersion{
		ID:        int64(cv.ID),
		Key:       cv.Key,
		Version:   cv.Version,
		Value:     cv.Value,
		Checksum:  cv.Checksum,
		Operator:  cv.Operator,
		CreatedAt: cv.CreatedAt.Unix(),
	}, nil
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
			Value:     row.Value,
			Checksum:  row.Checksum,
			Operator:  row.Operator,
			CreatedAt: row.CreatedAt.Unix(),
		})
	}
	return out, nil
}

func (s *EntStore) GetConfigVersion(ctx context.Context, key string, version int64) (*ConfigVersion, error) {
	row, err := s.client.ConfigVersion.Query().
		Where(configversion.KeyEQ(key), configversion.VersionEQ(version)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return &ConfigVersion{
		ID:        int64(row.ID),
		Key:       row.Key,
		Version:   row.Version,
		Value:     row.Value,
		Checksum:  row.Checksum,
		Operator:  row.Operator,
		CreatedAt: row.CreatedAt.Unix(),
	}, nil
}

func (s *EntStore) RollbackConfig(ctx context.Context, key string, version int64, operator string) (*ConfigVersion, error) {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	target, err := tx.ConfigVersion.Query().
		Where(configversion.KeyEQ(key), configversion.VersionEQ(version)).
		Only(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	entry, err := tx.ConfigEntry.Query().Where(configentry.KeyEQ(key)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		_ = tx.Rollback()
		return nil, err
	}

	newVersion := int64(1)
	var e *ent.ConfigEntry
	if entry == nil {
		e, err = tx.ConfigEntry.Create().
			SetKey(key).
			SetValue(target.Value).
			SetChecksum(target.Checksum).
			SetVersion(newVersion).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	} else {
		newVersion = entry.Version + 1
		e, err = entry.Update().
			SetValue(target.Value).
			SetChecksum(target.Checksum).
			SetVersion(newVersion).
			Save(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	cv, err := tx.ConfigVersion.Create().
		SetKey(key).
		SetVersion(newVersion).
		SetValue(target.Value).
		SetChecksum(target.Checksum).
		SetOperator(operator).
		SetEntry(e).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	_, _ = tx.AuditLog.Create().
		SetOperator(operator).
		SetAction("config.rollback").
		SetResource("config").
		SetResourceID(key).
		SetDetail("config rollback").
		Save(ctx)

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &ConfigVersion{
		ID:        int64(cv.ID),
		Key:       cv.Key,
		Version:   cv.Version,
		Value:     cv.Value,
		Checksum:  cv.Checksum,
		Operator:  cv.Operator,
		CreatedAt: cv.CreatedAt.Unix(),
	}, nil
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

func (s *EntStore) CreateAuditLog(ctx context.Context, in *AuditLog) (*AuditLog, error) {
	row, err := s.client.AuditLog.Create().
		SetOperator(in.Operator).
		SetAction(in.Action).
		SetResource(in.Resource).
		SetResourceID(in.ResourceID).
		SetDetail(in.Detail).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &AuditLog{
		ID:         int64(row.ID),
		Operator:   row.Operator,
		Action:     row.Action,
		Resource:   row.Resource,
		ResourceID: row.ResourceID,
		Detail:     row.Detail,
		CreatedAt:  row.CreatedAt.Unix(),
	}, nil
}

func (s *EntStore) GetGMUserByUsername(ctx context.Context, username string) (*GMUser, error) {
	row, err := s.client.GMUser.Query().Where(gmuser.UsernameEQ(username)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return mapGMUser(row), nil
}

func (s *EntStore) GetGMUserByID(ctx context.Context, id int64) (*GMUser, error) {
	row, err := s.client.GMUser.Get(ctx, int(id))
	if err != nil {
		return nil, err
	}
	return mapGMUser(row), nil
}

func (s *EntStore) ListGMUsers(ctx context.Context, filter GMUserFilter) ([]*GMUser, error) {
	query := s.client.GMUser.Query()
	if filter.Search != "" {
		query = query.Where(gmuser.UsernameContainsFold(filter.Search))
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	rows, err := query.Order(ent.Desc(gmuser.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*GMUser, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapGMUser(row))
	}
	return out, nil
}

func (s *EntStore) CreateGMUser(ctx context.Context, in *GMUser, roleIDs []int64) (*GMUser, error) {
	builder := s.client.GMUser.Create().
		SetUsername(in.Username).
		SetPasswordHash(in.PasswordHash).
		SetStatus(in.Status)
	if in.LastLoginAt > 0 {
		builder = builder.SetLastLoginAt(time.Unix(in.LastLoginAt, 0))
	}
	if len(roleIDs) > 0 {
		ids := make([]int, 0, len(roleIDs))
		for _, id := range roleIDs {
			ids = append(ids, int(id))
		}
		builder = builder.AddRoleIDs(ids...)
	}
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapGMUser(row), nil
}

func (s *EntStore) UpdateGMUser(ctx context.Context, in *GMUser, roleIDs []int64) (*GMUser, error) {
	builder := s.client.GMUser.UpdateOneID(int(in.ID))
	if in.Username != "" {
		builder = builder.SetUsername(in.Username)
	}
	if in.Status != "" {
		builder = builder.SetStatus(in.Status)
	}
	if in.LastLoginAt > 0 {
		builder = builder.SetLastLoginAt(time.Unix(in.LastLoginAt, 0))
	}
	if len(roleIDs) > 0 {
		ids := make([]int, 0, len(roleIDs))
		for _, id := range roleIDs {
			ids = append(ids, int(id))
		}
		builder = builder.ClearRoles().AddRoleIDs(ids...)
	}
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapGMUser(row), nil
}

func (s *EntStore) SetGMUserPassword(ctx context.Context, id int64, passwordHash string) error {
	_, err := s.client.GMUser.UpdateOneID(int(id)).SetPasswordHash(passwordHash).Save(ctx)
	return err
}

func (s *EntStore) SetGMUserStatus(ctx context.Context, id int64, status string) error {
	_, err := s.client.GMUser.UpdateOneID(int(id)).SetStatus(status).Save(ctx)
	return err
}

func (s *EntStore) ListGMRolesByUser(ctx context.Context, userID int64) ([]*GMRole, error) {
	user, err := s.client.GMUser.Query().Where(gmuser.IDEQ(int(userID))).WithRoles().Only(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*GMRole, 0, len(user.Edges.Roles))
	for _, row := range user.Edges.Roles {
		out = append(out, mapGMRole(row))
	}
	return out, nil
}

func (s *EntStore) SetGMUserRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	ids := make([]int, 0, len(roleIDs))
	for _, id := range roleIDs {
		ids = append(ids, int(id))
	}
	_, err := s.client.GMUser.UpdateOneID(int(userID)).ClearRoles().AddRoleIDs(ids...).Save(ctx)
	return err
}

func (s *EntStore) GetGMRoleByName(ctx context.Context, name string) (*GMRole, error) {
	row, err := s.client.Role.Query().Where(role.NameEQ(name)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return mapGMRole(row), nil
}

func (s *EntStore) GetGMRoleByID(ctx context.Context, id int64) (*GMRole, error) {
	row, err := s.client.Role.Get(ctx, int(id))
	if err != nil {
		return nil, err
	}
	return mapGMRole(row), nil
}

func (s *EntStore) ListGMRoles(ctx context.Context, filter GMRoleFilter) ([]*GMRole, error) {
	query := s.client.Role.Query()
	if filter.Search != "" {
		query = query.Where(role.NameContainsFold(filter.Search))
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	rows, err := query.Order(ent.Desc(role.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*GMRole, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapGMRole(row))
	}
	return out, nil
}

func (s *EntStore) CreateGMRole(ctx context.Context, in *GMRole, permIDs []int64) (*GMRole, error) {
	builder := s.client.Role.Create().
		SetName(in.Name).
		SetDescription(in.Description)
	if len(permIDs) > 0 {
		ids := make([]int, 0, len(permIDs))
		for _, id := range permIDs {
			ids = append(ids, int(id))
		}
		builder = builder.AddPermissionIDs(ids...)
	}
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapGMRole(row), nil
}

func (s *EntStore) UpdateGMRole(ctx context.Context, in *GMRole, permIDs []int64) (*GMRole, error) {
	builder := s.client.Role.UpdateOneID(int(in.ID)).
		SetDescription(in.Description)
	if in.Name != "" {
		builder = builder.SetName(in.Name)
	}
	if len(permIDs) > 0 {
		ids := make([]int, 0, len(permIDs))
		for _, id := range permIDs {
			ids = append(ids, int(id))
		}
		builder = builder.ClearPermissions().AddPermissionIDs(ids...)
	}
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapGMRole(row), nil
}

func (s *EntStore) DeleteGMRole(ctx context.Context, id int64) error {
	return s.client.Role.DeleteOneID(int(id)).Exec(ctx)
}

func (s *EntStore) ListPermissionsByRole(ctx context.Context, roleID int64) ([]*GMPermission, error) {
	row, err := s.client.Role.Query().Where(role.IDEQ(int(roleID))).WithPermissions().Only(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*GMPermission, 0, len(row.Edges.Permissions))
	for _, perm := range row.Edges.Permissions {
		out = append(out, mapGMPermission(perm))
	}
	return out, nil
}

func (s *EntStore) SetRolePermissions(ctx context.Context, roleID int64, permIDs []int64) error {
	ids := make([]int, 0, len(permIDs))
	for _, id := range permIDs {
		ids = append(ids, int(id))
	}
	_, err := s.client.Role.UpdateOneID(int(roleID)).ClearPermissions().AddPermissionIDs(ids...).Save(ctx)
	return err
}

func (s *EntStore) ListGMPermissions(ctx context.Context, filter GMPermissionFilter) ([]*GMPermission, error) {
	query := s.client.Permission.Query()
	if filter.Search != "" {
		query = query.Where(permission.CodeContainsFold(filter.Search))
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	rows, err := query.Order(ent.Desc(permission.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*GMPermission, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapGMPermission(row))
	}
	return out, nil
}

func (s *EntStore) GetGMPermissionByCode(ctx context.Context, code string) (*GMPermission, error) {
	row, err := s.client.Permission.Query().Where(permission.CodeEQ(code)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return mapGMPermission(row), nil
}

func (s *EntStore) CreateGMPermission(ctx context.Context, in *GMPermission) (*GMPermission, error) {
	row, err := s.client.Permission.Create().
		SetCode(in.Code).
		SetName(in.Name).
		SetDescription(in.Description).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapGMPermission(row), nil
}

func (s *EntStore) UpdateGMPermission(ctx context.Context, in *GMPermission) (*GMPermission, error) {
	builder := s.client.Permission.UpdateOneID(int(in.ID)).
		SetDescription(in.Description)
	if in.Code != "" {
		builder = builder.SetCode(in.Code)
	}
	if in.Name != "" {
		builder = builder.SetName(in.Name)
	}
	row, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return mapGMPermission(row), nil
}

func (s *EntStore) DeleteGMPermission(ctx context.Context, id int64) error {
	return s.client.Permission.DeleteOneID(int(id)).Exec(ctx)
}

var _ Store = (*EntStore)(nil)

func normalizeJSON(s string) string {
	if s == "" {
		return "{}"
	}
	return s
}

func normalizeJSONArray(s string) string {
	if s == "" {
		return "[]"
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
		Friends:             row.Friends,
		Blacklist:           row.Blacklist,
		Achievements:        row.Achievements,
		Titles:              row.Titles,
		TeamInfo:            row.TeamInfo,
		StudentIDs:          row.StudentIds,
		RoomID:              row.RoomID,
		Fitments:            row.Fitments,
		NonoInfo:            row.NonoInfo,
		Mailbox:             row.Mailbox,
		CurrentPetID:        row.CurrentPetID,
		CurrentPetCatchTime: row.CurrentPetCatchTime,
		CurrentPetDV:        row.CurrentPetDv,
	}
}

func mapGMUser(row *ent.GMUser) *GMUser {
	if row == nil {
		return nil
	}
	lastLogin := int64(0)
	if !row.LastLoginAt.IsZero() {
		lastLogin = row.LastLoginAt.Unix()
	}
	return &GMUser{
		ID:           int64(row.ID),
		Username:     row.Username,
		PasswordHash: row.PasswordHash,
		Status:       row.Status,
		LastLoginAt:  lastLogin,
		CreatedAt:    row.CreatedAt.Unix(),
	}
}

func mapGMRole(row *ent.Role) *GMRole {
	if row == nil {
		return nil
	}
	return &GMRole{
		ID:          int64(row.ID),
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt.Unix(),
	}
}

func mapGMPermission(row *ent.Permission) *GMPermission {
	if row == nil {
		return nil
	}
	return &GMPermission{
		ID:          int64(row.ID),
		Code:        row.Code,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt.Unix(),
	}
}
