package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config defines all runtime settings for jseer.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Log      LogConfig      `mapstructure:"log"`
	Database DatabaseConfig `mapstructure:"database"`
	Login    LoginConfig    `mapstructure:"login"`
	Gateway  GatewayConfig  `mapstructure:"gateway"`
	Game     GameConfig     `mapstructure:"game"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	GM       GMConfig       `mapstructure:"gm"`
	Security SecurityConfig `mapstructure:"security"`
}

type AppConfig struct {
	Env             string `mapstructure:"env"`
	Name            string `mapstructure:"name"`
	Version         string `mapstructure:"version"`
	ReloadIntervalS int    `mapstructure:"reload_interval_s"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

type LoginConfig struct {
	Address       string `mapstructure:"address"`
	PolicyPort    int    `mapstructure:"policy_port"`
	PolicyEnabled bool   `mapstructure:"policy_enabled"`
	AdminAddress  string `mapstructure:"admin_address"`
	AdminPprof    bool   `mapstructure:"admin_pprof"`
}

type GatewayConfig struct {
	Address           string `mapstructure:"address"`
	MaxConnections    int    `mapstructure:"max_connections"`
	ReadBufferBytes   int    `mapstructure:"read_buffer_bytes"`
	WriteBufferBytes  int    `mapstructure:"write_buffer_bytes"`
	HandshakeTimeoutS int    `mapstructure:"handshake_timeout_s"`
	AdminAddress      string `mapstructure:"admin_address"`
	AdminPprof        bool   `mapstructure:"admin_pprof"`
}

type GameConfig struct {
	PublicIP   string `mapstructure:"public_ip"`
	Port       int    `mapstructure:"port"`
	ServerID   int    `mapstructure:"server_id"`
	SpawnMap   int    `mapstructure:"spawn_map"`
	SpawnX     int    `mapstructure:"spawn_x"`
	SpawnY     int    `mapstructure:"spawn_y"`
	ForceSpawn bool   `mapstructure:"force_spawn"`
}

type HTTPConfig struct {
	Address        string `mapstructure:"address"`
	LoginIPAddress string `mapstructure:"login_ip_address"`
	EnablePprof    bool   `mapstructure:"enable_pprof"`
	AllowOrigins   string `mapstructure:"allow_origins"`
	IPTxt          string `mapstructure:"ip_txt"`
	StaticRoot     string `mapstructure:"static_root"`
	ProxyRoot      string `mapstructure:"proxy_root"`
	Upstream       string `mapstructure:"upstream"`
}

type GMConfig struct {
	Address            string `mapstructure:"address"`
	JWTSecret          string `mapstructure:"jwt_secret"`
	TokenTTLMinutes    int    `mapstructure:"token_ttl_minutes"`
	RequireTwoFactor   bool   `mapstructure:"require_2fa"`
	DefaultAdminUser   string `mapstructure:"default_admin_user"`
	DefaultAdminPass   string `mapstructure:"default_admin_pass"`
	ConfigCacheSeconds int    `mapstructure:"config_cache_seconds"`
}

type SecurityConfig struct {
	AllowedIPs []string `mapstructure:"allowed_ips"`
}

// Load reads configuration from file + env vars.
func Load(path string) (*Config, error) {
	v := viper.New()
	if path != "" {
		v.SetConfigFile(path)
		v.SetConfigType("yaml")
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if path != "" {
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("read config: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

// ResolvePath returns config path overridden by env when provided.
func ResolvePath(defaultPath string) string {
	if envPath := os.Getenv("JSEER_CONFIG"); envPath != "" {
		return envPath
	}
	return defaultPath
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.env", "dev")
	v.SetDefault("app.name", "jseer")
	v.SetDefault("app.version", "0.1.0")
	v.SetDefault("app.reload_interval_s", 0)
	v.SetDefault("log.level", "info")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "file:jseer.db?_fk=1")
	v.SetDefault("login.address", ":1863")
	v.SetDefault("login.policy_port", 843)
	v.SetDefault("login.policy_enabled", true)
	v.SetDefault("login.admin_address", "")
	v.SetDefault("login.admin_pprof", false)
	v.SetDefault("gateway.address", ":5000")
	v.SetDefault("gateway.max_connections", 5000)
	v.SetDefault("gateway.read_buffer_bytes", 65536)
	v.SetDefault("gateway.write_buffer_bytes", 65536)
	v.SetDefault("gateway.handshake_timeout_s", 5)
	v.SetDefault("gateway.admin_address", "")
	v.SetDefault("gateway.admin_pprof", false)
	v.SetDefault("game.public_ip", "127.0.0.1")
	v.SetDefault("game.port", 5000)
	v.SetDefault("game.server_id", 1)
	v.SetDefault("game.spawn_map", 1)
	v.SetDefault("game.spawn_x", 300)
	v.SetDefault("game.spawn_y", 270)
	v.SetDefault("game.force_spawn", true)
	v.SetDefault("http.address", ":32400")
	v.SetDefault("http.login_ip_address", ":32401")
	v.SetDefault("http.enable_pprof", false)
	v.SetDefault("http.allow_origins", "*")
	v.SetDefault("http.ip_txt", "127.0.0.1:1863")
	v.SetDefault("http.static_root", "./resources/root")
	v.SetDefault("http.proxy_root", "./resources_proxy/root")
	v.SetDefault("http.upstream", "")
	v.SetDefault("gm.address", ":3001")
	v.SetDefault("gm.jwt_secret", "change-me")
	v.SetDefault("gm.token_ttl_minutes", 120)
	v.SetDefault("gm.require_2fa", false)
	v.SetDefault("gm.default_admin_user", "admin")
	v.SetDefault("gm.default_admin_pass", "admin")
	v.SetDefault("gm.config_cache_seconds", 5)
	v.SetDefault("security.allowed_ips", []string{})
}
