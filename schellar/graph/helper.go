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

func StringToStatusType(s string) model.Status {
	for _, status := range model.AllStatus {
		if string(status) == s {
			return status
		}
	}
	return model.StatusUnknown
}

func ConvertIfcToModel(schedule_ifc *ifc.Schedule) *model.Schedule {

	schedule_model := &model.Schedule{
		Name:            schedule_ifc.Name,
		Enabled:         schedule_ifc.Enabled,
		ParallelRuns:    schedule_ifc.ParallelRuns,
		WorkflowName:    schedule_ifc.WorkflowName,
		WorkflowVersion: schedule_ifc.WorkflowVersion,
		CronString:      schedule_ifc.CronString,
		Status:          StringToStatusType(schedule_ifc.Status),
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
		schedule_model.ToDate = schedule_ifc.ToDate.Format(time.RFC3339)
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

func getSchedulesFirst(schedules []*model.Schedule, first *int) ([]*model.Schedule, bool, bool) {

	hasNextPage := true
	allSchedules := len(schedules)
	strip := *first

	if allSchedules <= strip {
		hasNextPage = false
		strip = allSchedules
	}

	filteredSchedules := schedules[0:strip]
	return filteredSchedules, false, hasNextPage
}

func getSchedulesLast(schedules []*model.Schedule, last *int) ([]*model.Schedule, bool, bool) {

	hasPreviousPage := true
	allSchedules := len(schedules)
	strip := *last

	if allSchedules <= strip {
		hasPreviousPage = false
		strip = allSchedules
	}

	filteredSchedules := schedules[allSchedules-strip : allSchedules]
	return filteredSchedules, hasPreviousPage, false
}

func getSchedulesFirstAfter(schedules []*model.Schedule, first *int, after *string) ([]*model.Schedule, bool, bool) {

	var hasPreviousPage bool = false
	var hasNextPage bool = false

	filteredSchedules, position := getScheduleAfterCursor(*after, schedules)
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

	return filteredSchedules, hasPreviousPage, hasNextPage
}

func getSchedulesLastBefore(schedules []*model.Schedule, last *int, before *string) ([]*model.Schedule, bool, bool) {

	var hasPreviousPage bool = false
	var hasNextPage bool = false

	filteredCountBefore := len(schedules)
	filteredSchedules, position := getScheduleBeforeCursor(*before, schedules)
	filteredCount := len(filteredSchedules)

	startIndex := filteredCount - *last
	if startIndex < 0 {
		startIndex = 0
	}

	filteredSchedules = filteredSchedules[startIndex:filteredCount]
	filteredCount = len(filteredSchedules)

	prevousScheduls := position - filteredCount
	nextScheduls := filteredCountBefore - (position + 1)

	if prevousScheduls > 0 {
		hasPreviousPage = true
	}
	if nextScheduls > 0 {
		hasNextPage = true
	}

	return filteredSchedules, hasPreviousPage, hasNextPage
}
