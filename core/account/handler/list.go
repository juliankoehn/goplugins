package handler

import (
	"goplugins/core/account/models"
	"goplugins/core/routing"
)

// ListAccounts lists all users
func ListAccounts(userStore models.UserStore) routing.HandlerFunc {
	return func(c routing.Context) error {
		return c.String(200, "listUsers")
	}
}
