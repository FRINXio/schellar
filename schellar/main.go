package main

import (
	"flag"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

var (
	conductorURL         string
	mongoAddress         string
	mongoUsername        string
	mongoPassword        string
	dbName               = "admin"
	checkIntervalSeconds = 10
)

//Schedule struct data
type Schedule struct {
	Name                string                 `json:"name,omitempty" bson:"name"`
	Enabled             bool                   `json:"enabled,omitempty" bson:"enabled"`
	Status              string                 `json:"status,omitempty" bson:"status"`
	WorkflowName        string                 `json:"workflowName,omitempty" bson:"workflowName"`
	WorkflowVersion     string                 `json:"workflowVersion,omitempty" bson:"workflowVersion"`
	WorkflowContext     map[string]interface{} `json:"workflowContext,omitempty" bson:"workflowContext"`
	CronString          string                 `json:"cronString,omitempty" bson:"cronString"`
	ParallelRuns        bool                   `json:"parallelRuns,omitempty" bson:"parallelRuns"`
	CheckWarningSeconds int                    `json:"checkWarningSeconds,omitempty" bson:"checkWarningSeconds"`
	FromDate            *time.Time             `json:"fromDate,omitempty" bson:"fromDate"`
	ToDate              *time.Time             `json:"toDate,omitempty" bson:"toDate"`
	LastUpdate          time.Time              `json:"lastUpdate,omitempty" bson:"lastUpdate"`
	CorrelationID       string                 `json:"correlationId,omitempty" bson:"correlationId"`
	TaskToDomain        map[string]string      `json:"taskToDomain,omitempty" bson:"taskToDomain"`
}

func (schedule Schedule) ValidateAndUpdate() error {
	if schedule.Name == "" {
		return errors.New("'name' is required")
	}
	if strings.Contains(schedule.Name, "/") {
		return errors.New("'name' cannot contain '/' character")
	}
	if schedule.WorkflowName == "" {
		return errors.New("'workflowName' is required")
	}
	if schedule.CronString == "" {
		return errors.New("'cronString' is required")
	}
	_, err := cron.ParseStandard(schedule.CronString)
	if err != nil {
		return errors.Wrap(err, "'cronString' is invalid")
	}
	if schedule.WorkflowVersion == "" {
		schedule.WorkflowVersion = "1"
	}
	if schedule.CheckWarningSeconds == 0 {
		schedule.CheckWarningSeconds = 3600
	}
	schedule.LastUpdate = time.Now()
	return nil
}

func main() {
	logLevel := flag.String("loglevel", "debug", "debug, info, warning, error")
	checkInterval0 := flag.Int("check-interval", 10, "Workflow check interval in seconds")
	conductorURL0 := flag.String("conductor-api-url", "", "Conductor API URL. Example: http://conductor-server:8080/api")
	mongoAddress0 := flag.String("mongo-address", "", "MongoDB address. Example: 'mongo', or 'mongdb://mongo1:1234/db1,mongo2:1234/db1")
	mongoUsername0 := flag.String("mongo-username", "root", "MongoDB username")
	mongoPassword0 := flag.String("mongo-password", "root", "MongoDB password")
	flag.Parse()

	switch *logLevel {
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

	conductorURL = *conductorURL0
	if conductorURL == "" {
		logrus.Errorf("'conductor-api-url' parameter is required")
		os.Exit(1)
	}

	mongoAddress = *mongoAddress0
	if mongoAddress == "" {
		logrus.Errorf("'mongo-address' parameter is required")
		os.Exit(1)
	}

	mongoUsername = *mongoUsername0
	mongoPassword = *mongoPassword0

	checkIntervalSeconds = *checkInterval0

	logrus.Info("====Starting Schellar====")

	err := initMongo()
	if err != nil {
		logrus.Fatalf("Couldn't init database: %v", err)
	}

	err = startScheduler()
	if err != nil {
		logrus.Fatalf("Error during scheduler startup: %v", err)
	}
	startRestAPI()
}
