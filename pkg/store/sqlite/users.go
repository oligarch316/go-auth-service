package sqlite

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/oligarch316/go-auth-service/pkg/model"
)

type user struct {
	ID           uint64  `db:"id"`
	Name         string  `db:"name"`
	DisplayName  *string `db:"display_name"`
	PasswordHash []byte  `db:"password_hash"`
	Admin        bool    `db:"admin"`
}

func (u *user) setID(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	u.ID = uint64(idInt)
	return err
}

func (u *user) setPasswordHash(password string) error {
	ph := make(model.PasswordHash, 0)
	err := ph.Set(password)
	u.PasswordHash = ph
	return err
}

func (u user) toModel() model.User {
	res := model.User{
		ID:           strconv.FormatInt(int64(u.ID), 10),
		Name:         u.Name,
		PasswordHash: u.PasswordHash,
		Admin:        u.Admin,
	}

	if u.DisplayName != nil {
		res.DisplayName = *u.DisplayName
	}

	return res
}

type usersStore struct {
	db        *sqlx.DB
	tableName string

	createStmt, readStmt, deleteStmt, lookupStmt *sqlx.NamedStmt
}

func newUsersStore(tableName string, db *sqlx.DB) (*usersStore, error) {
	createStmt, err := db.PrepareNamed("INSERT INTO " + tableName + " (name, display_name, password_hash, admin) VALUES (:name, :display_name, :password_hash, :admin)")
	if err != nil {
		return nil, err
	}

	readStmt, err := db.PrepareNamed("SELECT * FROM " + tableName + " WHERE id=:id")
	if err != nil {
		return nil, err
	}

	deleteStmt, err := db.PrepareNamed("DELETE FROM " + tableName + " WHERE id=:id")
	if err != nil {
		return nil, err
	}

	lookupStmt, err := db.PrepareNamed("SELECT * FROM " + tableName + " WHERE name=:name")
	if err != nil {
		return nil, err
	}

	return &usersStore{
		db:        db,
		tableName: tableName,

		createStmt: createStmt,
		readStmt:   readStmt,
		deleteStmt: deleteStmt,
		lookupStmt: lookupStmt,
	}, nil
}

func (us *usersStore) CreateUser(name, password string, mData model.UserUpdate) (model.User, error) {
	u := user{
		Name:        name,
		DisplayName: mData.DisplayName,
	}

	if err := u.setPasswordHash(password); err != nil {
		return model.User{}, err
	}

	if mData.Admin != nil {
		u.Admin = *mData.Admin
	}

	result, err := us.createStmt.Exec(u)
	if err != nil {
		return model.User{}, err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return model.User{}, err
	}

	u.ID = uint64(newID)

	return u.toModel(), nil
}

func (us *usersStore) ReadUser(id string) (res model.User, err error) {
	var u user

	if err = u.setID(id); err != nil {
		return
	}

	if err = us.readStmt.Get(&u, u); err != nil {
		return
	}

	return u.toModel(), nil
}

func (us *usersStore) UpdateUser(id string, mData model.UserUpdate) error {
	var u user

	if err := u.setID(id); err != nil {
		return err
	}

	var setItems []string

	if mData.Password != nil {
		if err := u.setPasswordHash(*mData.Password); err != nil {
			return err
		}
		setItems = append(setItems, "password_hash=:password_hash")
	}

	if mData.DisplayName != nil {
		u.DisplayName = mData.DisplayName
		setItems = append(setItems, "display_name=:display_name")
	}

	if mData.Admin != nil {
		u.Admin = *mData.Admin
		setItems = append(setItems, "admin=:admin")
	}

	if len(setItems) < 1 {
		return nil
	}

	qryStr := fmt.Sprintf("UPDATE %s SET %s WHERE id=:id", us.tableName, strings.Join(setItems, ","))
	_, err := us.db.NamedExec(qryStr, u)

	return err
}

func (us *usersStore) DeleteUser(id string) error {
	var u user

	if err := u.setID(id); err != nil {
		return err
	}

	_, err := us.deleteStmt.Exec(u)
	return err
}

func (us *usersStore) LookupUser(name string) (model.User, error) {
	u := user{Name: name}

	if err := us.lookupStmt.Get(&u, u); err != nil {
		return model.User{}, err
	}

	return u.toModel(), nil
}
