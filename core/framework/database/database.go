package database

import "gorm.io/gorm"

// Database driver enums.
const (
	Sqlite = iota + 1
	Mysql
	Postgres
)

type (
	// DB is a pool of zero or more underlying connections
	// to the database
	DB struct {
		*gorm.DB
		driver Driver
	}
	// Driver defines the database Driver
	Driver int
)

// Driver returns the name of the SQL driver.
func (db *DB) Driver() Driver {
	return db.driver
}
