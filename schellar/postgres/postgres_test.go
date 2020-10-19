package postgres

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/frinx/schellar/ifc"
)

func TestSqlParamsRange(t *testing.T) {
	actual := sqlParamsRange(3)
	expected := "($1,$2,$3)"
	if actual != expected {
		t.Fatalf("Unexpected: %v, should be %v", actual, expected)
	}
}

func initIntegration(t *testing.T) PostgresDB {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	db := InitDB()
	expectTableSize(db, 0, "before test", t)
	return db
}

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

func TestCRUDIntegration(t *testing.T) {
	db := initIntegration(t)

	someDomain := map[string]string{"*": "mydomain"}
	taskToDomains := []map[string]string{nil, someDomain}

	someCtx := map[string]interface{}{"key": map[string]interface{}{"k": "v"}}
	workflowCtxs := []map[string]interface{}{nil, someCtx}

	for _, taskToDomain := range taskToDomains {
		for _, workflowContext := range workflowCtxs {
			t.Run(fmt.Sprintf("taskToDomain=%v,workflowContext=%v", taskToDomain, workflowContext),
				func(t *testing.T) {
					// table should be empty
					expectTableSize(db, 0, "before test", t)

					now := time.Now().Truncate(time.Millisecond)
					schedule := makeSchedule(now)
					schedule.WorkflowContext = workflowContext
					schedule.TaskToDomain = taskToDomain
					err := db.Insert(schedule)
					if err != nil {
						t.Fatalf("Cannot insert. Err=%v", err)
					}
					// findall
					schedules := expectTableSize(db, 1, "after insert", t)
					actual := schedules[0]
					if !reflect.DeepEqual(schedule, actual) {
						t.Fatalf("Expected equality between inserted: \n%v and selected \n%v", schedule, actual)
					}
					// find by name
					found, err := db.FindByName(schedule.Name)
					if err != nil {
						t.Fatalf("Cannot find by name. Err=%v", err)
					}
					if !reflect.DeepEqual(schedule, *found) {
						t.Fatalf("Expected equality between inserted: \n%v and found by name \n%v", schedule, found)
					}

					defer db.RemoveByName(schedule.Name)
				})
		}
	}
	expectTableSize(db, 0, "after test", t)
}

func TestFindByNameIntegration(t *testing.T) {
	db := initIntegration(t)
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

func TestUpdateStatusIntegration(t *testing.T) {
	db := initIntegration(t)
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
	schedules := expectTableSize(db, 1, "after insert", t)
	actual := schedules[0]
	// check equality
	if !reflect.DeepEqual(schedule, actual) {
		t.Fatalf("Expected equality between inserted: \n%v and selected \n%v", schedule, actual)
	}
}

func TestUpdateStatusAndWorkflowContextIntegration(t *testing.T) {
	db := initIntegration(t)
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
	schedules := expectTableSize(db, 1, "after insert", t)
	actual := schedules[0]
	// check equality
	if !reflect.DeepEqual(schedule, actual) {
		t.Fatalf("Expected equality between inserted: \n%v and selected \n%v", schedule, actual)
	}
}

func TestUpdateIntegration(t *testing.T) {
	db := initIntegration(t)
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
	schedules := expectTableSize(db, 1, "after insert", t)
	actual := schedules[0]
	// check equality
	if !reflect.DeepEqual(schedule, actual) {
		t.Fatalf("Expected equality between inserted: \n%v and selected \n%v", schedule, actual)
	}
}

func expectTableSize(db PostgresDB, expectedSize int, hint string, t *testing.T) []ifc.Schedule {
	schedules, err := db.FindAll()
	if err != nil || len(schedules) != expectedSize {
		t.Fatalf("Unexpected state %s. Err=%v. Len=%d", hint, err, len(schedules))
	}
	return schedules
}
