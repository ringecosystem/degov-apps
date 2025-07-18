package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Environment int

const (
	Production Environment = iota
	Development
	Staging
)

func (e Environment) String() string {
	switch e {
	case Development:
		return "development"
	case Staging:
		return "staging"
	case Production:
		return "production"
	default:
		return "production"
	}
}

func (e Environment) IsDevelopment() bool {
	return e == Development
}

func (e Environment) IsProduction() bool {
	return e == Production
}

func (e Environment) IsStaging() bool {
	return e == Staging
}

// Config holds all application configuration
type Config struct {
	viper *viper.Viper
}

// Global configuration instance
var globalConfig *Config

// InitConfig initializes the global configuration using viper
func InitConfig() error {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Configure viper to read environment variables
	v.AutomaticEnv()
	v.AllowEmptyEnv(true)

	// Set key replacer to handle case conversion
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	globalConfig = &Config{viper: v}

	slog.Info("Configuration initialized successfully")
	return nil
}

// GetConfig returns the global configuration instance
func GetConfig() *Config {
	if globalConfig == nil {
		if err := InitConfig(); err != nil {
			slog.Error("Failed to initialize config", "error", err)
			os.Exit(1)
		}
	}
	return globalConfig
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("PORT", "8080")

	// Database defaults
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_SSLMODE", "disable")

	// Log defaults
	v.SetDefault("LOG_FORMAT", "json")

	// Environment defaults
	v.SetDefault("APP_ENV", "production")

	// Task defaults
	v.SetDefault("TASK_DAO_SYNC_ENABLED", true)
	v.SetDefault("TASK_DAO_SYNC_INTERVAL", "5m")
	v.SetDefault("TASK_NOTIFICATION_CLEANUP_ENABLED", true)
	v.SetDefault("TASK_NOTIFICATION_CLEANUP_INTERVAL", "30m")
}

// Server configuration methods
func (c *Config) GetPort() string {
	return c.viper.GetString("PORT")
}

// Database configuration methods
func (c *Config) GetDBHost() string {
	return c.viper.GetString("DB_HOST")
}

func (c *Config) GetDBUser() string {
	return c.viper.GetString("DB_USER")
}

func (c *Config) GetDBPassword() string {
	return c.viper.GetString("DB_PASSWORD")
}

func (c *Config) GetDBName() string {
	return c.viper.GetString("DB_NAME")
}

func (c *Config) GetDBPort() string {
	return c.viper.GetString("DB_PORT")
}

func (c *Config) GetDBSSLMode() string {
	return c.viper.GetString("DB_SSLMODE")
}

// Log configuration methods
func (c *Config) GetLogFormat() string {
	return c.viper.GetString("LOG_FORMAT")
}

// Environment configuration methods
func (c *Config) GetAppEnv() Environment {
	// Try GO_ENV first, then APP_ENV
	env := c.viper.GetString("APP_ENV")
	return parseEnvironment(strings.ToLower(strings.TrimSpace(env)))
}

// Task configuration methods
func (c *Config) GetTaskDAOSyncEnabled() bool {
	return c.viper.GetBool("TASK_DAO_SYNC_ENABLED")
}

func (c *Config) GetTaskDAOSyncInterval() time.Duration {
	return c.viper.GetDuration("TASK_DAO_SYNC_INTERVAL")
}

func (c *Config) GetTaskNotificationCleanupEnabled() bool {
	return c.viper.GetBool("TASK_NOTIFICATION_CLEANUP_ENABLED")
}

func (c *Config) GetTaskNotificationCleanupInterval() time.Duration {
	return c.viper.GetDuration("TASK_NOTIFICATION_CLEANUP_INTERVAL")
}

// Generic configuration methods
func (c *Config) GetString(key string) string {
	return c.viper.GetString(key)
}

func (c *Config) GetStringWithDefault(key, defaultValue string) string {
	value := c.viper.GetString(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func (c *Config) GetStringRequired(key string) string {
	value := c.viper.GetString(key)
	if value == "" {
		slog.Error("Required configuration key is not set or empty", slog.String("key", key))
		os.Exit(1)
	}
	return value
}

func (c *Config) GetBool(key string) bool {
	return c.viper.GetBool(key)
}

func (c *Config) GetInt(key string) int {
	return c.viper.GetInt(key)
}

func (c *Config) GetDuration(key string) time.Duration {
	return c.viper.GetDuration(key)
}

func parseEnvironment(env string) Environment {
	switch env {
	case "development", "dev":
		return Development
	case "staging", "stage":
		return Staging
	case "production", "prod":
		return Production
	default:
		return Production
	}
}

// Convenience functions for backward compatibility
func GetAppEnv() Environment {
	return GetConfig().GetAppEnv()
}

func GetLogFormat() string {
	return GetConfig().GetLogFormat()
}

func GetString(key string) string {
	return GetConfig().GetString(key)
}

func GetStringWithDefault(key, defaultValue string) string {
	return GetConfig().GetStringWithDefault(key, defaultValue)
}

func GetStringRequired(key string) string {
	return GetConfig().GetStringRequired(key)
}

func GetBool(key string) bool {
	return GetConfig().GetBool(key)
}

func GetInt(key string) int {
	return GetConfig().GetInt(key)
}

func GetDuration(key string) time.Duration {
	return GetConfig().GetDuration(key)
}
