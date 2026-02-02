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
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		accounts: make(map[int64]*Account),
		players:  make(map[int64]*Player),
		config:   make(map[string]*ConfigEntry),
		versions: make(map[string][]*ConfigVersion),
	}
}

func (s *memoryStore) Ping(ctx context.Context) error { return nil }

func (s *memoryStore) Close() error { return nil }

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

var _ Store = (*memoryStore)(nil)
