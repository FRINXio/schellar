package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.30

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/frinx/schellar/graph/model"
	"github.com/frinx/schellar/ifc"
	"github.com/frinx/schellar/scheduler"
	"github.com/sirupsen/logrus"
)

// CreateSchedule is the resolver for the createSchedule field.
func (r *mutationResolver) CreateSchedule(ctx context.Context, input model.CreateScheduleInput) (*model.Schedule, error) {

	var workflowContext map[string]interface{}
	json.Unmarshal([]byte(input.WorkflowContext), &workflowContext)

	dateFrom, err := time.Parse(time.RFC3339, input.FromDate)
	if err != nil {
		fmt.Println("Error while parsing the date time :", err)
	}

	dateTo, err := time.Parse(time.RFC3339, input.ToDate)
	if err != nil {
		fmt.Println("Error while parsing the date time :", err)
	}
	var schedule = ifc.Schedule{
		Name:            input.Name,
		Enabled:         input.Enabled,
		ParallelRuns:    input.ParallelRuns,
		WorkflowName:    input.WorkflowName,
		WorkflowVersion: input.WorkflowVersion,
		CronString:      input.CronString,
		WorkflowContext: workflowContext,
		FromDate:        &dateFrom,
		ToDate:          &dateTo,
	}

	err = schedule.ValidateAndUpdate()
	if err != nil {
		logrus.Debugf("Error validating schedule. err=%v", err)
		return nil, fmt.Errorf("Error validating schedule %s", err)
	}

	found, err := scheduler.Configuration.Db.FindByName(schedule.Name)
	if err != nil {
		logrus.Debugf("Error checking for existing schedule name. err=%v", err)
		return nil, fmt.Errorf("Error checking for existing schedule name. err=%v", err)

	}
	if found != nil {
		logrus.Debugf("Duplicate schedule name '%s'", schedule.Name)
		return nil, fmt.Errorf("Duplicate schedule name '%s'", schedule.Name)

	}

	err = scheduler.Configuration.Db.Insert(schedule)
	if err != nil {
		logrus.Debugf("Error storing schedule to the database. err=%s", err)
		return nil, fmt.Errorf("Error storing schedule to the database. err=%s", err)

	}
	scheduler.PrepareTimers()

	return &model.Schedule{
		Name:            input.Name,
		Enabled:         input.Enabled,
		ParallelRuns:    input.ParallelRuns,
		WorkflowName:    input.WorkflowName,
		WorkflowVersion: input.WorkflowVersion,
		CronString:      input.CronString,
		WorkflowContext: input.WorkflowContext,
		FromDate:        input.FromDate,
		ToDate:          input.ToDate,
	}, nil
}

// UpdateSchedule is the resolver for the updateSchedule field.
func (r *mutationResolver) UpdateSchedule(ctx context.Context, input model.UpdateScheduleInput) (*model.Schedule, error) {

	var workflowContext map[string]interface{}
	json.Unmarshal([]byte(input.WorkflowContext), &workflowContext)

	dateFrom, err := time.Parse(time.RFC3339, input.FromDate)
	if err != nil {
		fmt.Println("Error while parsing the date time :", err)
	}

	dateTo, err := time.Parse(time.RFC3339, input.ToDate)
	if err != nil {
		fmt.Println("Error while parsing the date time :", err)
	}

	var schedule = ifc.Schedule{
		Name:            input.Name,
		Enabled:         input.Enabled,
		ParallelRuns:    input.ParallelRuns,
		WorkflowName:    input.WorkflowName,
		WorkflowVersion: input.WorkflowVersion,
		CronString:      input.CronString,
		WorkflowContext: workflowContext,
		FromDate:        &dateFrom,
		ToDate:          &dateTo,
	}

	err = schedule.ValidateAndUpdate()
	if err != nil {
		logrus.Debugf("Error validating schedule. err=%v", err)
		return nil, fmt.Errorf("Error validating schedule %s", err)
	}

	found, err := scheduler.Configuration.Db.FindByName(schedule.Name)
	if err != nil {
		logrus.Debugf("Error checking for existing schedule name. err=%v", err)
		return nil, fmt.Errorf("Error checking for existing schedule name")

	}
	if found == nil {
		logrus.Debugf("Schedule not found with name '%s'", schedule.Name)
		return nil, fmt.Errorf("Schedule not found with name '%s'", schedule.Name)

	}

	err = scheduler.Configuration.Db.Update(schedule)
	if err != nil {
		logrus.Debugf("Error storing schedule to the database. err=%s", err)
		return nil, fmt.Errorf("Error storing schedule to the database. err=%s", err)

	}

	scheduler.PrepareTimers()
	return &model.Schedule{
		Name:            input.Name,
		Enabled:         input.Enabled,
		ParallelRuns:    input.ParallelRuns,
		WorkflowName:    input.WorkflowName,
		WorkflowVersion: input.WorkflowVersion,
		CronString:      input.CronString,
		WorkflowContext: input.WorkflowContext,
		FromDate:        input.FromDate,
		ToDate:          input.ToDate,
	}, nil
}

// DeleteSchedule is the resolver for the deleteSchedule field.
func (r *mutationResolver) DeleteSchedule(ctx context.Context, input model.DeleteScheduleInput) (bool, error) {
	err := scheduler.Configuration.Db.RemoveByName(input.Name)
	if err != nil {
		logrus.Debugf("Error deleting schedule. err=%v", err)
		return false, fmt.Errorf("Error deleting schedule. err=%v", err)
	}

	scheduler.PrepareTimers()
	return true, nil
}

// Schedule is the resolver for the schedule field.
func (r *queryResolver) Schedule(ctx context.Context, name string) (*model.Schedule, error) {
	schedule, err := scheduler.Configuration.Db.FindByName(name)
	if err != nil {
		logrus.Debugf("Error getting schedule with name '%s'. err=%v", name, err)
		return nil, fmt.Errorf("Error getting schedule with name '%s'. err=%v", name, err)
	}

	if schedule == nil {
		logrus.Debugf("Error getting schedule with name '%s'", name)
		return nil, fmt.Errorf("Error getting schedule with name '%s'", name)
	}

	return ConvertIfcToModel(schedule), nil
}

// Schedules is the resolver for the schedules field.
func (r *queryResolver) Schedules(ctx context.Context) ([]*model.Schedule, error) {
	schedules, err := scheduler.Configuration.Db.FindAll()
	if err != nil {
		logrus.Debugf("Error listing schedules. err=%v", err)
		return nil, fmt.Errorf("Error listing schedules. err=%v", err)
	}

	model_schedules := make([]*model.Schedule, len(schedules))
	for i, v := range schedules {
		model_schedules[i] = ConvertIfcToModel(&v)
	}

	return model_schedules, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
