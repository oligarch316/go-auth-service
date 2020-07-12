package sqlite

import (
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/oligarch316/go-auth-service/pkg/model"
)

type invite struct {
	ID      uint64 `db:"id"`
	OwnerID uint64 `db:"owner_id"`
}

func (i *invite) setID(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	i.ID = uint64(idInt)
	return err
}

func (i *invite) setOwnerID(id string) error {
	idInt, err := strconv.ParseInt(id, 10, 64)
	i.OwnerID = uint64(idInt)
	return err
}

func (i invite) toModel() model.Invite {
	return model.Invite{
		ID:      strconv.FormatInt(int64(i.ID), 10),
		OwnerID: strconv.FormatInt(int64(i.OwnerID), 10),
	}
}

type invitesStore struct {
	createStmt, readStmt, updateStmt, deleteStmt, lookupStmt *sqlx.NamedStmt
}

func newInvitesStore(tableName string, db *sqlx.DB) (*invitesStore, error) {
	createStmt, err := db.PrepareNamed("INSERT INTO " + tableName + " (owner_id) VALUES (:owner_id)")
	if err != nil {
		return nil, err
	}

	readStmt, err := db.PrepareNamed("SELECT owner_id FROM " + tableName + " WHERE id=:id")
	if err != nil {
		return nil, err
	}

	updateStmt, err := db.PrepareNamed("UPDATE " + tableName + " SET owner_id=:owner_id WHERE id=:id")
	if err != nil {
		return nil, err
	}

	deleteStmt, err := db.PrepareNamed("DELETE FROM " + tableName + " WHERE id=:id")
	if err != nil {
		return nil, err
	}

	lookupStmt, err := db.PrepareNamed("SELECT * FROM " + tableName + " WHERE owner_id=:owner_id")
	if err != nil {
		return nil, err
	}

	return &invitesStore{
		createStmt: createStmt,
		readStmt:   readStmt,
		updateStmt: updateStmt,
		deleteStmt: deleteStmt,
		lookupStmt: lookupStmt,
	}, nil
}

func (is *invitesStore) CreateInvites(ownerID string, count int) ([]model.Invite, error) {
	var (
		inv invite
		res []model.Invite
	)

	if err := inv.setOwnerID(ownerID); err != nil {
		return res, err
	}

	for i := 0; i < count; i++ {
		result, err := is.createStmt.Exec(inv)
		if err != nil {
			return res, err
		}

		newID, err := result.LastInsertId()
		if err != nil {
			return res, err
		}

		inv.ID = uint64(newID)
		res = append(res, inv.toModel())
	}

	return res, nil
}

func (is *invitesStore) ReadInvite(id string) (res model.Invite, err error) {
	var inv invite

	if err = inv.setID(id); err != nil {
		return
	}

	if err = is.readStmt.Get(&inv, inv); err != nil {
		return
	}

	return inv.toModel(), nil
}

func (is *invitesStore) UpdateInvite(id string, mData model.InviteUpdate) error {
	if mData.OwnerID == nil {
		return nil
	}

	var inv invite

	if err := inv.setID(id); err != nil {
		return err
	}

	if err := inv.setOwnerID(*mData.OwnerID); err != nil {
		return err
	}

	_, err := is.updateStmt.Exec(inv)
	return err
}

func (is *invitesStore) DeleteInvite(id string) error {
	var inv invite

	if err := inv.setID(id); err != nil {
		return err
	}

	_, err := is.deleteStmt.Exec(inv)
	return err
}

func (is *invitesStore) LookupInvites(ownerID string) ([]model.Invite, error) {
	var (
		inv     invite
		invList = make([]invite, 0)
	)

	if err := inv.setOwnerID(ownerID); err != nil {
		return nil, err
	}

	if err := is.lookupStmt.Select(&invList, inv); err != nil {
		return nil, err
	}

	res := make([]model.Invite, len(invList))
	for i, item := range invList {
		res[i] = item.toModel()
	}

	return res, nil
}
