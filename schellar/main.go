package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type DB interface {
	FindAll() ([]Schedule, error)
	FindAllByEnabled(enabled bool) ([]Schedule, error)
	FindByName(scheduleName string) (*Schedule, error)
	FindByStatus(status string) ([]Schedule, error)
	UpdateStatus(scheduleName string, scheduleStatus string) error
	UpdateStatusAndWorkflowContext(schedule Schedule) error
	Insert(schedule Schedule) error
	Update(schedule Schedule) error
	RemoveByName(scheduleName string) error
}

type DBFactory interface {
	InitDB() DB
}

var (
	db                   DB
	conductorURL         string
	checkIntervalSeconds int
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

func GetEnvOrDefault(key string, defaultValue string) string {
	found, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return found
}

func main() {
	logLevel := GetEnvOrDefault("LOG_LEVEL", "INFO")
	var err error
	checkIntervalSecondsString := GetEnvOrDefault("CHECK_INTERVAL_SECONDS", "10")
	checkIntervalSeconds, err = strconv.Atoi(checkIntervalSecondsString)
	if err != nil {
		logrus.Fatalf("Canot parse CHECK_INTERVAL_SECONDS value '%s'. Error: %v", checkIntervalSecondsString, err)
	}
	conductorURL = GetEnvOrDefault("CONDUCTOR_API_URL", "http://conductor-server:8080/api")

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

	db = InitDB()

	err = startScheduler()
	if err != nil {
		logrus.Fatalf("Error during scheduler startup: %v", err)
	}
	startRestAPI()
}
