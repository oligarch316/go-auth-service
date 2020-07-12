package sqlite

import (
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/oligarch316/go-auth-service/pkg/model"
	"github.com/oligarch316/go-skeleton/pkg/observ"
	"go.uber.org/zap/zapcore"

	// TODO: Explain to golint for some reason
	_ "github.com/mattn/go-sqlite3"
)

const driverName = "sqlite3"

// ConfigTableNames TODO.
type ConfigTableNames struct {
	Invites string `json:"invites"`
	Users   string `json:"users"`
}

// MarshalLogObject TODO.
func (ctn ConfigTableNames) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("invites", ctn.Invites)
	enc.AddString("users", ctn.Users)
	return nil
}

// Config TODO.
type Config struct {
	DBPath     string           `json:"dbPath"`
	TableNames ConfigTableNames `json:"tableNames"`
}

// MarshalLogObject TODO.
func (c Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("dbPath", c.DBPath)
	enc.AddObject("tables", c.TableNames)
	return nil
}

// DefaultConfig TODO.
func DefaultConfig() Config {
	return Config{
		DBPath: ":memory:",
		TableNames: ConfigTableNames{
			Invites: "invites",
			Users:   "users",
		},
	}
}

// Store TODO.
type Store struct {
	*observ.Corelet

	db *sqlx.DB
	*usersStore
	*invitesStore
}

// New TODO.
func New(cfg Config, corelet *observ.Corelet) (*Store, error) {
	db, err := sqlx.Connect(driverName, cfg.DBPath)
	if err != nil {
		return nil, err
	}

	invites, err := newInvitesStore(cfg.TableNames.Invites, db)
	if err != nil {
		return nil, err
	}

	users, err := newUsersStore(cfg.TableNames.Users, db)
	if err != nil {
		return nil, err
	}

	return &Store{
		Corelet: corelet,

		db:           db,
		invitesStore: invites,
		usersStore:   users,
	}, nil
}

// Close TODO.
func (s *Store) Close() error { return s.db.Close() }

// CreateUserAndDeleteInvite TODO.
func (s *Store) CreateUserAndDeleteInvite(inviteID, name, password string, mData model.UserUpdate) (model.User, error) {
	// TODO
	return model.User{}, errors.New("CreateUserAndDeleteInvite not yet implemented")
}
