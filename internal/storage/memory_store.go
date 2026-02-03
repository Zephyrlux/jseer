package storage

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

// memoryStore is an in-memory implementation for dev/testing.
type memoryStore struct {
	mu            sync.RWMutex
	nextAccountID int64
	nextPlayerID  int64
	nextGMUserID  int64
	nextGMRoleID  int64
	nextGMPermID  int64
	accounts      map[int64]*Account
	players       map[int64]*Player
	config        map[string]*ConfigEntry
	versions      map[string][]*ConfigVersion
	items         map[int64][]*Item
	pets          map[int64][]*Pet
	audit         []*AuditLog
	gmUsers       map[int64]*GMUser
	gmRoles       map[int64]*GMRole
	gmPerms       map[int64]*GMPermission
	userRoles     map[int64][]int64
	rolePerms     map[int64][]int64
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		accounts:  make(map[int64]*Account),
		players:   make(map[int64]*Player),
		config:    make(map[string]*ConfigEntry),
		versions:  make(map[string][]*ConfigVersion),
		items:     make(map[int64][]*Item),
		pets:      make(map[int64][]*Pet),
		audit:     make([]*AuditLog, 0),
		gmUsers:   make(map[int64]*GMUser),
		gmRoles:   make(map[int64]*GMRole),
		gmPerms:   make(map[int64]*GMPermission),
		userRoles: make(map[int64][]int64),
		rolePerms: make(map[int64][]int64),
	}
}

func (s *memoryStore) Ping(ctx context.Context) error { return nil }

func (s *memoryStore) Close() error { return nil }

func (s *memoryStore) CreateAuditLog(ctx context.Context, in *AuditLog) (*AuditLog, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := *in
	copy.ID = time.Now().UnixNano()
	copy.CreatedAt = time.Now().Unix()
	s.audit = append(s.audit, &copy)
	return &copy, nil
}

func (s *memoryStore) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.accounts {
		if a.Email == email {
			copy := *a
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *memoryStore) CreateAccount(ctx context.Context, in *Account) (*Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextAccountID++
	copy := *in
	copy.ID = s.nextAccountID
	s.accounts[copy.ID] = &copy
	return &copy, nil
}

func (s *memoryStore) GetPlayerByID(ctx context.Context, id int64) (*Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.players[id]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *p
	return &copy, nil
}

func (s *memoryStore) GetPlayerByAccount(ctx context.Context, accountID int64) (*Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.players {
		if p.Account == accountID {
			copy := *p
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *memoryStore) CreatePlayer(ctx context.Context, in *Player) (*Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextPlayerID++
	copy := *in
	copy.ID = s.nextPlayerID
	s.players[copy.ID] = &copy
	return &copy, nil
}

func (s *memoryStore) UpdatePlayer(ctx context.Context, in *Player) (*Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if in == nil {
		return nil, errors.New("nil player")
	}
	_, ok := s.players[in.ID]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *in
	s.players[copy.ID] = &copy
	return &copy, nil
}

func (s *memoryStore) ListItemsByPlayer(ctx context.Context, playerID int64) ([]*Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := s.items[playerID]
	out := make([]*Item, 0, len(list))
	for _, it := range list {
		copy := *it
		out = append(out, &copy)
	}
	return out, nil
}

func (s *memoryStore) UpsertItem(ctx context.Context, playerID int64, itemID int, count int, meta string) (*Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[playerID]
	for _, it := range list {
		if it.ItemID == itemID {
			it.Count = count
			it.Meta = meta
			copy := *it
			return &copy, nil
		}
	}
	newItem := &Item{
		ID:       time.Now().UnixNano(),
		PlayerID: playerID,
		ItemID:   itemID,
		Count:    count,
		Meta:     meta,
	}
	s.items[playerID] = append(list, newItem)
	copy := *newItem
	return &copy, nil
}

func (s *memoryStore) DeleteItem(ctx context.Context, playerID int64, itemID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.items[playerID]
	if len(list) == 0 {
		return nil
	}
	next := list[:0]
	for _, it := range list {
		if it.ItemID != itemID {
			next = append(next, it)
		}
	}
	if len(next) == 0 {
		delete(s.items, playerID)
	} else {
		s.items[playerID] = next
	}
	return nil
}

func (s *memoryStore) ListPetsByPlayer(ctx context.Context, playerID int64) ([]*Pet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := s.pets[playerID]
	out := make([]*Pet, 0, len(list))
	for _, it := range list {
		copy := *it
		out = append(out, &copy)
	}
	return out, nil
}

func (s *memoryStore) UpsertPet(ctx context.Context, in *Pet) (*Pet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.pets[in.PlayerID]
	for _, it := range list {
		if it.CatchTime == in.CatchTime {
			it.SpeciesID = in.SpeciesID
			it.Level = in.Level
			it.Exp = in.Exp
			it.HP = in.HP
			it.DV = in.DV
			it.Skills = in.Skills
			it.Nature = in.Nature
			copy := *it
			return &copy, nil
		}
	}
	inCopy := *in
	inCopy.ID = time.Now().UnixNano()
	s.pets[in.PlayerID] = append(list, &inCopy)
	return &inCopy, nil
}

func (s *memoryStore) ListConfigKeys(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.config))
	for k := range s.config {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *memoryStore) GetConfig(ctx context.Context, key string) (*ConfigEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.config[key]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *entry
	return &copy, nil
}

func (s *memoryStore) SaveConfig(ctx context.Context, entry *ConfigEntry, operator string) (*ConfigVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing := s.config[entry.Key]
	ver := int64(1)
	if existing != nil {
		ver = existing.Version + 1
	}
	copy := *entry
	copy.Version = ver
	s.config[entry.Key] = &copy
	cv := &ConfigVersion{
		ID:        time.Now().UnixNano(),
		Key:       entry.Key,
		Version:   ver,
		Value:     entry.Value,
		Checksum:  entry.Checksum,
		Operator:  operator,
		CreatedAt: time.Now().Unix(),
	}
	s.versions[entry.Key] = append(s.versions[entry.Key], cv)
	s.audit = append(s.audit, &AuditLog{
		ID:         time.Now().UnixNano(),
		Operator:   operator,
		Action:     "config.save",
		Resource:   "config",
		ResourceID: entry.Key,
		Detail:     "config updated",
		CreatedAt:  time.Now().Unix(),
	})
	return cv, nil
}

func (s *memoryStore) ListConfigVersions(ctx context.Context, key string, limit int) ([]*ConfigVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := s.versions[key]
	if limit <= 0 || len(list) <= limit {
		return append([]*ConfigVersion{}, list...), nil
	}
	return append([]*ConfigVersion{}, list[len(list)-limit:]...), nil
}

func (s *memoryStore) GetConfigVersion(ctx context.Context, key string, version int64) (*ConfigVersion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := s.versions[key]
	for _, v := range list {
		if v.Version == version {
			copy := *v
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *memoryStore) RollbackConfig(ctx context.Context, key string, version int64, operator string) (*ConfigVersion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var target *ConfigVersion
	for _, v := range s.versions[key] {
		if v.Version == version {
			target = v
			break
		}
	}
	if target == nil {
		return nil, errors.New("not found")
	}
	existing := s.config[key]
	newVersion := int64(1)
	if existing != nil {
		newVersion = existing.Version + 1
	}
	entry := &ConfigEntry{
		Key:      key,
		Value:    target.Value,
		Version:  newVersion,
		Checksum: target.Checksum,
	}
	s.config[key] = entry
	cv := &ConfigVersion{
		ID:        time.Now().UnixNano(),
		Key:       key,
		Version:   newVersion,
		Value:     target.Value,
		Checksum:  target.Checksum,
		Operator:  operator,
		CreatedAt: time.Now().Unix(),
	}
	s.versions[key] = append(s.versions[key], cv)
	s.audit = append(s.audit, &AuditLog{
		ID:         time.Now().UnixNano(),
		Operator:   operator,
		Action:     "config.rollback",
		Resource:   "config",
		ResourceID: key,
		Detail:     "config rollback",
		CreatedAt:  time.Now().Unix(),
	})
	return cv, nil
}

func (s *memoryStore) ListAuditLogs(ctx context.Context, limit int) ([]*AuditLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || len(s.audit) <= limit {
		return append([]*AuditLog{}, s.audit...), nil
	}
	return append([]*AuditLog{}, s.audit[len(s.audit)-limit:]...), nil
}

func (s *memoryStore) GetGMUserByUsername(ctx context.Context, username string) (*GMUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.gmUsers {
		if u.Username == username {
			copy := *u
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *memoryStore) GetGMUserByID(ctx context.Context, id int64) (*GMUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.gmUsers[id]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *u
	return &copy, nil
}

func (s *memoryStore) ListGMUsers(ctx context.Context, filter GMUserFilter) ([]*GMUser, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*GMUser, 0, len(s.gmUsers))
	for _, u := range s.gmUsers {
		if filter.Search != "" && !strings.Contains(strings.ToLower(u.Username), strings.ToLower(filter.Search)) {
			continue
		}
		copy := *u
		out = append(out, &copy)
	}
	return applyLimitOffset(out, filter.Offset, filter.Limit), nil
}

func (s *memoryStore) CreateGMUser(ctx context.Context, in *GMUser, roleIDs []int64) (*GMUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextGMUserID++
	copy := *in
	copy.ID = s.nextGMUserID
	if copy.Status == "" {
		copy.Status = "active"
	}
	if copy.CreatedAt == 0 {
		copy.CreatedAt = time.Now().Unix()
	}
	s.gmUsers[copy.ID] = &copy
	if len(roleIDs) > 0 {
		s.userRoles[copy.ID] = append([]int64{}, roleIDs...)
	}
	return &copy, nil
}

func (s *memoryStore) UpdateGMUser(ctx context.Context, in *GMUser, roleIDs []int64) (*GMUser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.gmUsers[in.ID]
	if !ok {
		return nil, errors.New("not found")
	}
	if in.Username != "" {
		u.Username = in.Username
	}
	if in.Status != "" {
		u.Status = in.Status
	}
	if in.LastLoginAt != 0 {
		u.LastLoginAt = in.LastLoginAt
	}
	if len(roleIDs) > 0 {
		s.userRoles[in.ID] = append([]int64{}, roleIDs...)
	}
	copy := *u
	return &copy, nil
}

func (s *memoryStore) SetGMUserPassword(ctx context.Context, id int64, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.gmUsers[id]
	if !ok {
		return errors.New("not found")
	}
	u.PasswordHash = passwordHash
	return nil
}

func (s *memoryStore) SetGMUserStatus(ctx context.Context, id int64, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.gmUsers[id]
	if !ok {
		return errors.New("not found")
	}
	u.Status = status
	return nil
}

func (s *memoryStore) ListGMRolesByUser(ctx context.Context, userID int64) ([]*GMRole, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	roleIDs := s.userRoles[userID]
	out := make([]*GMRole, 0, len(roleIDs))
	for _, id := range roleIDs {
		r, ok := s.gmRoles[id]
		if ok {
			copy := *r
			out = append(out, &copy)
		}
	}
	return out, nil
}

func (s *memoryStore) SetGMUserRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.gmUsers[userID]; !ok {
		return errors.New("not found")
	}
	s.userRoles[userID] = append([]int64{}, roleIDs...)
	return nil
}

func (s *memoryStore) GetGMRoleByName(ctx context.Context, name string) (*GMRole, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.gmRoles {
		if r.Name == name {
			copy := *r
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *memoryStore) GetGMRoleByID(ctx context.Context, id int64) (*GMRole, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.gmRoles[id]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *r
	return &copy, nil
}

func (s *memoryStore) ListGMRoles(ctx context.Context, filter GMRoleFilter) ([]*GMRole, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*GMRole, 0, len(s.gmRoles))
	for _, r := range s.gmRoles {
		if filter.Search != "" && !strings.Contains(strings.ToLower(r.Name), strings.ToLower(filter.Search)) {
			continue
		}
		copy := *r
		out = append(out, &copy)
	}
	return applyLimitOffset(out, filter.Offset, filter.Limit), nil
}

func (s *memoryStore) CreateGMRole(ctx context.Context, in *GMRole, permIDs []int64) (*GMRole, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextGMRoleID++
	copy := *in
	copy.ID = s.nextGMRoleID
	if copy.CreatedAt == 0 {
		copy.CreatedAt = time.Now().Unix()
	}
	s.gmRoles[copy.ID] = &copy
	if len(permIDs) > 0 {
		s.rolePerms[copy.ID] = append([]int64{}, permIDs...)
	}
	return &copy, nil
}

func (s *memoryStore) UpdateGMRole(ctx context.Context, in *GMRole, permIDs []int64) (*GMRole, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.gmRoles[in.ID]
	if !ok {
		return nil, errors.New("not found")
	}
	if in.Name != "" {
		r.Name = in.Name
	}
	r.Description = in.Description
	if len(permIDs) > 0 {
		s.rolePerms[in.ID] = append([]int64{}, permIDs...)
	}
	copy := *r
	return &copy, nil
}

func (s *memoryStore) DeleteGMRole(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.gmRoles, id)
	delete(s.rolePerms, id)
	for userID, roles := range s.userRoles {
		next := roles[:0]
		for _, r := range roles {
			if r != id {
				next = append(next, r)
			}
		}
		s.userRoles[userID] = next
	}
	return nil
}

func (s *memoryStore) ListPermissionsByRole(ctx context.Context, roleID int64) ([]*GMPermission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	permIDs := s.rolePerms[roleID]
	out := make([]*GMPermission, 0, len(permIDs))
	for _, id := range permIDs {
		p, ok := s.gmPerms[id]
		if ok {
			copy := *p
			out = append(out, &copy)
		}
	}
	return out, nil
}

func (s *memoryStore) SetRolePermissions(ctx context.Context, roleID int64, permIDs []int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.gmRoles[roleID]; !ok {
		return errors.New("not found")
	}
	s.rolePerms[roleID] = append([]int64{}, permIDs...)
	return nil
}

func (s *memoryStore) ListGMPermissions(ctx context.Context, filter GMPermissionFilter) ([]*GMPermission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*GMPermission, 0, len(s.gmPerms))
	for _, p := range s.gmPerms {
		if filter.Search != "" && !strings.Contains(strings.ToLower(p.Code), strings.ToLower(filter.Search)) {
			continue
		}
		copy := *p
		out = append(out, &copy)
	}
	return applyLimitOffset(out, filter.Offset, filter.Limit), nil
}

func (s *memoryStore) GetGMPermissionByCode(ctx context.Context, code string) (*GMPermission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.gmPerms {
		if p.Code == code {
			copy := *p
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (s *memoryStore) CreateGMPermission(ctx context.Context, in *GMPermission) (*GMPermission, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextGMPermID++
	copy := *in
	copy.ID = s.nextGMPermID
	if copy.CreatedAt == 0 {
		copy.CreatedAt = time.Now().Unix()
	}
	s.gmPerms[copy.ID] = &copy
	return &copy, nil
}

func (s *memoryStore) UpdateGMPermission(ctx context.Context, in *GMPermission) (*GMPermission, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.gmPerms[in.ID]
	if !ok {
		return nil, errors.New("not found")
	}
	if in.Code != "" {
		p.Code = in.Code
	}
	if in.Name != "" {
		p.Name = in.Name
	}
	p.Description = in.Description
	copy := *p
	return &copy, nil
}

func (s *memoryStore) DeleteGMPermission(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.gmPerms, id)
	for roleID, perms := range s.rolePerms {
		next := perms[:0]
		for _, p := range perms {
			if p != id {
				next = append(next, p)
			}
		}
		s.rolePerms[roleID] = next
	}
	return nil
}

func applyLimitOffset[T any](items []*T, offset int, limit int) []*T {
	if offset < 0 {
		offset = 0
	}
	if offset >= len(items) {
		return []*T{}
	}
	if limit <= 0 {
		return items[offset:]
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

var _ Store = (*memoryStore)(nil)
