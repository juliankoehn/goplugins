package handler

import (
	"goplugins/core/account/models"
	"goplugins/core/routing"
	"net/http"
)

type createParams struct {
	Email     string `json:"email" validate:"empty=false & format=email"`
	FirstName string `json:"firstName" validate:"empty=false & gte=2 & lte=25"`
	LastName  string `json:"lastName"  validate:"empty=false & gte=2 & lte=25"`
	Password  string `json:"password" validate:"empty=false & gte=6"`
}

// CreateHandler returns a routing.HandlerFunc to create a new User
func CreateHandler(userStore models.UserStore) routing.HandlerFunc {
	return func(c routing.Context) error {
		req := new(createParams)
		if err := c.Bind(req); err != nil {
			return c.String(http.StatusBadRequest, "invalid params")
		}

		if err := c.Validate(req); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		user := &models.User{
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
		}

		if err := userStore.Create(user, req.Password); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(200, user)
	}
}
