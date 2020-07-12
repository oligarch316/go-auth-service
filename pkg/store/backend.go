package store

import (
	"encoding/json"
	"fmt"

	"github.com/oligarch316/go-auth-service/pkg/model"
	"github.com/oligarch316/go-auth-service/pkg/store/sqlite"
	"github.com/oligarch316/go-skeleton/pkg/observ"
	"go.uber.org/zap/zapcore"
)

// TODO: Should the Backend methods all take context as 1st arg?

// Backend TODO.
type Backend interface {
	Close() error

	// Invites
	CreateInvites(ownerID string, count int) ([]model.Invite, error)
	ReadInvite(id string) (model.Invite, error)
	UpdateInvite(id string, mData model.InviteUpdate) error
	DeleteInvite(id string) error
	LookupInvites(ownerID string) ([]model.Invite, error)

	// Users
	CreateUser(name, password string, mData model.UserUpdate) (model.User, error)
	ReadUser(id string) (model.User, error)
	UpdateUser(id string, mData model.UserUpdate) error
	DeleteUser(id string) error
	LookupUser(name string) (model.User, error)

	// Combined
	CreateUserAndDeleteInvite(inviteID, name, password string, mData model.UserUpdate) (model.User, error)
}

const (
	backendTypeSQLite = "sqlite"
	backendTypeMongo  = "mongo"
)

type sqliteConfig struct{ sqlite.Config }

func (sc sqliteConfig) Build(corelet *observ.Corelet) (Backend, error) {
	res, err := sqlite.New(sc.Config, corelet)
	return res, err
}

// Config TODO.
type Config struct {
	dType   string
	dynamic interface {
		Build(*observ.Corelet) (Backend, error)
		zapcore.ObjectMarshaler
	}
}

// DefaultConfig TODO.
func DefaultConfig() Config {
	return Config{
		dType:   backendTypeSQLite,
		dynamic: &sqliteConfig{Config: sqlite.DefaultConfig()},
	}
}

// Build TODO.
func (c Config) Build(corelet *observ.Corelet) (Backend, error) { return c.dynamic.Build(corelet) }

// UnmarshalJSON TODO.
func (c *Config) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	c.dType = tmp.Type

	switch tmp.Type {
	case backendTypeSQLite:
		c.dynamic = &sqliteConfig{Config: sqlite.DefaultConfig()}
	case backendTypeMongo:
		return fmt.Errorf("type '%s' not yet implemented", tmp.Type)
	default:
		return fmt.Errorf("unknown type '%s'", tmp.Type)
	}

	if tmp.Data == nil {
		return nil
	}

	return json.Unmarshal(tmp.Data, c.dynamic)
}

// MarshalJSON TODO.
func (c Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: c.dType,
		Data: c.dynamic,
	})
}

// MarshalLogObject TODO.
func (c Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", c.dType)
	enc.AddObject("data", c.dynamic)
	return nil
}
