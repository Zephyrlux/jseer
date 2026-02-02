package storage

import (
	"fmt"
	"strings"

	"jseer/ent"

	"entgo.io/ent/dialect"

	"jseer/internal/config"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func openEntClient(cfg config.DatabaseConfig) (*ent.Client, error) {
	dsn := strings.TrimSpace(cfg.DSN)
	driver := strings.ToLower(cfg.Driver)
	var drv string
	switch driver {
	case "mysql":
		drv = dialect.MySQL
	case "sqlite", "sqlite3":
		drv = dialect.SQLite
	case "postgres", "postgresql":
		drv = dialect.Postgres
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
	if drv == dialect.SQLite {
		dsn = ensureSQLiteFK(dsn)
	}
	client, err := ent.Open(drv, dsn)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func ensureSQLiteFK(dsn string) string {
	// sqlite foreign_keys pragma must be enabled: _fk=1 (mattn/go-sqlite3).
	if dsn == "" {
		return "file:jseer.db?_fk=1"
	}
	if strings.Contains(dsn, "_fk=1") || strings.Contains(dsn, "_foreign_keys=1") {
		return dsn
	}
	if strings.Contains(dsn, "?") {
		return dsn + "&_fk=1"
	}
	return dsn + "?_fk=1"
}
