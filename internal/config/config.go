package config

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Plaid    PlaidConfig
	AI       AIConfig
	Email    EmailConfig
	Features FeatureFlags
}

type ServerConfig struct {
	Host        string        `env:"SERVER_HOST,default=0.0.0.0"`
	Port        string        `env:"SERVER_PORT,default=8080"`
	Environment string        `env:"ENVIRONMENT,default=development"`
	ReadTimeout time.Duration `env:"SERVER_READ_TIMEOUT,default=30s"`
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

func (s ServerConfig) IsProd() bool {
	return s.Environment == "production"
}

type DatabaseConfig struct {
	URL             string        `env:"DATABASE_URL,required"`
	MaxConns        int32         `env:"DB_MAX_CONNS,default=10"`
	MinConns        int32         `env:"DB_MIN_CONNS,default=2"`
	MaxConnLifetime time.Duration `env:"DB_MAX_CONN_LIFETIME,default=1h"`
	MaxConnIdleTime time.Duration `env:"DB_MAX_CONN_IDLE_TIME,default=15m"`
}

type AuthConfig struct {
	JWTSecret         string        `env:"JWT_SECRET,required"`
	EncryptionKey     string        `env:"ENCRYPTION_KEY,required"`
	AccessTokenTTL    time.Duration `env:"ACCESS_TOKEN_TTL,default=15m"`
	RefreshTokenTTL   time.Duration `env:"REFRESH_TOKEN_TTL,default=168h"` // 7 days
	MaxLoginAttempts  int           `env:"MAX_LOGIN_ATTEMPTS,default=5"`
	LockoutDuration   time.Duration `env:"LOCKOUT_DURATION,default=15m"`
	MaxTOTPAttempts   int           `env:"MAX_TOTP_ATTEMPTS,default=5"`
}

type PlaidConfig struct {
	ClientID string `env:"PLAID_CLIENT_ID"`
	Secret   string `env:"PLAID_SECRET"`
	Env      string `env:"PLAID_ENV,default=sandbox"`
}

type AIConfig struct {
	ServiceURL string `env:"AI_SERVICE_URL,default=http://localhost:8081"`
	ServiceKey string `env:"AI_SERVICE_KEY"`
	HMACSecret string `env:"AI_HMAC_SECRET"`
	GeminiKey  string `env:"GEMINI_API_KEY"`
}

type EmailConfig struct {
	ResendAPIKey string `env:"RESEND_API_KEY"`
	FromAddress  string `env:"EMAIL_FROM,default=noreply@finai.app"`
}

type FeatureFlags struct {
	AIChat          bool `env:"FF_AI_CHAT,default=true"`
	BillNegotiation bool `env:"FF_BILL_NEGOTIATION,default=true"`
	PWA             bool `env:"FF_PWA,default=true"`
	WeeklySummary   bool `env:"FF_WEEKLY_SUMMARY,default=true"`
	CreditScore     bool `env:"FF_CREDIT_SCORE,default=false"`
	DemoMode        bool `env:"DEMO_MODE,default=false"`
}

func Load(ctx context.Context) (*Config, error) {
	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters (got %d)", len(c.Auth.JWTSecret))
	}
	if len(c.Auth.EncryptionKey) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes (got %d)", len(c.Auth.EncryptionKey))
	}
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Server.IsProd() {
		if c.AI.ServiceKey == "" {
			return fmt.Errorf("AI_SERVICE_KEY is required in production")
		}
		if c.AI.HMACSecret == "" {
			return fmt.Errorf("AI_HMAC_SECRET is required in production")
		}
		if c.Features.AIChat && c.AI.GeminiKey == "" {
			return fmt.Errorf("GEMINI_API_KEY is required in production when AI chat is enabled")
		}
	}
	return nil
}
