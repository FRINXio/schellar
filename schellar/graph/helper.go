package graph

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/frinx/schellar/graph/model"
	"github.com/frinx/schellar/ifc"
)

func ConvertIfcToModel(schedule_ifc *ifc.Schedule) *model.Schedule {

	schedule_model := &model.Schedule{
		Name:            schedule_ifc.Name,
		Enabled:         schedule_ifc.Enabled,
		ParallelRuns:    schedule_ifc.ParallelRuns,
		WorkflowName:    schedule_ifc.WorkflowName,
		WorkflowVersion: schedule_ifc.WorkflowVersion,
		CronString:      schedule_ifc.CronString,
		Status:          schedule_ifc.Status,
		WorkflowContext: "",
		FromDate:        "",
		ToDate:          "",
	}

	if schedule_ifc.WorkflowContext != nil {
		contextBytes, _ := json.Marshal(schedule_ifc.WorkflowContext)
		schedule_model.WorkflowContext = string(contextBytes)
	}

	if schedule_ifc.FromDate != nil {
		schedule_model.FromDate = schedule_ifc.FromDate.Format(time.RFC3339)
	}

	if schedule_ifc.FromDate != nil {
		schedule_model.FromDate = schedule_ifc.FromDate.Format(time.RFC3339)
	}

	return schedule_model
}

func ConvertWorkflowContext(modelWorkflowContext string) (map[string]interface{}, error) {

	var workflowContext map[string]interface{}

	if modelWorkflowContext != "" {
		json.Unmarshal([]byte(modelWorkflowContext), &workflowContext)
	}
	return workflowContext, nil
}

func ConvertDateTime(modelDateTime string) (time.Time, error) {

	dateFrom, err := time.Parse(time.RFC3339, modelDateTime)

	if err != nil {
		fmt.Println("Error while parsing the date time :", err)
	}

	return dateFrom, err
}
