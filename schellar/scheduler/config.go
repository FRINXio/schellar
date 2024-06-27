package scheduler

import (
	"log"
	"os"
	"strconv"

	"github.com/frinx/schellar/ifc"
	"github.com/frinx/schellar/mongo"
	"github.com/frinx/schellar/postgres"
	"github.com/sirupsen/logrus"
)

var Configuration Config

func init() {
	log.Println("Init configuration from ENV")
	Configuration = Config{
		Db:                   dbConf(),
		CheckIntervalSeconds: intervalConf(),
		ConductorURL:         conductorUrlConf(),
		AdminRoles:           conductorAdminRolesHeadersConf(),
		AdminGroups:          conductorAdminGroupHeadersConf(),
		From:                 "schellar",
	}
}

type Config struct {
	Db                   ifc.DB
	ConductorURL         string
	CheckIntervalSeconds int
	AdminRoles           string
	AdminGroups          string
	From                 string
}

func conductorUrlConf() string {
	conductorURL := ifc.GetEnvOrDefault("CONDUCTOR_API_URL", "http://conductor-server:8080/api")
	logrus.Infof("CONDUCTOR_API_URL=%s", conductorURL)
	return conductorURL
}

func conductorAdminGroupHeadersConf() string {
	conductorAdminGroups := ifc.GetEnvOrDefault("ADMIN_GROUPS", "network-admin")
	logrus.Infof("ADMIN_GROUPS=%s", conductorAdminGroups)
	return conductorAdminGroups
}

func conductorAdminRolesHeadersConf() string {
	conductorAdminRoles := ifc.GetEnvOrDefault("ADMIN_ROLES", "OWNER")
	logrus.Infof("ADMIN_ROLES=%s", conductorAdminRoles)
	return conductorAdminRoles
}

func intervalConf() int {

	checkIntervalSecondsString := ifc.GetEnvOrDefault("CHECK_INTERVAL_SECONDS", "10")
	checkIntervalSeconds, err := strconv.Atoi(checkIntervalSecondsString)
	if err != nil {
		logrus.Fatalf("Canot parse CHECK_INTERVAL_SECONDS value '%s'. Error: %v", checkIntervalSecondsString, err)
		os.Exit(1)
	}
	logrus.Infof("CHECK_INTERVAL_SECONDS=%d", checkIntervalSeconds)
	return checkIntervalSeconds
}

func dbConf() ifc.DB {

	backend := ifc.GetEnvOrDefault("BACKEND", "postgres")

	if backend == "mongo" {
		return mongo.InitDB()
	} else if backend == "postgres" {
		return postgres.InitDB()
	} else {
		logrus.Fatalf("Cannot initialize backend '%s'", backend)
		os.Exit(1)
	}
	return nil
}
