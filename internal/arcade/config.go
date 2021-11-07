package arcade

import (
	"arcadium.dev/core/config"
)

type (
	// Config contains the configuration of the infra server.
	Config struct {
		Logger config.Logger
		DB     config.SQLDatabase
		Server config.Server
	}
)

// NewConfig returns the configuration of the infra server.
func NewConfig(opts ...config.Option) (*Config, error) {
	var err error
	c := &Config{}
	if c.Logger, err = config.NewLogger(opts...); err != nil {
		return nil, err
	}
	if c.DB, err = config.NewSQLDatabase(opts...); err != nil {
		return nil, err
	}
	if c.Server, err = config.NewServer(opts...); err != nil {
		return nil, err
	}
	return c, nil
}
