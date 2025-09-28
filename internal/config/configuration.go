package config

type ServerConfiguration struct {
	Port int `env:"PORT, default=8080"`
	//PostgresConnectionString string `env:"POSTGRES_CONNECTION_STRING, required"`
}
