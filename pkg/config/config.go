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

// GetServerAddress get server and client address
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

// GetDbHost returns DB host
func (e *EnvCache) GetDbHost() string {
	return e.DbHost
}

// GetDbPort returns DB port
func (e *EnvCache) GetDbPort() string {
	return e.DbPort
}

// GetDbUser returns DB username
func (e *EnvCache) GetDbUser() string {
	return e.DbUser
}

// GetDbPassword returns DB user password
func (e *EnvCache) GetDbPassword() string {
	return e.DbPassword
}

// GetDbName returns DB name
func (e *EnvCache) GetDbName() string {
	return e.DbName
}

// GetDbSslmode returns DB sslmode
func (e *EnvCache) GetDbSslmode() string {
	return e.DbSslmode
}
