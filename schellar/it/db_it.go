package it

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/frinx/schellar/ifc"
)

func makeSchedule(now time.Time) ifc.Schedule {
	return ifc.Schedule{
		Name:                "Name",
		Enabled:             true,
		Status:              "Status",
		WorkflowName:        "WorkflowName",
		WorkflowVersion:     "1",
		WorkflowContext:     nil,
		CronString:          "CronString",
		ParallelRuns:        false,
		CheckWarningSeconds: 10,
		FromDate:            nil,
		ToDate:              nil,
		LastUpdate:          now,
		CorrelationID:       "CorrelationID",
		TaskToDomain:        nil,
	}
}

func AllIntegration(t *testing.T, dbGetter func(*testing.T) ifc.DB) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Run("CRUDIntegration", func(t *testing.T) {
		CRUDIntegration(t, dbGetter)
	})
	t.Run("FindByNameIntegration", func(t *testing.T) {
		FindByNameIntegration(t, dbGetter)
	})
	t.Run("UpdateStatusIntegration", func(t *testing.T) {
		UpdateStatusIntegration(t, dbGetter)
	})
	t.Run("UpdateStatusAndWorkflowContextIntegration", func(t *testing.T) {
		UpdateStatusAndWorkflowContextIntegration(t, dbGetter)
	})
	t.Run("UpdateIntegration", func(t *testing.T) {
		UpdateIntegration(t, dbGetter)
	})
}

func assertEquals(t *testing.T, expected ifc.Schedule, actual ifc.Schedule, hint string) {
	if !reflect.DeepEqual(expected, actual) {
		expectedJSON, err1 := json.Marshal(expected)
		actualJSON, err2 := json.Marshal(actual)
		if err1 != nil || err2 != nil {
			t.Fatalf("%s Expected vs Actual:\n%v\n%v", hint, expected, actual)
		} else {
			t.Fatalf("%s Expected vs Actual:\n%v\n%v", hint, string(expectedJSON), string(actualJSON))
		}
	}
}

func testCRUD(t *testing.T, db ifc.DB, schedule ifc.Schedule) {
	// table should be empty
	ExpectTableSize(db, 0, "before test", t)
	// insert
	err := db.Insert(schedule)
	if err != nil {
		t.Fatalf("Cannot insert. Err=%v", err)
	}
	// defer remove
	defer db.RemoveByName(schedule.Name)
	// findall
	schedules := ExpectTableSize(db, 1, "after insert", t)
	actual := schedules[0]
	if schedule.WorkflowContext == nil {
		// selected WorkflowContext is never null
		schedule.WorkflowContext = make(map[string]interface{})
	}
	assertEquals(t, schedule, actual, "Insert error")
	// find by name
	found, err := db.FindByName(schedule.Name)
	if err != nil {
		t.Fatalf("Cannot find by name. Err=%v", err)
	}
	assertEquals(t, schedule, *found, "FindByName error")
}

func CRUDIntegration(t *testing.T, dbGetter func(*testing.T) ifc.DB) {
	db := dbGetter(t)

	someDomain := map[string]string{"*": "mydomain"}
	taskToDomains := []map[string]string{nil, someDomain}

	someCtx := map[string]interface{}{"key": map[string]interface{}{"k": "v"}}
	workflowCtxs := []map[string]interface{}{nil, someCtx}

	for _, taskToDomain := range taskToDomains {
		for _, workflowContext := range workflowCtxs {
			t.Run(fmt.Sprintf("taskToDomain=%v,workflowContext=%v", taskToDomain, workflowContext),
				func(t *testing.T) {
					now := time.Now().Truncate(time.Millisecond)
					schedule := makeSchedule(now)
					schedule.WorkflowContext = workflowContext
					schedule.TaskToDomain = taskToDomain
					testCRUD(t, db, schedule)
				})
		}
	}
	ExpectTableSize(db, 0, "after test", t)
}

func ExpectTableSize(db ifc.DB, expectedSize int, hint string, t *testing.T) []ifc.Schedule {
	schedules, err := db.FindAll()
	if err != nil || len(schedules) != expectedSize {
		t.Fatalf("Unexpected state %s. Err=%v. Len=%d", hint, err, len(schedules))
	}
	return schedules
}

func FindByNameIntegration(t *testing.T, dbGetter func(*testing.T) ifc.DB) {
	db := dbGetter(t)
	now := time.Now().Truncate(time.Millisecond)

	found, err := db.FindByName("404")
	if err != nil {
		t.Fatalf("Cannot FindByName: %v", err)
	}
	if found != nil {
		t.Fatalf("Unexpected FindByName: %v", found)
	}
	schedule := makeSchedule(now)
	err = db.Insert(schedule)
	if err != nil {
		t.Fatalf("Cannot insert: %v", err)
	}
	defer db.RemoveByName(schedule.Name)
	found, err = db.FindByName(schedule.Name)
	if err != nil {
		t.Fatalf("Unexpected FindByName: %v", err)
	}
	if found == nil {
		t.Fatalf("Not found after insert")
	}
}

func UpdateStatusIntegration(t *testing.T, dbGetter func(*testing.T) ifc.DB) {
	db := dbGetter(t)
	now := time.Now().Truncate(time.Millisecond)
	schedule := makeSchedule(now)
	err := db.Insert(schedule)
	if err != nil {
		t.Fatalf("Cannot insert: %v", err)
	}
	defer db.RemoveByName(schedule.Name)

	schedule.Status = "COMPLETED"
	err = db.UpdateStatus(schedule.Name, schedule.Status)
	if err != nil {
		t.Fatalf("Cannot update: %v", err)
	}
	schedules := ExpectTableSize(db, 1, "after insert", t)
	actual := schedules[0]
	// selected WorkflowContext is never null
	schedule.WorkflowContext = make(map[string]interface{})
	// check equality
	assertEquals(t, schedule, actual, "Inserted != selected")
}

func UpdateStatusAndWorkflowContextIntegration(t *testing.T, dbGetter func(*testing.T) ifc.DB) {
	db := dbGetter(t)
	now := time.Now().Truncate(time.Millisecond)
	schedule := makeSchedule(now)
	err := db.Insert(schedule)
	if err != nil {
		t.Fatalf("Cannot insert: %v", err)
	}
	defer db.RemoveByName(schedule.Name)

	schedule.Status = "COMPLETED"
	schedule.WorkflowContext = map[string]interface{}{"key": map[string]interface{}{"k": "v"}}
	err = db.UpdateStatusAndWorkflowContext(schedule)
	if err != nil {
		t.Fatalf("Cannot update: %v", err)
	}
	schedules := ExpectTableSize(db, 1, "after insert", t)
	actual := schedules[0]
	// check equality
	assertEquals(t, schedule, actual, "Inserted != selected")
}

func UpdateIntegration(t *testing.T, dbGetter func(*testing.T) ifc.DB) {
	db := dbGetter(t)
	now := time.Now().Truncate(time.Millisecond)
	schedule := makeSchedule(now)
	err := db.Insert(schedule)
	if err != nil {
		t.Fatalf("Cannot insert: %v", err)
	}
	defer db.RemoveByName(schedule.Name)
	schedule.FromDate = &now
	err = db.Update(schedule)
	if err != nil {
		t.Fatalf("Cannot update: %v", err)
	}
	schedules := ExpectTableSize(db, 1, "after insert", t)
	actual := schedules[0]
	// selected WorkflowContext is never null
	schedule.WorkflowContext = make(map[string]interface{})
	// check equality
	assertEquals(t, schedule, actual, "Inserted != selected")
}
