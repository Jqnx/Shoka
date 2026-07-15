package config

import (
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
	ConfigFile string = "config.yaml"
	DataDir    string = "./data"
)

func LoadConfig() (*Config, error) {
	setDefaults()
	// Set config file
	viper.SetConfigName(filepath.Base(ConfigFile))
	viper.SetConfigType(strings.TrimLeft(filepath.Ext(ConfigFile), "."))
	viper.AddConfigPath(DataDir)

	// Read config file, create new one with defaults if one doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			_, err := os.Create(filepath.Join(DataDir, ConfigFile))
			if err != nil {
				return nil, err
			}
			if err := viper.WriteConfig(); err != nil {
				return nil, err
			}
		} else {
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

func setDefaults() {
	// Directories
	viper.SetDefault("temp_dir", "./tmp")
	viper.SetDefault("cache.dir", "../cache")
	viper.SetDefault("cache.lru_size", 128)

	// Images
	viper.SetDefault("images.retention_period", 14)
}

func getEnv(c *Config) error {
	if err := godotenv.Load(); err != nil {
		return err
	}

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
