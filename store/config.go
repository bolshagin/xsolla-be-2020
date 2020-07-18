package store

type Config struct {
	DBName   string `toml:"dbname"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

func NewConfig() *Config {
	return &Config{}
}
