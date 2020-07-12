package model

import "golang.org/x/crypto/bcrypt"

const hashCost = bcrypt.DefaultCost

// PasswordHash TODO
type PasswordHash []byte

// Compare TODO
func (p PasswordHash) Compare(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(p), []byte(password))
}

// Set TODO
func (p *PasswordHash) Set(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	*p = PasswordHash(hash)
	return err
}

// User TODO
type User struct {
	ID           string
	Name         string
	DisplayName  string
	PasswordHash PasswordHash
	Admin        bool
}

// UserUpdate TODO
type UserUpdate struct {
	DisplayName *string
	Password    *string
	Admin       *bool
}
