package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/frinx/schellar/graph/model"
	"github.com/frinx/schellar/ifc"
	"github.com/frinx/schellar/utils"

	"github.com/frinx/schellar/scheduler"

	"github.com/99designs/gqlgen/graphql"
)

func ValidateName(name string) error {
	if name == "" {
		return errors.New("'name' is required")
	}
	if strings.Contains(name, "/") {
		return errors.New("'name' cannot contain '/' character")
	}
	return nil
}

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

func extractAuthHeader(ctx context.Context) []string {

	// Extract auth headers from request
	headers := graphql.GetOperationContext(ctx).Headers
	rbacHeaders := []string{headers.Get("x-auth-user-roles"), headers.Get("x-auth-user-groups")}

	// Filter out empty strings
	var nonEmptyHeaders []string
	for _, header := range rbacHeaders {
		if header != "" {
			nonEmptyHeaders = append(nonEmptyHeaders, header)
		}
	}

	// Join non-empty headers with a comma
	userHeaderList := strings.Split(strings.Join(nonEmptyHeaders, ","), ",")
	return utils.RemoveDuplicates(userHeaderList)
}

func extractUserHeader(ctx context.Context) error {

	// Extract auth headers from request
	headers := graphql.GetOperationContext(ctx).Headers
	userHeader := headers.Get("from")

	if userHeader == "" {
		return errors.New("Missing header From")
	}
	return nil
}

func getAdminValues() []string {

	adminRoles := []string{os.Getenv("ADMIN_ROLES"), os.Getenv("ADMIN_GROUPS")}

	// Filter out empty strings
	var nonEmptyRoles []string
	for _, header := range adminRoles {
		if header != "" {
			nonEmptyRoles = append(nonEmptyRoles, header)
		}
	}

	// Join non-empty headers with a comma
	adminRolesList := strings.Split(strings.Join(nonEmptyRoles, ","), ",")
	return utils.RemoveDuplicates(adminRolesList)
}

// hasCommonElement checks if at least one string from list1 exists in list2
func hasCommonElement(list1, list2 []string) bool {
	// Create a map to store adminHeaders for O(1) average time complexity lookups
	adminHeaderMap := make(map[string]bool)
	for _, header := range list2 {
		adminHeaderMap[header] = true
	}

	// Check each userHeader if it exists in adminHeaderMap
	for _, header := range list1 {
		if adminHeaderMap[header] {
			return true
		}
	}

	return false
}

func checkPermissions(ctx context.Context) error {

	userHaders := extractAuthHeader(ctx)
	adminRoles := utils.GetAdminValues()

	// Check if at least one userHeader exists in adminHeaders
	if hasCommonElement(userHaders, adminRoles) {
		return nil
	} else {
		return errors.New("User has no permission to process operation")
	}

}
