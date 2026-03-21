package config

type ServerConfiguration struct {
	Port int `env:"PORT, default=8080"`

	PostgresConnectionString string `env:"OPENFOODDIARY_POSTGRES_CONNECTION_STRING"`
	SqliteFile               string `env:"OPENFOODDIARY_SQLITE_PATH, default=.sqlite"`

	UserIdHeader string `env:"OPENFOODDIARY_USERIDHEADER, default=x-openfooddiary-userid"`
	UserId       string `env:"OPENFOODDIARY_USERID"`

	SignoutEndpoint string `env:"OPENFOODDIARY_LOGOUT_ENDPOINT"`

	TemplateDirectory string
	StaticDirectory   string
}
