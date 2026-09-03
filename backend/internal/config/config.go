package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Env      string         `mapstructure:"env"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	MinIO    MinIOConfig    `mapstructure:"minio"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// JWTConfig holds JWT authentication configuration.
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
	Issuer     string `mapstructure:"issuer"`
}

// MinIOConfig holds MinIO object storage configuration.
type MinIOConfig struct {
	Endpoint   string `mapstructure:"endpoint"`
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	BucketName string `mapstructure:"bucket_name"`
	UseSSL     bool   `mapstructure:"use_ssl"`
}

// Load reads configuration from files and environment variables.
func Load() (*Config, error) {
	v := viper.New()

	// Set default config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")

	// Read config.yaml
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// Merge config.dev.yaml if exists
	v.SetConfigName("config.dev")
	if err := v.MergeInConfig(); err != nil {
		// Ignore if no dev config file
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Environment variable overrides
	v.SetEnvPrefix("LIMS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Apply defaults for zero values
	cfg.applyDefaults()

	// Validate critical configuration
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Env == "" {
		c.Env = "development"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Database.Port == 0 {
		c.Database.Port = 5432
	}
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 10
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 100
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 30
	}
	if c.JWT.ExpireHour == 0 {
		c.JWT.ExpireHour = 24
	}
	if c.JWT.Issuer == "" {
		c.JWT.Issuer = "lims-system"
	}
}

func (c *Config) validate() error {
	var errs []string

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, "server.port must be between 1 and 65535")
	}

	if c.Database.Host == "" {
		errs = append(errs, "database.host is required")
	}
	if c.Database.User == "" {
		errs = append(errs, "database.user is required")
	}
	if c.Database.DBName == "" {
		errs = append(errs, "database.dbname is required")
	}

	if c.JWT.Secret == "" {
		errs = append(errs, "jwt.secret is required")
	} else if c.Env == "production" && (c.JWT.Secret == "lims-jwt-secret-change-in-production" || len(c.JWT.Secret) < 32) {
		errs = append(errs, "jwt.secret must be at least 32 chars and not use default value in production")
	}
	if c.JWT.ExpireHour < 1 {
		errs = append(errs, "jwt.expire_hour must be >= 1")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}
