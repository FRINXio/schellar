package graph

import (
	"encoding/json"
	"time"

	"github.com/frinx/schellar/graph/model"
	"github.com/frinx/schellar/ifc"
)

func ConvertIfcToModel(schedule *ifc.Schedule) *model.Schedule {

	schedule_model := &model.Schedule{
		Name:            schedule.Name,
		Enabled:         schedule.Enabled,
		ParallelRuns:    schedule.ParallelRuns,
		WorkflowName:    schedule.WorkflowName,
		WorkflowVersion: schedule.WorkflowVersion,
		CronString:      schedule.CronString,
		Status:          schedule.Status,
		WorkflowContext: "",
		FromDate:        "",
		ToDate:          "",
	}

	if schedule.WorkflowContext != nil {
		contextBytes, _ := json.Marshal(schedule.WorkflowContext)
		schedule_model.WorkflowContext = string(contextBytes)
	}

	if schedule.FromDate != nil {
		schedule_model.FromDate = schedule.FromDate.Format(time.RFC3339)
	}

	if schedule.FromDate != nil {
		schedule_model.FromDate = schedule.FromDate.Format(time.RFC3339)
	}

	return schedule_model
}
