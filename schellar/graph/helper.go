package graph

import (
	"encoding/json"
	"errors"
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

	if schedule_ifc.ToDate != nil {
		schedule_model.FromDate = schedule_ifc.ToDate.Format(time.RFC3339)
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
	return model_schedules
}

func GetSchedulesFilter(filter *model.SchedulesFilterInput) []*model.Schedule {
	ifc_schedules, _ := scheduler.Configuration.Db.FindAllByWorkflowType(filter.WorkflowName, filter.WorkflowVersion)
	model_schedules := make([]*model.Schedule, len(ifc_schedules))
	for i, v := range ifc_schedules {
		model_schedules[i] = ConvertIfcToModel(&v)
	}
	return model_schedules
}

func getScheduleAfterCursor(cursor string, schedules []*model.Schedule) ([]*model.Schedule, int) {
	for i, schedule := range schedules {
		if schedule.Name == cursor {
			return schedules[i+1 : len(schedules)], i
		}
	}
	return nil, 0
}

func getScheduleBeforeCursor(cursor string, schedules []*model.Schedule) ([]*model.Schedule, int) {
	for i, schedule := range schedules {
		if schedule.Name == cursor {
			return schedules[0:i], i
		}
	}
	return nil, 0
}

func handlePagination(after *string, before *string, first *int, last *int) error {

	if first != nil && last != nil {
		return errors.New("Only one of 'first' and 'last' can be set")
	}

	if before != nil && after != nil {
		return errors.New("Only one of 'after' and 'before' can be set")
	}

	if after != nil && first == nil {
		return errors.New("'after' needs to be used with 'first'")
	}

	if before != nil && last == nil {
		return errors.New("'before' needs to be used with 'last'")
	}

	if first != nil && *first <= 0 {
		return errors.New("'first' has to be positive")
	}

	if last != nil && *last <= 0 {
		return errors.New("'last' has to be positive")
	}

	return nil
}

func getSchedulesWithFilter(after *string, before *string, schedules []*model.Schedule, first *int, last *int) ([]*model.Schedule, bool, bool) {
	var filteredSchedules []*model.Schedule

	var hasPreviousPage bool = false
	var hasNextPage bool = false
	var cursor string
	var position int

	if after != nil {
		cursor = *after
		filteredSchedules, position = getScheduleAfterCursor(cursor, schedules)
		filteredCount := len(filteredSchedules)

		endIndex := *first
		if endIndex > filteredCount {
			endIndex = filteredCount
		}
		filteredSchedules = filteredSchedules[0:endIndex]
		if position > 0 {
			hasPreviousPage = true
		}
		if endIndex < filteredCount {
			hasNextPage = true
		}

	} else if before != nil {
		cursor = *before

		filteredCountBefore := len(schedules)
		filteredSchedules, position = getScheduleBeforeCursor(cursor, schedules)
		filteredCount := len(filteredSchedules)

		startIndex := filteredCount - *last
		if startIndex < 0 {
			startIndex = 0
		}

		filteredSchedules = filteredSchedules[startIndex:filteredCount]
		filteredCount = len(filteredSchedules)

		prevousScheduls := position - filteredCount
		nextScheduls := filteredCountBefore - position + 1

		if prevousScheduls > 0 {
			hasPreviousPage = true
		}
		if nextScheduls > 0 {
			hasNextPage = true
		}

	} else {
		return schedules, hasPreviousPage, hasNextPage
	}

	return filteredSchedules, hasPreviousPage, hasNextPage
}
