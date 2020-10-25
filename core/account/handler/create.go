package handler

import (
	"goplugins/core/account/models"
	"goplugins/core/routing"
)

// Create allows to create a new User
func Create(userStore models.UserStore) routing.HandlerFunc {
	return func(c routing.Context) error {
		return c.String(200, "Create User Endpoint")
	}
}
