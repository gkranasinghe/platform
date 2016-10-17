package mongo

/* CHECKLIST
 * [x] Uses interfaces as appropriate
 * [x] Private package variables use underscore prefix
 * [x] All parameters validated
 * [x] All errors handled
 * [x] Reviewed for concurrency safety
 * [x] Code complete
 * [x] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/app"
	"github.com/tidepool-org/platform/store/mongo"
)

type Config struct {
	*mongo.Config `anonymous:"true"`
	PasswordSalt  string `json:"passwordSalt"`
}

func (c *Config) Validate() error {
	if c.Config == nil {
		return app.Error("mongo", "config is missing")
	}
	if err := c.Config.Validate(); err != nil {
		return err
	}
	if c.PasswordSalt == "" {
		return app.Error("mongo", "password salt is missing")
	}
	return nil
}

func (c *Config) Clone() *Config {
	clone := &Config{
		Config:       c.Config.Clone(),
		PasswordSalt: c.PasswordSalt,
	}
	return clone
}