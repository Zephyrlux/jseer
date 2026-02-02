package storage

import (
	"errors"
	"strings"

	"jseer/internal/config"
)

// NewStore builds a store implementation based on config.
func NewStore(cfg config.DatabaseConfig) (Store, error) {
	switch cfg.Driver {
	case "mysql", "sqlite", "postgres":
		return newEntStore(cfg)
	case "ent-mysql", "ent-sqlite", "ent-postgres":
		return newEntStore(config.DatabaseConfig{Driver: strings.TrimPrefix(cfg.Driver, "ent-"), DSN: cfg.DSN})
	default:
		return nil, errors.New("unsupported database driver (ent required)")
	}
}
