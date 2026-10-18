package config

import (
	"cmp"
	"net"
	"net/url"
	"strconv"
)

type Config struct {
	App          AppConfig      `mapstructure:"app"`
	Database     DatabaseConfig `mapstructure:"database"`
	JwtSecretKey string         `mapstructure:"jwt_secret_key"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Port        int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Name       string `mapstructure:"name"`
	User       string `mapstructure:"user"`
	SSLMode    string `mapstructure:"sslmode"`
	Password   string `mapstructure:"password"`
	Migrations struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"migrations"`
}

// URL is the connection string for this database. The pool and the migrations
// both connect with it, so where the app connects is decided in one place.
func (d DatabaseConfig) URL() string {
	// Build a URL rather than a key=value string so special characters in the
	// password are escaped correctly.
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.User, d.Password),
		Host:   net.JoinHostPort(d.Host, strconv.Itoa(d.Port)),
		Path:   "/" + d.Name,
		// Local Postgres has no TLS; set sslmode to require (or verify-full) for RDS.
		RawQuery: url.Values{"sslmode": {cmp.Or(d.SSLMode, "disable")}}.Encode(),
	}
	return u.String()
}
