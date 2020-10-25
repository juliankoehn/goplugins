package framework

import (
	"goplugins/core/framework/config"
	"goplugins/core/framework/database"
	"goplugins/core/routing"

	"github.com/sirupsen/logrus"
)

type (
	// Framework holds all kind of information
	// about our Framework
	// it manages routes, handels plugins and so on...
	Framework struct {
		config *config.Config
		db     *database.DB
		mux    *routing.Mux
	}
)

// New returns a new Framework Application Instance
func New(config config.Config) *Framework {
	db, err := database.Connect(config.Database.Driver, config.Database.Datasource, config.Database.MaxConnections)
	if err != nil {
		logger := logrus.WithError(err)
		logger.Fatalln("framework: could not connect to database")
	}

	mux := routing.New()

	files, err := ListAvailablePlugins()
	if err != nil {
		logger := logrus.WithError(err)
		logger.Fatalln("framework: could not list plugins")
	}

	for _, v := range files {
		err := InitializePlugin(v, mux)
		if err != nil {
			logger := logrus.WithError(err)
			logger.Fatalln("framework: could not initialize plugins")
		}
	}

	f := &Framework{
		config: &config,
		db:     db,
		mux:    mux,
	}

	return f
}

// AddService allows to register a new Service to our Framework
func (f *Framework) AddService(fn func(*database.DB, *routing.Mux)) {
	fn(f.db, f.mux)
}

// Start starts the framework service
func (f *Framework) Start() {
	f.mux.Logger.Fatal(f.mux.Start(":3000"))
}
