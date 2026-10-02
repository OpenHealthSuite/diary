package config

type ServerConfiguration struct {
	Port int `env:"PORT, default=8080"`

	PostgresConnectionString string `env:"OPENFOODDIARY_POSTGRES_CONNECTION_STRING"`
	SqliteFile               string `env:"OPENFOODDIARY_SQLITE_PATH, default=.sqlite"`

	UserId string `env:"OPENFOODDIARY_USERID"`

	Oauth2Issuer                 string `env:"OPENFOODDIARY_OAUTH2_ISSUER"`
	Oauth2ClientId               string `env:"OPENFOODDIARY_OAUTH2_CLIENT_ID"`
	Oauth2ClientSecret           string `env:"OPENFOODDIARY_OAUTH2_CLIENT_SECRET"`
	Oauth2RedirectUrl            string `env:"OPENFOODDIARY_OAUTH2_REDIRECT_URL"`
	Oauth2SkipIssuerVerification bool   `env:"OPENFOODDIARY_OAUTH2_SKIP_ISSUER_VERIFICATION, default=false"`

	SessionCookieName string `env:"OPENFOODDIARY_SESSION_COOKIE_NAME, default=openfooddiary-session"`
	SessionMaxAge     int    `env:"OPENFOODDIARY_SESSION_MAX_AGE, default=86400"`
	SessionSecure     bool   `env:"OPENFOODDIARY_SESSION_SECURE, default=true"`
	SessionRedisUrl   string `env:"OPENFOODDIARY_SESSION_REDIS_URL"`
	SessionSecret     string `env:"OPENFOODDIARY_SESSION_SECRET"`

	SignoutEndpoint string `env:"OPENFOODDIARY_LOGOUT_ENDPOINT"`

	TemplateDirectory string
	StaticDirectory   string
}
