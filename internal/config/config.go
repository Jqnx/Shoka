package config

import (
	"Shoka/internal/logger"
	"os"

	"github.com/spf13/viper"
)

var (
	Logger            = "zerolog"
	ConfigFile        = "config.yaml"
	ImageExtensions   = []string{"png", "jpg", "jpeg", "gif", "webp"}
	ArchiveExtensions = []string{"zip", "cbz"}
	SourcesList       = []string{SourceNhentai, SourceNhentaiSearch, SourceComicInfo}
)

type Server struct {
	Port int64
}

type Sources struct {
	NHentai NHentai `mapstructure:"nhentai"`
}

type NHentai struct {
	CSRFToken string `mapstructure:"csrftoken"`
	UserAgent string `mapstructure:"useragent"`
}

type Database struct {
	DBHost     string `mapstructure:"host"`
	DBPort     string `mapstructure:"port"`
	DBDatabase string `mapstructure:"database"`
	DBUser     string `mapstructure:"user"`
	DBPassword string `mapstructure:"password"`
	DBSchema   string `mapstructure:"schema"`
}

type Workers struct {
	RedisHost string ` mapstructure:"redis_host"`
	RedisPort string ` mapstructure:"redis_port"`
	Max       int64  `mapstructure:"max"`
}

type Downloader struct {
	DownloadDir string `mapstructure:"download_dir"`
	RateLimit   int64  `mapstructure:"rate_limit"`
	SaveFileExt string `mapstructure:"save_file_ext"`
}

type Config struct {
	TimeZone   string `mapstructure:"tz"`
	ContentDir string `mapstructure:"content_dir"`
	ThumbDir   string `mapstructure:"thumb_dir"`
	TempDir    string `mapstructure:"temp_dir"`
	Server     Server
	Database   Database   `mapstructure:"db"`
	Workers    Workers    `mapstructure:"workers"`
	Sources    Sources    `mapstructure:"sources"`
	Downloader Downloader `mapstructure:"downloader"`
}

// TODO: Fix that config/settings set through ENV variables don't get written to file!
func LoadConfig(log logger.Logger) (*Config, error) {
	viper.AutomaticEnv()

	// Set config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Read config file, create new one with defaults if one doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			os.Create(ConfigFile)
			setDefaults()
			viper.WriteConfig()
		} else {
			return nil, err
		}
	}

	var c Config
	setDefaults()
	viper.WriteConfig()

	c.Server.Port = 8081
	viper.SetDefault("db.schema", "public")

	getEnv()

	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func setDefaults() {
	viper.SetDefault("content_dir", "content")
	viper.SetDefault("thumb_dir", "thumb")
	viper.SetDefault("temp_dir", "tmp")
	viper.SetDefault("downloader.download_dir", "downloads")
	viper.SetDefault("downloader.rate_limit", 500)
	viper.SetDefault("downloader.save_file_ext", "cbz")
	viper.SetDefault("workers.max", 5)
	viper.SetDefault("tz", "Etc/UTC")
}

func getEnv() {
	// Config
	viper.BindEnv("tz", "TZ")

	// Config.Database
	viper.BindEnv("db.host", "DB_HOST")
	viper.BindEnv("db.port", "DB_PORT")
	viper.BindEnv("db.database", "DB_DATABASE")
	viper.BindEnv("db.user", "DB_USERNAME")
	viper.BindEnv("db.password", "DB_PASSWORD")
	viper.BindEnv("db.schema", "DB_SCHEMA")

	// Config.Workers
	viper.BindEnv("workers.redis_host", "REDIS_HOST")
	viper.BindEnv("workers.redis_port", "REDIS_PORT")
}
