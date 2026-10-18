package config

type Config struct {
	App struct {
		Name        string `mapstructure:"name"`
		Environment string `mapstructure:"environment"`
		Port        int    `mapstructure:"port"`
	} `mapstructure:"app"`

	Database struct {
		Host       string `mapstructure:"host"`
		Port       int    `mapstructure:"port"`
		Name       string `mapstructure:"name"`
		User       string `mapstructure:"user"`
		SSLMode    string `mapstructure:"sslmode"`
		Password   string `mapstructure:"password"`
		Migrations struct {
			Path string `mapstructure:"path"`
		} `mapstructure:"migrations"`
		URL string `mapstructure:"url"`
	} `mapstructure:"database"`

	DatabaseURL  string `mapstructure:"-"`
	JwtSecretKey string `mapstructure:"-"`
}
