package ifc

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

//Schedule struct data
type Schedule struct {
	Name                string                 `json:"name,omitempty" bson:"name"`
	Enabled             bool                   `json:"enabled,omitempty" bson:"enabled"`
	Status              string                 `json:"status,omitempty" bson:"status"`
	WorkflowName        string                 `json:"workflowName,omitempty" bson:"workflowName"`
	WorkflowVersion     int                    `json:"workflowVersion,omitempty" bson:"workflowVersion"`
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
	if schedule.CheckWarningSeconds == 0 {
		schedule.CheckWarningSeconds = 3600
	}
	schedule.LastUpdate = time.Now()
	return nil
}

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
