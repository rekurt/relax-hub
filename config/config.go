package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	Redis       RedisConfig    `mapstructure:"redis"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	Logger      LoggerConfig   `mapstructure:"logger"`
	Storage     StorageConfig  `mapstructure:"storage"`
	OAuth       OAuthConfig    `mapstructure:"oauth"`
}

type OAuthConfig struct {
	VK     OAuthProviderConfig `mapstructure:"vk"`
	Yandex OAuthProviderConfig `mapstructure:"yandex"`
	Google OAuthProviderConfig `mapstructure:"google"`
}

type OAuthProviderConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret   string        `mapstructure:"secret"`
	TokenTTL time.Duration `mapstructure:"token_ttl"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

type StorageConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	Bucket    string `mapstructure:"bucket"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Region    string `mapstructure:"region"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	v.SetEnvPrefix("BANI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("environment", "dev")
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("database.dsn", "postgres://postgres:postgres@localhost:5432/bani?sslmode=disable")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.token_ttl", "24h")
	v.SetDefault("logger.level", "info")
	v.SetDefault("storage.endpoint", "localhost:9000")
	v.SetDefault("storage.bucket", "bani-avatars")
	v.SetDefault("storage.access_key", "minioadmin")
	v.SetDefault("storage.secret_key", "minioadmin")
	v.SetDefault("storage.region", "us-east-1")
	v.SetDefault("storage.use_ssl", false)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks configuration for required fields and production-specific constraints
func (c *Config) Validate() error {
	// Validate DSN is set
	if c.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required (set BANI_DATABASE_DSN)")
	}

	// Validate JWT Secret
	if c.JWT.Secret == "" || c.JWT.Secret == "change-me-in-production" {
		return fmt.Errorf("jwt.secret must be configured with a secure value (set BANI_JWT_SECRET)")
	}

	// Validate Redis Addr
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis.addr is required (set BANI_REDIS_ADDR)")
	}

	// Production-specific validation
	if c.Environment == "production" {
		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("jwt.secret must be at least 32 characters long in production (current length: %d)", len(c.JWT.Secret))
		}
		if !strings.Contains(c.Database.DSN, "sslmode=require") {
			return fmt.Errorf("database.dsn must use sslmode=require in production")
		}
		if c.Storage.AccessKey == "" || c.Storage.AccessKey == "minioadmin" ||
			c.Storage.SecretKey == "" || c.Storage.SecretKey == "minioadmin" {
			return fmt.Errorf("storage credentials must be configured in production (set BANI_STORAGE_ACCESS_KEY and BANI_STORAGE_SECRET_KEY)")
		}
		if !c.Storage.UseSSL {
			return fmt.Errorf("storage.use_ssl must be true in production")
		}
	}

	return nil
}
