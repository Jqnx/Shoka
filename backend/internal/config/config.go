// Package config implements all utility concerning the configuration
// of the application.
package config

import (
	"os"
	"strconv"

	"Shoka/internal/logger"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	Logger            = "zerolog"
	ConfigFile        = "config.yaml"
	ImageExtensions   = []string{"png", "jpg", "jpeg", "gif", "webp"}
	ArchiveExtensions = []string{"zip", "cbz"}
	SourcesList       = []string{SourceNhentai, SourceComicInfo}
)

type Server struct {
	Port int64
}

type Sources struct {
	Flaresolverr Flaresolverr `mapstructure:"flaresolverr"`
	File         File         `mapstructure:"file"`
}

type Flaresolverr struct {
	URL string `mapstructure:"url"`
}

type File struct {
	Format string `mapstructure:"format"`
}

type Database struct {
	DBHost     string
	DBPort     string
	DBDatabase string
	DBUser     string
	DBPassword string
	DBSchema   string
	DBMaxConn  int
}

type Workers struct {
	RedisHost string
	RedisPort string
	Max       int64 `mapstructure:"max"`
}

type Images struct {
	RetentionPeriod int16 `mapstructure:"retention_period"`
}

type Config struct {
	TimeZone   string
	ContentDir string `mapstructure:"content_dir"`
	ThumbDir   string `mapstructure:"thumb_dir"`
	TempDir    string `mapstructure:"temp_dir"`
	Server     Server
	Database   Database
	Images     Images  `mapstructure:"images"`
	Workers    Workers `mapstructure:"workers"`
	Sources    Sources `mapstructure:"sources"`
}

// TODO: Fix that config/settings set through ENV variables don't get written to file!

func LoadConfig(log logger.Logger) (*Config, error) {
	// Set config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Read config file, create new one with defaults if one doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			_, err := os.Create(ConfigFile)
			if err != nil {
				return nil, err
			}
			setDefaults()
			if err := viper.WriteConfig(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var c Config
	setDefaults()
	if err := viper.WriteConfig(); err != nil {
		return nil, err
	}

	if err := getEnv(&c); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func setDefaults() {
	// Directories
	viper.SetDefault("content_dir", "../content")
	viper.SetDefault("thumb_dir", "./thumb")
	viper.SetDefault("temp_dir", "./tmp")

	// Workers
	viper.SetDefault("workers.max", 5)

	// Metadata
	viper.SetDefault("sources.file.format", SourceComicInfo)

	// Images
	viper.SetDefault("images.retention_period", 14)
}

func getEnv(c *Config) error {
	if err := godotenv.Load("../.env"); err != nil {
		return err
	}
	// Config
	c.TimeZone = os.Getenv("TZ")
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		return err
	}
	c.Server.Port = int64(port)

	// Config.Database
	c.Database.DBHost = os.Getenv("DB_HOST")
	c.Database.DBPort = os.Getenv("DB_PORT")
	c.Database.DBDatabase = os.Getenv("DB_DATABASE")
	c.Database.DBUser = os.Getenv("DB_USERNAME")
	c.Database.DBPassword = os.Getenv("DB_PASSWORD")
	c.Database.DBSchema = os.Getenv("DB_SCHEMA")
	c.Database.DBMaxConn = 70

	// Config.Workers
	c.Workers.RedisHost = os.Getenv("REDIS_HOST")
	c.Workers.RedisPort = os.Getenv("REDIS_PORT")

	return nil
}
