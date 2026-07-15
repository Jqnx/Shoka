package config

type Config struct {
	Server   ServerConfig
	LogLevel string
	// NOTE: library roots are no longer configured here — they live in the
	// `library` DB table and are managed via the admin API (see internal/library).
	Cache    CacheConfig    `mapstructure:"cache"`
	Images   ImageConfig    `mapstructure:"images"`
	Metadata MetadataConfig `mapstructure:"metadata"`
}

type ServerConfig struct {
	Host string
	Port int
}

type CacheConfig struct {
	Dir     string `mapstructure:"dir"`
	LRUSize int    `mapstructure:"lru_size"`
}

type ImageConfig struct {
	RetentionPeriod int16 `mapstructure:"retention_period"`
}

// MetadataConfig only holds infra-level settings shared by every library.
// Per-source enablement/cookies/api keys/blocklists are per-library now —
// see the library_source DB table and internal/metadata.SourceSettings.
type MetadataConfig struct {
	Flaresolverr Flaresolverr `mapstructure:"flaresolverr"`
}

type Flaresolverr struct {
	URL string `mapstructure:"url"`
}
