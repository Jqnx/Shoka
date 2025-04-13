package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
)

var (
	ImageExtensions   = []string{"png", "jpg", "jpeg", "gif", "webp"}
	ArchiveExtensions = []string{"zip", "cbz"}
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
	Max int
}

type Config struct {
	Dir      string
	Server   Server
	Database Database
	Workers  Workers
}

func setDefaults() Config {
	path := GetDefaultPath()
	return Config{
		Dir: path,
		Workers: Workers{
			Max: 100,
		},
	}
}

func GetDefaultPath() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Panic(err)
		return ""
	}
	path := filepath.Join(wd, "content")
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
