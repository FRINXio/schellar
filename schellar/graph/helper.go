package graph

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/frinx/schellar/graph/model"
	"github.com/frinx/schellar/ifc"
	"github.com/frinx/schellar/scheduler"
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

func GetSchedules() []*model.Schedule {
	ifc_schedules, _ := scheduler.Configuration.Db.FindAll()
	model_schedules := make([]*model.Schedule, len(ifc_schedules))
	for i, v := range ifc_schedules {
		model_schedules[i] = ConvertIfcToModel(&v)
	}
	// sort.Slice(model_schedules, func(i, j int) bool {
	// 	return model_schedules[i].Name < model_schedules[j].Name
	// })
	return model_schedules
}

func GetSchedulesFilter(filter *model.SchedulesFilterInput) []*model.Schedule {
	ifc_schedules, _ := scheduler.Configuration.Db.FindAllByWorkflowType(filter.WorkflowName, filter.WorkflowVersion)
	model_schedules := make([]*model.Schedule, len(ifc_schedules))
	for i, v := range ifc_schedules {
		model_schedules[i] = ConvertIfcToModel(&v)
	}
	// sort.Slice(model_schedules, func(i, j int) bool {
	// 	return model_schedules[i].Name < model_schedules[j].Name
	// })
	return model_schedules
}

func FilterSchedules(filter *model.SchedulesFilterInput, schedules []*model.Schedule) []*model.Schedule {

	if filter != nil {
		filteredSchedules := make([]*model.Schedule, 0)

		for _, schedule := range schedules {
			if schedule.WorkflowName == filter.WorkflowName && schedule.WorkflowVersion == filter.WorkflowVersion {
				filteredSchedules = append(filteredSchedules, schedule)
			}
		}

		return filteredSchedules
	}
	return schedules
}

func getScheduleAfterCursor(cursor string, schedules []*model.Schedule) []*model.Schedule {
	for i, schedule := range schedules {
		if schedule.Name == cursor {
			return schedules[i+1 : len(schedules)]
		}
	}
	return nil
}

func getScheduleBeforeCursor(cursor string, schedules []*model.Schedule) []*model.Schedule {
	for i, schedule := range schedules {
		if schedule.Name == cursor {
			return schedules[0:i]
		}
	}
	return nil
}
