package config

type ServerConfiguration struct {
	Port                     int    `env:"PORT, default=8080"`
	PostgresConnectionString string `env:"OPENFOODDIARY_POSTGRES_CONNECTION_STRING, required"`

	UserIdHeader string `env:"OPENFOODDIARY_USERIDHEADER, default=x-openfooddiary-userid"`
	UserId       string `env:"OPENFOODDIARY_USERID"`
}
