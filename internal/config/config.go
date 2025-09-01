// File: internal/config/config.go
package config

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config stores all configuration for the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	AppEnv   string         `mapstructure:"APP_ENV" validate:"required,oneof=development staging production"`
	Server   ServerConfig   `mapstructure:",squash"`
	Mongo    MongoConfig    `mapstructure:",squash"`
	Redis    RedisConfig    `mapstructure:",squash"`
	JWT      JWTConfig      `mapstructure:",squash"`
	Log      LogConfig      `mapstructure:",squash"`
	CSRF     CSRFConfig     `mapstructure:",squash"`
	Security SecurityConfig `mapstructure:",squash"`
	BaseURL  string         `mapstructure:"BASE_URL" validate:"required,url"`
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Host string `mapstructure:"SERVER_HOST" validate:"required"`
	Port int    `mapstructure:"SERVER_PORT" validate:"required"`
}

// MongoConfig holds MongoDB connection details.
type MongoConfig struct {
	URI string `mapstructure:"MONGO_URI" validate:"required"`
}

// RedisConfig holds Redis connection details.
type RedisConfig struct {
	Addr     string `mapstructure:"REDIS_ADDR" validate:"required"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

// JWTConfig holds JWT signing and validation details.
type JWTConfig struct {
	SecretKey        string `mapstructure:"JWT_SECRET_KEY" validate:"required,min=32"`
	PrivateKeyBase64 string `mapstructure:"JWT_PRIVATE_KEY_BASE64" validate:"required"`
	Issuer           string `mapstructure:"JWT_ISSUER" validate:"required"`

	// These fields are for viper to read the integer values from .env
	AccessTokenLifespanMinutes int64 `mapstructure:"JWT_ACCESS_TOKEN_LIFESPAN_MINUTES" validate:"required"`
	RefreshTokenLifespanHours  int64 `mapstructure:"JWT_REFRESH_TOKEN_LIFESPAN_HOURS" validate:"required"`

	// These fields are for the application to use, populated after loading config.
	// They don't have mapstructure tags, so viper ignores them.
	AccessTokenLifespan  time.Duration
	RefreshTokenLifespan time.Duration
}

// CSRFConfig holds CSRF protection configuration.
type CSRFConfig struct {
	AuthKey string `mapstructure:"CSRF_AUTH_KEY" validate:"required,len=32"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `mapstructure:"LOG_LEVEL" validate:"required,oneof=debug info warn error"`
}

// SecurityConfig holds security-related configuration.
type SecurityConfig struct {
	AllowedOrigins []string `mapstructure:"CORS_ALLOWED_ORIGINS"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (*Config, error) {
	// Set default values
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("JWT_ACCESS_TOKEN_LIFESPAN_MINUTES", 15)
	viper.SetDefault("JWT_REFRESH_TOKEN_LIFESPAN_HOURS", 168)
	viper.SetDefault("JWT_ISSUER", "oauth2-provider")
	viper.SetDefault("BASE_URL", "http://localhost:8080")
	viper.SetDefault("CSRF_AUTH_KEY", "01234567890123456789012345678901")

	// Tell viper to look for a file named .env in the current directory
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// Enable automatic environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Attempt to read the config file. It's okay if it doesn't exist;
	// environment variables will take precedence.
	_ = viper.ReadInConfig()

	var config Config
	// Unmarshal the config into our struct
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Manually convert the loaded integer values into time.Duration
	config.JWT.AccessTokenLifespan = time.Duration(config.JWT.AccessTokenLifespanMinutes) * time.Minute
	config.JWT.RefreshTokenLifespan = time.Duration(config.JWT.RefreshTokenLifespanHours) * time.Hour

	// Validate the configuration
	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
