package config

type Config struct {
	Server     ServerConfig
	LogLevel   string
	LibraryDir string         `mapstructure:"library_dir"`
	Cache      CacheConfig    `mapstructure:"cache"`
	Images     ImageConfig    `mapstructure:"images"`
	Metadata   MetadataConfig `mapstructure:"metadata"`
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

type MetadataConfig struct {
	Flaresolverr Flaresolverr            `mapstructure:"flaresolverr"`
	Sources      map[string]SourceConfig `mapstructure:"sources"`
}

type SourceConfig struct {
	Enabled           bool     `mapstructure:"enabled"`
	Cookies           string   `mapstructure:"cookies"`
	APIKey            string   `mapstructure:"api_key"`
	MagazineBlocklist []string `mapstructure:"magazine_blocklist"`
	MiscBlocklist     []string `mapstructure:"misc_blocklist"`
}

type Flaresolverr struct {
	URL string `mapstructure:"url"`
}

func (m *MetadataConfig) IsEnabled(name string) bool {
	sc, ok := m.Sources[name]
	if !ok {
		return false
	}
	return sc.Enabled
}

func (m *MetadataConfig) GetSource(name string) SourceConfig {
	return m.Sources[name]
}
