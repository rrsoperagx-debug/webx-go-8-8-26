
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	DB       DBConfig       `mapstructure:"db"`
	Security SecurityConfig `mapstructure:"security"`
	Features FeaturesConfig `mapstructure:"features"`
	OTLP     OTLPConfig     `mapstructure:"otlp"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Env     string `mapstructure:"env"`
}

type ServerConfig struct {
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	TimeoutSecs      int    `mapstructure:"timeout_secs"`
	BodyLimitMB      int    `mapstructure:"body_limit_mb"`
	ConcurrencyLimit int    `mapstructure:"concurrency_limit"`
	ReadTimeoutSecs  int    `mapstructure:"read_timeout_secs"`
	WriteTimeoutSecs int    `mapstructure:"write_timeout_secs"`
}

type DBConfig struct {
	Path           string `mapstructure:"path"`
	MaxConnections int    `mapstructure:"max_connections"`
	Key            string `mapstructure:"key"`
}

type SecurityConfig struct {
	CorsOrigin     string `mapstructure:"cors_origin"`
	JWTSecret      string `mapstructure:"jwt_secret"`
	JWTExpiryHours int    `mapstructure:"jwt_expiry_hours"`
}

type FeaturesConfig struct {
	UseVacuumV2 bool `mapstructure:"use_vacuum_v2"`
	EnablePprof bool `mapstructure:"enable_pprof"`
}

type OTLPConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Enabled  bool   `mapstructure:"enabled"`
}

type MetricsConfig struct {
	Namespace string `mapstructure:"namespace"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigName("default")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath("./webx-metrics-pro-go/config")
	v.AddConfigPath("../config")

	// defaults
	v.SetDefault("app.name", "webx-metrics-pro-go")
	v.SetDefault("app.version", "2.2.1")
	v.SetDefault("app.env", "development")
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.timeout_secs", 5)
	v.SetDefault("server.body_limit_mb", 5)
	v.SetDefault("server.concurrency_limit", 512)
	v.SetDefault("server.read_timeout_secs", 10)
	v.SetDefault("server.write_timeout_secs", 10)
	v.SetDefault("db.path", "./data/metrics.db")
	v.SetDefault("db.max_connections", 10)
	v.SetDefault("security.cors_origin", "*")
	v.SetDefault("security.jwt_secret", "change_me_in_prod_32_chars_minimum_go")
	v.SetDefault("security.jwt_expiry_hours", 24)
	v.SetDefault("features.use_vacuum_v2", true)

	_ = v.ReadInConfig()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("WEBX_APP_ENV")
	}
	if env != "" && env != "development" {
		v.SetConfigName(env)
		_ = v.MergeInConfig()
	}

	v.AutomaticEnv()
	v.SetEnvPrefix("WEBX")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__", "-", "_"))

	// direct overrides for 12-factor
	if port := os.Getenv("PORT"); port != "" {
		v.Set("server.port", port)
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		v.Set("db.path", dbPath)
	}
	if jwt := os.Getenv("JWT_SECRET"); jwt != "" {
		v.Set("security.jwt_secret", jwt)
	}
	if cors := os.Getenv("CORS_ORIGIN"); cors != "" {
		v.Set("security.cors_origin", cors)
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		// parse sqlite://./data/metrics.db?key=xxx
		if strings.HasPrefix(dbURL, "sqlite://") {
			parts := strings.Split(strings.TrimPrefix(dbURL, "sqlite://"), "?")
			if parts[0] != "" {
				v.Set("db.path", parts[0])
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// ensure data dir
	if cfg.DB.Path != "" {
		if dir := pathlibDir(cfg.DB.Path); dir != "" {
			_ = os.MkdirAll(dir, 0755)
		}
	}

	return &cfg, nil
}

func pathlibDir(p string) string {
	if idx := strings.LastIndex(p, "/"); idx != -1 {
		return p[:idx]
	}
	return ""
}

func (c *Config) IsDev() bool {
	return c.App.Env != "production"
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
