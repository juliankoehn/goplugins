package config

import "github.com/kelseyhightower/envconfig"

type (
	// Config provides the system configuration.
	Config struct {
		App      App
		Database Database
	}

	// App the basic Application configuration
	App struct {
		Name  string `envconfig:"APP_NAME" default:"Luminate"`
		Env   string `envconfig:"APP_ENV" default:"production"`
		Debug bool   `envconfig:"APP_DEBUG" default:"false"`
		URL   string `envconfig:"APP_URL" default:"http://localhost"`
		// This key is used by the encrypter service and should be set
		// to a random, 32 character string, otherwise these encrypted strings
		// will not be safe. Please do this before deploying an application!
		Key string `envconfig:"APP_KEY" required:"true"`
	}

	// Database provides the database configuration.
	Database struct {
		Driver         string `envconfig:"DATABASE_DRIVER"     default:"sqlite3"`
		Datasource     string `envconfig:"DATABASE_DATASOURCE" default:"core.sqlite"`
		MaxConnections int    `envconfig:"DATABASE_CONNECTIONS" default:"11"`
	}
)

// Environ returns the settings from the environment.
func Environ() (Config, error) {
	cfg := Config{}
	err := envconfig.Process("", &cfg)
	// defaultAddress(&cfg)
	return cfg, err
}
