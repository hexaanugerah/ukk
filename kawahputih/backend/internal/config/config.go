package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	SMTP        SMTPConfig
	GoogleOAuth GoogleOAuthConfig
	Midtrans    MidtransConfig
}

type AppConfig struct {
	Env         string
	Port        string
	FrontendURL string
}

type DatabaseConfig struct {
	Host, Port, User, Password, Name string
	MaxOpenConns, MaxIdleConns       int
	ConnMaxLifetime                  time.Duration
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FJakarta",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

type RedisConfig struct {
	Addr, Password string
	DB             int
}

type JWTConfig struct {
	AccessSecret, RefreshSecret string
	AccessTTL, RefreshTTL       time.Duration
	Issuer                      string
}

type SMTPConfig struct {
	Host, Port, User, Password, From string
}

type GoogleOAuthConfig struct {
	ClientID, ClientSecret, RedirectURL string
}

type MidtransConfig struct {
	ServerKey    string
	IsProduction bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	env := getEnv("APP_ENV", "local")
	switch env {
	case "local", "test", "staging", "production":
	default:
		return nil, fmt.Errorf("configuration error: unsupported APP_ENV %q", env)
	}
	productionDefault := "false"
	if env == "production" {
		productionDefault = "true"
	}
	midtransProduction, err := strconv.ParseBool(getEnv("MIDTRANS_IS_PRODUCTION", productionDefault))
	if err != nil {
		return nil, fmt.Errorf("configuration error: MIDTRANS_IS_PRODUCTION must be true or false")
	}

	cfg := &Config{
		App: AppConfig{
			Env: env, Port: getEnv("APP_PORT", "8080"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		Database: DatabaseConfig{
			Host: getEnv("DB_HOST", "127.0.0.1"), Port: getEnv("DB_PORT", "3306"),
			User: getEnv("DB_USER", "root"), Password: getEnv("DB_PASSWORD", ""), Name: getEnv("DB_NAME", "kawah_putih_db"),
			MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 50), MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: time.Hour,
		},
		Redis: RedisConfig{Addr: getEnv("REDIS_ADDR", "127.0.0.1:6379"), Password: getEnv("REDIS_PASSWORD", ""), DB: getEnvInt("REDIS_DB", 0)},
		JWT: JWTConfig{
			AccessSecret: getEnv("JWT_ACCESS_SECRET", ""), RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
			AccessTTL:  time.Duration(getEnvInt("JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute,
			RefreshTTL: time.Duration(getEnvInt("JWT_REFRESH_TTL_DAYS", 7)) * 24 * time.Hour,
			Issuer:     getEnv("JWT_ISSUER", "KawahPutih"),
		},
		SMTP: SMTPConfig{
			Host: getEnv("SMTP_HOST", ""), Port: getEnv("SMTP_PORT", "587"),
			User: getEnv("SMTP_USER", ""), Password: getEnv("SMTP_PASSWORD", ""), From: getEnv("SMTP_FROM", "no-reply@kawahputih.id"),
		},
		GoogleOAuth: GoogleOAuthConfig{
			ClientID: getEnv("GOOGLE_CLIENT_ID", ""), ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL: getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		},
		Midtrans: MidtransConfig{
			ServerKey:    getEnv("MIDTRANS_SERVER_KEY", ""),
			IsProduction: midtransProduction,
		},
	}

	for _, name := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_NAME"} {
		if strings.TrimSpace(os.Getenv(name)) == "" {
			return nil, fmt.Errorf("Configuration error: required database setting %s is missing.", name)
		}
	}
	if env == "staging" || env == "production" {
		for _, secret := range []struct{ name, value string }{
			{"JWT_ACCESS_SECRET", cfg.JWT.AccessSecret},
			{"JWT_REFRESH_SECRET", cfg.JWT.RefreshSecret},
			{"MIDTRANS_SERVER_KEY", cfg.Midtrans.ServerKey},
		} {
			if strings.TrimSpace(secret.value) == "" {
				return nil, fmt.Errorf("Configuration error: required secret %s is missing.", secret.name)
			}
		}
		if strings.TrimSpace(os.Getenv("DB_PASSWORD")) == "" {
			return nil, fmt.Errorf("Configuration error: required database setting DB_PASSWORD is missing.")
		}
	}
	if env == "production" && !cfg.Midtrans.IsProduction {
		return nil, fmt.Errorf("Configuration error: MIDTRANS_IS_PRODUCTION must be true in production.")
	}
	if env == "local" {
		if cfg.JWT.AccessSecret == "" {
			cfg.JWT.AccessSecret = localSecret()
			_, _ = fmt.Fprintln(os.Stderr, "WARNING: generated a temporary local JWT access secret; configure JWT_ACCESS_SECRET for persistent local sessions.")
		}
		if cfg.JWT.RefreshSecret == "" {
			cfg.JWT.RefreshSecret = localSecret()
			_, _ = fmt.Fprintln(os.Stderr, "WARNING: generated a temporary local JWT refresh secret; configure JWT_REFRESH_SECRET for persistent local sessions.")
		}
	}
	return cfg, nil
}

func localSecret() string {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		panic("configuration error: could not generate temporary local secret")
	}
	return hex.EncodeToString(secret)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
