package store

import (
	"goplugins/core/account/models"
	"goplugins/core/framework/database"
)

// New returns a new UserStore.
func New(db *database.DB) models.UserStore {
	return &userStore{db}
}

type userStore struct {
	db *database.DB
}
