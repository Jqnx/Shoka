package config

// Config is the top-level application configuration assembled from the YAML
// file and environment overrides.
type Config struct {
	Server   ServerConfig
	LogLevel string
	// NOTE: library roots are no longer configured here — they live in the
	// `library` DB table and are managed via the admin API (see internal/library).
	Cache    CacheConfig    `mapstructure:"cache"`
	Images   ImageConfig    `mapstructure:"images"`
	Metadata MetadataConfig `mapstructure:"metadata"`
}

// ServerConfig holds the HTTP listener settings for the API server.
type ServerConfig struct {
	Host string
	Port int
}

// CacheConfig holds the on-disk cache location and its in-memory LRU size.
type CacheConfig struct {
	Dir     string `mapstructure:"dir"`
	LRUSize int    `mapstructure:"lru_size"`
}

// ImageConfig holds image processing settings such as how long generated
// thumbnails are retained.
type ImageConfig struct {
	RetentionPeriod int16 `mapstructure:"retention_period"`
}

// MetadataConfig only holds infra-level settings shared by every library.
// Per-source enablement/cookies/api keys/blocklists are per-library now —
// see the library_source DB table and internal/metadata.SourceSettings.
type MetadataConfig struct {
	Flaresolverr Flaresolverr `mapstructure:"flaresolverr"`
}

// Flaresolverr configures the FlareSolverr proxy used to bypass Cloudflare
// challenges when fetching remote metadata.
type Flaresolverr struct {
	URL string `mapstructure:"url"`
}
