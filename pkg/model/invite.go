package model

// Invite TODO
type Invite struct {
	ID      string
	OwnerID string
}

// InviteUpdate TODO
type InviteUpdate struct{ OwnerID *string }
