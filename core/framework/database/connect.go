package database

import (
	"time"

	msql "gorm.io/driver/mysql"
	pgsql "gorm.io/driver/postgres"
	lite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Connect opens a new persistent connection to a database
func Connect(driver, dataSource string, maxOpens int) (*DB, error) {
	var db *gorm.DB
	var err error
	var engine Driver

	switch driver {
	case "mysql":
		db, err = gorm.Open(msql.Open(dataSource), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		engine = Mysql
	case "postgres":
		db, err = gorm.Open(pgsql.Open(dataSource), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		engine = Postgres
		sqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}
		// higher MaxOpenConns and MaxIdleConns values will lead to better performance.
		// But the returns are diminishing, and you should be aware that having a
		// too-large idle connection pool (with connections that are not re-used and
		// eventually go bad) can actually lead to reduced performance.
		//
		// To mitigate the risk from point above, you may want to set a relatively
		// short ConnMaxLifetime. But you don't want this to be so short that leads
		// to connections being killed and recreated unnecessarily often.
		// MaxIdleConns should always be less than or equal to MaxOpenConns.

		// Set the maximum number of concurrently idle connections to 25. Setting this
		// to less than or equal to 0 will mean that no idle connections are retained.
		sqlDB.SetMaxIdleConns(maxOpens)
		// Set the maximum lifetime of a connection to 5 minutes. Setting it to 0
		// means that there is no maximum lifetime and the connection is reused
		// forever (which is the default behavior).
		//
		// This isn't an idle timeout. The connection will expire 5 minutes after it
		// was first created — not 5 minutes after it last became idle.
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		// postgresql.conf the default is 100
		// Set the number of open connections (in-use + idle)
		sqlDB.SetMaxOpenConns(maxOpens)
	default:
		db, err = gorm.Open(lite.Open(dataSource), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		engine = Sqlite

	}

	return &DB{
		db,
		engine,
	}, nil
}
