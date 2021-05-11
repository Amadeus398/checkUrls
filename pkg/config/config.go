package config

import "github.com/rs/zerolog"

type EnvCache struct {
	ServerAddress string `envconfig:"SERVERADDRESS" required:true`
	LogLevel      string `envconfig:"LOGLEVEL" required:true`
	DbHost        string `envconfig:"HOST" required:true`
	DbPort        string `envconfig:"PORT" required:true`
	DbUser        string `envconfig:"USER" required:true`
	DbPassword    string `envconfig:"PASSWORD" required:true`
	DbName        string `envconfig:"NAME" required:true`
	DbSslmode     string `envconfig:"SSLMODE" required:true`
}

func (e *EnvCache) GetServerAddress() string {
	return e.ServerAddress
}

// GetLogLevel determines which level
// the LogLevel environment
// corresponds to.
func (e EnvCache) GetLogLevel() zerolog.Level {
	switch e.LogLevel {
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.Disabled
	}
}

func (e *EnvCache) GetDbHost() string {
	return e.DbHost
}

func (e *EnvCache) GetDbPort() string {
	return e.DbPort
}

func (e *EnvCache) GetDbUser() string {
	return e.DbUser
}

func (e *EnvCache) GetDbPassword() string {
	return e.DbPassword
}

func (e *EnvCache) GetDbName() string {
	return e.DbName
}

func (e *EnvCache) GetDbSslmode() string {
	return e.DbSslmode
}
