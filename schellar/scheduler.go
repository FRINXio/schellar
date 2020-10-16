package main

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

var (
	scheduledRoutineHashes = make(map[string]*cron.Cron)
)

func startScheduler() error {
	err := prepareTimers()
	if err != nil {
		return err
	}
	go checkRunningWorkflows()
	return nil
}

func prepareTimers() error {
	logrus.Debugf("Refreshing timers according to active schedules")

	activeSchedules, err := db.FindAllByEnabled(true)
	if err != nil {
		return err
	}

	//activate go routines for schedules that weren't activated yet
	for _, activeSchedule := range activeSchedules {
		isScheduled := false
		activeRoutineHash := fmt.Sprintf("%s|%s)", activeSchedule.Name, activeSchedule.CronString)
		for hashRoutine := range scheduledRoutineHashes {
			if activeRoutineHash == hashRoutine {
				isScheduled = true
				// break
			}
		}
		if !isScheduled {
			err := launchSchedule(activeSchedule.Name)
			if err != nil {
				return err
			}
		}
	}

	//remove go routines that are not currenctly active
	for hashRoutine, cronJob := range scheduledRoutineHashes {
		isActive := false
		for _, activeSchedule := range activeSchedules {
			activeRoutineHash := fmt.Sprintf("%s|%s)", activeSchedule.Name, activeSchedule.CronString)
			if hashRoutine == activeRoutineHash {
				isActive = true
				break
			}
		}
		if !isActive {
			logrus.Infof("Schedule %s: Stopping timer", hashRoutine)
			cronJob.Stop()
			delete(scheduledRoutineHashes, hashRoutine)
		}
	}

	return nil
}

func launchSchedule(scheduleName string) error {

	schedule0, err := db.FindByName(scheduleName)
	if err != nil {
		logrus.Errorf("Couldn't get schedule %s. err=%s", scheduleName, err)
		return err
	}

	c := cron.New()
	logrus.Infof("Schedule %s: Creating timer. cron=%s. workflow=%s", schedule0.Name, schedule0.CronString, schedule0.WorkflowName)
	c.AddFunc(schedule0.CronString, func() {
		logrus.Debugf("Processing timer trigger for schedule %s", scheduleName)

		schedule, err := db.FindByName(scheduleName)
		if err != nil {
			logrus.Errorf("Couldn't get schedule %s. err=%s", scheduleName, err)
			return
		}

		isBefore := false
		if schedule.ToDate == nil || time.Now().Before(*schedule.ToDate) {
			isBefore = true
		}
		isAfter := false
		if schedule.FromDate == nil || time.Now().After(*schedule.FromDate) {
			isAfter = true
		}
		if isBefore && isAfter {

			runningWorkflows, err2 := findWorkflows(schedule.WorkflowName, schedule.Name, true)
			if err2 != nil {
				logrus.Errorf("Error finding currently running workflows. err=%s", err2)
				return
			}

			runningTotalHits := int(runningWorkflows["totalHits"].(float64))

			scheduleStatus := "RUNNING"
			if runningTotalHits > 0 {
				if !schedule.ParallelRuns {
					wresults := runningWorkflows["results"]
					if wresults != nil {
						wf0 := wresults.([]interface{})[0]
						wf1 := wf0.(map[string]interface{})
						workflowID := wf1["workflowId"]
						logrus.Debugf("Schedule %s trigger skipped. Previous workflow id (%s) has not finished yet", schedule.Name, workflowID)
						return
					}
				}
				logrus.Infof("Schedule %s: Launching concurrent workflow (%s). count=%d", schedule.Name, schedule.WorkflowName, runningTotalHits)
			}

			logrus.Debugf("Launching workflow '%s' for schedule '%s'", schedule.WorkflowName, scheduleName)
			err := launchWorkflow(scheduleName)
			if err != nil {
				logrus.Errorf("Error launching Workflow err=%s", err)
				return
			}

			logrus.Debugf("Updating Schedule status. name=%s. status=%s", scheduleName, "RUNNING")
			err0 := db.UpdateStatus(scheduleName, scheduleStatus)
			if err0 != nil {
				logrus.Errorf("Error saving Schedule status err=%s", err0)
			}

		} else {
			logrus.Debugf("Schedule %s active, but not within activation date", scheduleName)
		}

	})
	routineHash := fmt.Sprintf("%s|%s)", schedule0.Name, schedule0.CronString)
	scheduledRoutineHashes[routineHash] = c
	go c.Start()
	return nil
}

func checkRunningWorkflows() {
	logrus.Debugf("Starting to check running workflow status")
	for {
		startTime := time.Now()
		schedules, err0 := db.FindByStatus("RUNNING")

		if err0 != nil {
			logrus.Errorf("Error getting running schedules. err=%s", err0)
			continue
		}

		if len(schedules) > 0 {
			logrus.Debugf("Checking running workflows on Conductor...")
		}
		for _, schedule := range schedules {
			runningWorkflows, err := findWorkflows(schedule.WorkflowName, schedule.Name, true)
			if err != nil {
				logrus.Errorf("Error finding workflows for schedule %s. err=%s", schedule.Name, err)
				continue
			}
			finishedWorkflows, err := findWorkflows(schedule.WorkflowName, schedule.Name, false)
			if err != nil {
				logrus.Errorf("Error finding workflows for schedule %s. err=%s", schedule.Name, err)
				continue
			}
			runningTotalHits := int(runningWorkflows["totalHits"].(float64))
			finishedTotalHits := int(finishedWorkflows["totalHits"].(float64))

			logrus.Debugf("Running workflows hits for schedule %s: %d", schedule.Name, runningTotalHits)
			logrus.Debugf("Finished workflows hits for schedule %s: %d", schedule.Name, finishedTotalHits)

			scheduleStatus := "RUNNING"
			var wfoutput map[string]interface{}
			if runningTotalHits == 0 {
				if finishedTotalHits == 0 {
					logrus.Errorf("No workflows found for schedule %s, but it is in state RUNNING", schedule.Name)
					continue
				} else {
					wf0 := finishedWorkflows["results"].([]interface{})[0]
					wf1 := wf0.(map[string]interface{})
					wf2, err := getWorkflowInstance(wf1["workflowId"].(string))
					if err != nil {
						logrus.Errorf("Could not get workflow instance. err=%s", err)
						continue
					}
					scheduleStatus = wf2["status"].(string)
					out, exists := wf2["output"]
					if exists {
						wfoutput = out.(map[string]interface{})
					}
				}
			}

			logrus.Debugf("Schedule status is %s", scheduleStatus)
			if scheduleStatus != schedule.Status {
				logrus.Infof("Schedule %s: Changing status to %s", schedule.Name, scheduleStatus)
			}
			schedule.Status = scheduleStatus
			if len(wfoutput) > 0 {
				logrus.Debugf("Adding last workflow output to schedule context. output=%s, workflowContext=%s",
					wfoutput, schedule.WorkflowContext)
				schedule.WorkflowContext["lastExecution"] = wfoutput
			}
			err0 = db.UpdateStatusAndWorkflowContext(schedule)
			if err0 != nil {
				logrus.Errorf("Error updating schedule %s to status %s. err=%s", schedule.Name, scheduleStatus, err0)
			}
		}

		elapsedTime := time.Now().Sub(startTime)
		remainingSleep := float64(checkIntervalSeconds) - elapsedTime.Seconds()
		if remainingSleep > 0 {
			logrus.Debugf("Sleeping for %d seconds...", int(remainingSleep))
			time.Sleep(time.Duration(remainingSleep) * time.Second)
		}
	}
}

func getStringValue(m map[string]interface{}, keyName string, defaultValue string) string {
	v, exists := m[keyName]
	if !exists {
		return defaultValue
	}
	return v.(string)
}
