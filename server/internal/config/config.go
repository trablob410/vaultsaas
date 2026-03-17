package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Server
	Port        string `envconfig:"PORT" default:"8080"`
	CORSOrigins string `envconfig:"CORS_ORIGINS" default:"http://localhost:3000,http://localhost:8443"`

	// Database
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// MinIO
	MinioEndpoint  string `envconfig:"MINIO_ENDPOINT" default:"localhost:9000"`
	MinioAccessKey string `envconfig:"MINIO_ACCESS_KEY" required:"true"`
	MinioSecretKey string `envconfig:"MINIO_SECRET_KEY" required:"true"`
	MinioUseSSL    bool   `envconfig:"MINIO_USE_SSL" default:"false"`
	MinioBucket    string `envconfig:"MINIO_BUCKET" default:"valt-secrets"`

	// JWT
	JWTPrivateKeyPath string `envconfig:"JWT_PRIVATE_KEY_PATH" default:"./keys/private.pem"`
	JWTPublicKeyPath  string `envconfig:"JWT_PUBLIC_KEY_PATH" default:"./keys/public.pem"`

	// SMTP (optional - no-op mode when host is empty)
	SMTPHost     string `envconfig:"SMTP_HOST" default:""`
	SMTPPort     int    `envconfig:"SMTP_PORT" default:"587"`
	SMTPUser     string `envconfig:"SMTP_USER" default:""`
	SMTPPassword string `envconfig:"SMTP_PASSWORD" default:""`
	SMTPFrom     string `envconfig:"SMTP_FROM" default:"noreply@valt.dev"`

	// Google OAuth
	GoogleClientID     string `envconfig:"GOOGLE_CLIENT_ID" default:""`
	GoogleClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET" default:""`
	GoogleRedirectURL  string `envconfig:"GOOGLE_REDIRECT_URL" default:"http://localhost:8080/api/v1/auth/google/callback"`
	DashboardURL       string `envconfig:"DASHBOARD_URL" default:"http://localhost:3000"`
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return &cfg, nil
}
