package store

import (
	"goplugins/core/account/errs"
	"goplugins/core/account/models"
	"goplugins/core/framework/database"

	"errors"

	"github.com/google/uuid"
)

// New returns a new UserStore.
func New(db *database.DB) models.UserStore {
	return &userStore{db}
}

type userStore struct {
	db *database.DB
}

func (u *userStore) Create(user *models.User, password string) error {
	return u.db.Create(user).Error
}

// SetEmail updates the email of the given user
func (u *userStore) SetEmail(user *models.User, email string) error {
	user.Email = email
	return u.db.Model(user).UpdateColumn("email", email).Error
}

func (u *userStore) findUser(query string, args ...interface{}) (*models.User, error) {
	obj := &models.User{}
	if err := u.db.Where(query, args...).First(obj).Error; err != nil {
		if errors.Is(err, database.ErrRecordNotFound) {
			return nil, errs.ErrUserNotFound
		}
	}

	return obj, nil
}

func (u *userStore) Find(id uuid.UUID) (*models.User, error) {
	return u.findUser("id = ?", id)
}

func (u *userStore) FindByConfirmationToken(token string) (*models.User, error) {
	return u.findUser("confirmation_token = ?", token)
}

func (u *userStore) FindByEmail(email string) (*models.User, error) {
	return u.findUser("email = ?", email)
}

func (u *userStore) FindByRecoveryToken(token string) (*models.User, error) {
	return u.findUser("recovery_token = ?", token)
}
