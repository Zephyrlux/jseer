package storage

import (
	"context"
	"errors"
	"sync"
	"time"
)

// memoryStore is an in-memory implementation for dev/testing.
type memoryStore struct {
	mu            sync.RWMutex
	nextAccountID int64
	nextPlayerID  int64
	accounts      map[int64]*Account
	players       map[int64]*Player
	config        map[string]*ConfigEntry
	versions      map[string][]*ConfigVersion
	items         map[int64][]*Item
	pets          map[int64][]*Pet
	audit         []*AuditLog
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		accounts: make(map[int64]*Account),
		players:  make(map[int64]*Player),
		config:   make(map[string]*ConfigEntry),
		versions: make(map[string][]*ConfigVersion),
		items:    make(map[int64][]*Item),
		pets:     make(map[int64][]*Pet),
		audit:    make([]*AuditLog, 0),
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

func (s *memoryStore) ListAuditLogs(ctx context.Context, limit int) ([]*AuditLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || len(s.audit) <= limit {
		return append([]*AuditLog{}, s.audit...), nil
	}
	return append([]*AuditLog{}, s.audit[len(s.audit)-limit:]...), nil
}

var _ Store = (*memoryStore)(nil)
