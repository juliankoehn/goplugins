package account

import (
	"goplugins/core/account/handler"
	"goplugins/core/account/models"
	"goplugins/core/account/store"
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
	userStore := store.New(db)

	db.AutoMigrate(
		models.Group{},
		models.Permission{},
		models.User{},
	)

	mux.GET("/users", handler.ListAccounts(userStore))
	mux.POST("/user", handler.CreateHandler(userStore))
}
