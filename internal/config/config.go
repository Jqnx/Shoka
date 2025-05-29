package config

import (
	"log"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
)

var (
	ImageExtensions   = []string{"png", "jpg", "jpeg", "gif", "webp"}
	ArchiveExtensions = []string{"zip", "cbz"}
	ComicInfoFile     = "ComicInfo.xml"
)

type Server struct {
	Port string `envconfig:"PORT"`
}

type Database struct {
	DBHost     string `envconfig:"DB_HOST"`
	DBPort     string `envconfig:"DB_PORT"`
	DBDatabase string `envconfig:"DB_DATABASE"`
	DBUser     string `envconfig:"DB_USERNAME"`
	DBPassword string `envconfig:"DB_PASSWORD"`
	DBSchema   string `envconfig:"DB_SCHEMA"`
}

type Workers struct {
	RedisHost string `envconfig:"REDIS_HOST"`
	RedisPort string `envconfig:"REDIS_PORT"`
	Max       int
}

type Config struct {
	ContentDir string
	ThumbDir   string
	Server     Server
	Database   Database
	Workers    Workers
}

func setDefaults() Config {
	// content := GetDefaultContentPath()
	// thumb := GetThumbPath()
	return Config{
		ContentDir: "content",
		ThumbDir:   "thumb",
		Workers: Workers{
			Max: 5,
		},
	}
}

func GetDefaultContentPath() string {
	path, err := filepath.Abs("content")
	if err != nil {
		log.Panic(err)
		return ""
	}
	return path
}

func GetThumbPath() string {
	path, err := filepath.Abs("thumb")
	if err != nil {
		log.Panic(err)
		return ""
	}
	return path
}

func LoadConfig() *Config {
	cfg := setDefaults()
	readEnv(&cfg)
	return &cfg
}

func readEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Panic(err)
	}
}
