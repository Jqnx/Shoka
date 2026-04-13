package config

type Config struct {
	LogLevel   string
	LibraryDir string `mapstructure:"library_dir"`
	CacheDir   string `mapstructure:"cache_dir"`
	Images     Images `mapstructure:"images"`
	Server     Server
}

type Server struct {
	Host string
	Port int
}

type Images struct {
	RetentionPeriod int16 `mapstructure:"retention_period"`
}

type Metadata struct {
	EnabledSources []string     `mapstructure:"enabled_sources"`
	Flaresolverr   Flaresolverr `mapstructure:"flaresolverr"`
}

type Flaresolverr struct {
	URL string `mapstructure:"url"`
}
