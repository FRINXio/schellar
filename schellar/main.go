package main

import (
	"strconv"
	"strings"

	"github.com/frinx/schellar/ifc"
	"github.com/frinx/schellar/mongo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

var (
	db                   ifc.DB
	conductorURL         string
	checkIntervalSeconds int
)

func main() {
	logLevel := ifc.GetEnvOrDefault("LOG_LEVEL", "INFO")
	var err error
	checkIntervalSecondsString := ifc.GetEnvOrDefault("CHECK_INTERVAL_SECONDS", "10")
	checkIntervalSeconds, err = strconv.Atoi(checkIntervalSecondsString)
	if err != nil {
		logrus.Fatalf("Canot parse CHECK_INTERVAL_SECONDS value '%s'. Error: %v", checkIntervalSecondsString, err)
	}
	conductorURL = ifc.GetEnvOrDefault("CONDUCTOR_API_URL", "http://conductor-server:8080/api")

	switch strings.ToLower(logLevel) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Info("Starting Schellar with CONDUCTOR_API_URL=%s, CHECK_INTERVAL_SECONDS=%d", conductorURL, checkIntervalSeconds)

	backend := ifc.GetEnvOrDefault("BACKEND", "mongo")
	if backend == "mongo" {
		db = mongo.InitDB()
	} else {
		logrus.Fatalf("Cannot initialize backend '%s'", backend)
	}

	err = startScheduler()
	if err != nil {
		logrus.Fatalf("Error during scheduler startup: %v", err)
	}
	startRestAPI()
}
