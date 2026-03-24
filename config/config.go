package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Environment string           `mapstructure:"environment"`
	BaseURL     string           `mapstructure:"base_url"`
	FrontendURL string           `mapstructure:"frontend_url"`
	Server      ServerConfig     `mapstructure:"server"`
	Database    DatabaseConfig   `mapstructure:"database"`
	Redis       RedisConfig      `mapstructure:"redis"`
	JWT         JWTConfig        `mapstructure:"jwt"`
	Logger      LoggerConfig     `mapstructure:"logger"`
	CORS        CORSConfig       `mapstructure:"cors"`
	Storage     StorageConfig    `mapstructure:"storage"`
	OAuth       OAuthConfig      `mapstructure:"oauth"`
	Moderation  ModerationConfig `mapstructure:"moderation"`
	Telegram    TelegramConfig   `mapstructure:"telegram"`
	Admin       AdminConfig      `mapstructure:"admin"`
	Escrow       EscrowConfig       `mapstructure:"escrow"`
	Payment      PaymentConfig      `mapstructure:"payment"`
	WebPush      WebPushConfig      `mapstructure:"webpush"`
	SMS          SMSConfig          `mapstructure:"sms"`
	WelcomeBonus WelcomeBonusConfig `mapstructure:"welcome_bonus"`
	Fiscal       FiscalConfig       `mapstructure:"fiscal"`
}

type FiscalConfig struct {
	Provider      string `mapstructure:"provider"`       // "none" or "atol"
	ATOLLogin     string `mapstructure:"atol_login"`
	ATOLPassword  string `mapstructure:"atol_password"`
	ATOLGroupCode string `mapstructure:"atol_group_code"`
}

type SMSConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
}

type WelcomeBonusConfig struct {
	Amount    int64 `mapstructure:"amount"`     // in kopecks, default 50000 (500 RUB)
	ExpiryDays int  `mapstructure:"expiry_days"` // default 30
}

type WebPushConfig struct {
	VAPIDPublicKey  string `mapstructure:"vapid_public_key"`
	VAPIDPrivateKey string `mapstructure:"vapid_private_key"`
	VAPIDContact    string `mapstructure:"vapid_contact"`
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

type ModerationConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	AutoApprove bool `mapstructure:"auto_approve"`
}

type TelegramConfig struct {
	BotToken   string `mapstructure:"bot_token"`
	WebhookURL string `mapstructure:"webhook_url"`
	Mode       string `mapstructure:"mode"` // "polling" or "webhook"
}

type AdminConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Prefix   string `mapstructure:"prefix"`
	Language string `mapstructure:"language"`
	Theme    string `mapstructure:"theme"`
}

type EscrowConfig struct {
	ClaimHours int `mapstructure:"claim_hours"` // hours to hold funds before release (24-168, default 48)
}

type PaymentConfig struct {
	YooKassa               YooKassaConfig `mapstructure:"yookassa"`
	ReturnURL              string         `mapstructure:"return_url"`
	WalletRefundBonusPercent int          `mapstructure:"wallet_refund_bonus_percent"` // bonus % for wallet refund (0-15, default 5)
}

type YooKassaConfig struct {
	ShopID    string `mapstructure:"shop_id"`
	SecretKey string `mapstructure:"secret_key"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
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
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
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
	// Load .env file if it exists (ignored in production where real env vars are used)
	_ = godotenv.Load()

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
	v.SetDefault("database.max_conns", 20)
	v.SetDefault("database.min_conns", 2)
	v.SetDefault("database.max_conn_lifetime", "1h")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.token_ttl", "24h")
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("storage.endpoint", "localhost:9000")
	v.SetDefault("storage.bucket", "bani-avatars")
	v.SetDefault("storage.access_key", "minioadmin")
	v.SetDefault("storage.secret_key", "minioadmin")
	v.SetDefault("storage.region", "us-east-1")
	v.SetDefault("storage.use_ssl", false)
	v.SetDefault("moderation.enabled", true)
	v.SetDefault("moderation.auto_approve", false)
	v.SetDefault("telegram.mode", "polling")
	v.SetDefault("admin.enabled", false)
	v.SetDefault("admin.prefix", "/admin-panel")
	v.SetDefault("admin.language", "ru")
	v.SetDefault("admin.theme", "adminlte")
	v.SetDefault("cors.allowed_origins", []string{"*"})
	v.SetDefault("payment.yookassa.shop_id", "")
	v.SetDefault("payment.yookassa.secret_key", "")
	v.SetDefault("payment.return_url", "http://localhost:3000/payment/callback")
	v.SetDefault("sms.provider", "smsru")
	v.SetDefault("sms.api_key", "")
	v.SetDefault("welcome_bonus.amount", 50000)      // 500 RUB in kopecks
	v.SetDefault("welcome_bonus.expiry_days", 30)
	v.SetDefault("escrow.claim_hours", 48)
	v.SetDefault("payment.wallet_refund_bonus_percent", 5)
	v.SetDefault("fiscal.provider", "none")
	v.SetDefault("fiscal.atol_login", "")
	v.SetDefault("fiscal.atol_password", "")
	v.SetDefault("fiscal.atol_group_code", "")
	v.SetDefault("frontend_url", "http://localhost:3000")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Viper doesn't split comma-separated env vars into slices automatically.
	// Handle BANI_CORS_ALLOWED_ORIGINS="origin1,origin2" by splitting manually.
	if len(cfg.CORS.AllowedOrigins) == 1 && strings.Contains(cfg.CORS.AllowedOrigins[0], ",") {
		parts := strings.Split(cfg.CORS.AllowedOrigins[0], ",")
		cfg.CORS.AllowedOrigins = make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				cfg.CORS.AllowedOrigins = append(cfg.CORS.AllowedOrigins, trimmed)
			}
		}
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

	// Validate pool settings
	if c.Database.MaxConns > 100 {
		return fmt.Errorf("database.max_conns=%d exceeds recommended maximum of 100 (set BANI_DATABASE_MAX_CONNS)", c.Database.MaxConns)
	}
	if c.Database.MaxConns > 0 && c.Database.MinConns > c.Database.MaxConns {
		return fmt.Errorf("database.min_conns (%d) must not exceed database.max_conns (%d)", c.Database.MinConns, c.Database.MaxConns)
	}

	// Production-specific validation
	if strings.EqualFold(c.Environment, "production") {
		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("jwt.secret must be at least 32 characters long in production (current length: %d)", len(c.JWT.Secret))
		}
		if strings.Contains(c.Database.DSN, "sslmode=disable") ||
			strings.Contains(c.Database.DSN, "sslmode=prefer") ||
			strings.Contains(c.Database.DSN, "sslmode=allow") ||
			!strings.Contains(c.Database.DSN, "sslmode=") {
			return fmt.Errorf("database.dsn must use sslmode=require, sslmode=verify-ca, or sslmode=verify-full in production")
		}
		if c.Storage.AccessKey == "" || c.Storage.AccessKey == "minioadmin" ||
			c.Storage.SecretKey == "" || c.Storage.SecretKey == "minioadmin" {
			return fmt.Errorf("storage credentials must be configured in production (set BANI_STORAGE_ACCESS_KEY and BANI_STORAGE_SECRET_KEY)")
		}
		if !c.Storage.UseSSL {
			return fmt.Errorf("storage.use_ssl must be true in production")
		}
		for _, origin := range c.CORS.AllowedOrigins {
			if origin == "*" {
				return fmt.Errorf("cors.allowed_origins must not contain wildcard '*' in production (set BANI_CORS_ALLOWED_ORIGINS)")
			}
		}
	}

	return nil
}

// Warnings returns non-fatal warnings about the configuration.
// These should be logged at startup but do not prevent the application from running.
func (c *Config) Warnings() []string {
	var warnings []string
	if c.Payment.YooKassa.ShopID == "" {
		warnings = append(warnings, "payment.yookassa.shop_id is empty: online payments will be disabled (set BANI_PAYMENT_YOOKASSA_SHOP_ID)")
	}
	if c.Telegram.BotToken == "" {
		warnings = append(warnings, "telegram.bot_token is empty: telegram notifications will be disabled (set BANI_TELEGRAM_BOT_TOKEN)")
	}
	return warnings
}
