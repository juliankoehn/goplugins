package core

import (
	"flag"
	"goplugins/core/account"
	"goplugins/core/framework"
	"goplugins/core/framework/config"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Bootstrap starts our framework
func Bootstrap() {
	var envfile string
	flag.StringVar(&envfile, "env-file", ".env", "Read in a file of environment variables")
	flag.Parse()
	godotenv.Load(envfile)

	config, err := config.Environ()
	if err != nil {
		logger := logrus.WithError(err)
		logger.Fatalln("main: invalid configuration")
	}

	fw := framework.New(config)

	fw.AddService(account.NewService)

	fw.Start()
}
