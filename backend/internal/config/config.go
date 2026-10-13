// Package config loads the application configuration from the YAML file in
// DataDir, applies environment overrides, and validates the result.
package config

import (
	"Shoka/internal/util"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	// ConfigFile is the name of the YAML configuration file loaded from DataDir.
	ConfigFile = "config.yaml"

	// DataDir is the directory holding the configuration file and other runtime state.
	DataDir = "./data"
)

// LoadConfig reads the YAML configuration from DataDir, creating it with
// defaults if it does not exist yet, then applies environment overrides and
// validates the result.
func LoadConfig() (*Config, error) {
	if err := util.EnsureDir(DataDir); err != nil {
		return nil, err
	}

	setDefaults()
	// Set config file
	viper.SetConfigName(filepath.Base(ConfigFile))
	viper.SetConfigType(strings.TrimLeft(filepath.Ext(ConfigFile), "."))
	viper.AddConfigPath(DataDir)

	// Read config file, create new one with defaults if one doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if err := createDefaultConfig(err); err != nil {
			return nil, err
		}
	}

	viper.WatchConfig()

	var c Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	if err := getEnv(&c); err != nil {
		return nil, err
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &c, nil
}

// createDefaultConfig writes a fresh config file with the registered
// defaults when readErr says no config file exists yet. Any other read
// failure — or a failure actually creating the file — is returned as-is.
func createDefaultConfig(readErr error) error {
	if _, ok := errors.AsType[viper.ConfigFileNotFoundError](readErr); !ok {
		return readErr
	}

	if _, err := os.Create(filepath.Join(DataDir, ConfigFile)); err != nil {
		return err
	}

	return viper.WriteConfig()
}

func setDefaults() {
	// Directories
	viper.SetDefault("temp_dir", "./tmp")
	viper.SetDefault("cache.dir", "../cache")
	viper.SetDefault("cache.lru_size", 128)

	// Images
	viper.SetDefault("images.retention_period", 14)
}

func getEnv(c *Config) error {
	// A missing .env file is expected; environment variables may come from the real environment.
	_ = godotenv.Load()

	if os.Getenv("HOST") == "" {
		c.Server.Host = "localhost"
	} else {
		c.Server.Host = os.Getenv("HOST")
	}

	if os.Getenv("PORT") == "" {
		c.Server.Port = 8080
	} else {
		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			return err
		}

		c.Server.Port = port
	}

	if os.Getenv("LOG_LEVEL") == "" {
		c.LogLevel = "info"
	} else {
		c.LogLevel = strings.ToLower(os.Getenv("LOG_LEVEL"))
	}

	return nil
}

// Validate reports whether the resolved configuration is usable, checking the
// server host, the server port range, and the log level.
func (c *Config) Validate() error {
	ip := net.ParseIP(c.Server.Host)
	if ip == nil && c.Server.Host != "localhost" {
		return fmt.Errorf("invalid server host %s", c.Server.Host)
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port %d", c.Server.Port)
	}

	validLogLevels := map[string]bool{
		"panic": true,
		"fatal": true,
		"error": true,
		"warn":  true,
		"info":  true,
		"debug": true,
		"trace": true,
	}

	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level %s", c.LogLevel)
	}

	return nil
}
