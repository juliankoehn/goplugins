package account

import (
	"goplugins/core/account/handler"
	"goplugins/core/account/models"
	"goplugins/core/framework/database"
	"goplugins/core/routing"
)

// service holds the Service
type service struct {
	db        *database.DB
	userStore *models.UserStore
}

// NewService creates a new Account Service
func NewService(
	db *database.DB,
	mux *routing.Mux,
) {
	mux.GET("/users", handler.ListAccounts(db))
	mux.POST("/user", handler.Create(db))
}
