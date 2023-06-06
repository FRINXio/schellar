package mongo

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/frinx/schellar/ifc"

	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

type MongoDB struct {
	mongoSession *mgo.Session
	dbName       string
}

func InitDB() ifc.DB {
	address := ifc.GetEnvOrDefault("MONGO_ADDRESS", "mongo")
	username := ifc.GetEnvOrDefault("MONGO_USERNAME", "root")
	password := ifc.GetEnvOrDefault("MONGO_PASSWORD", "root")
	dbName := ifc.GetEnvOrDefault("MONGO_DB", "admin")

	if address == "" {
		logrus.Fatalf("'MONGO_ADDRESS' is required")
	}

	logrus.Debugf("Connecting to MongoDB MONGO_ADDRESS='%s', MONGO_DB='%s', "+
		"MONGO_USERNAME='%s', len(MONGO_PASSWORD)=%d",
		address, dbName, username, len(password))

	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    strings.Split(address, ","),
		Timeout:  2 * time.Second,
		Database: dbName,
		Username: username,
		Password: password,
	}

	var mongoSession *mgo.Session
	ms, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		logrus.Fatalf("Couldn't connect to MongoDB. err=%s", err)
	}
	mongoSession = ms
	logrus.Infof("Connected to MongoDB successfully")
	return MongoDB{mongoSession, dbName}
}

func (db MongoDB) FindAll() ([]ifc.Schedule, error) {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(db.dbName).C("schedules")
	schedules := make([]ifc.Schedule, 0)
	err := st.Find(nil).All(&schedules)
	return schedules, err
}

func (db MongoDB) FindAllByWorkflowType(workflowName string, workflowId string) ([]ifc.Schedule, error) {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(db.dbName).C("schedules")
	var activeSchedules []ifc.Schedule
	err := st.Find(map[string]interface{}{"workflowName": workflowName, "workflowId": workflowId}).All(&activeSchedules) // TODO SORT BY SCHEDULE_NAME ASC
	return activeSchedules, err
}

func (db MongoDB) FindAllByEnabled(enabled bool) ([]ifc.Schedule, error) {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(db.dbName).C("schedules")
	var activeSchedules []ifc.Schedule
	err := st.Find(map[string]interface{}{"enabled": enabled}).All(&activeSchedules)
	return activeSchedules, err
}

func (db MongoDB) FindByName(scheduleName string) (*ifc.Schedule, error) {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(db.dbName).C("schedules")
	schedules := make([]ifc.Schedule, 0)
	err := st.Find(map[string]interface{}{"name": scheduleName}).All(&schedules)
	if err != nil {
		return nil, err
	}
	if len(schedules) == 1 {
		return &schedules[0], nil
	} else if len(schedules) == 0 {
		return nil, nil
	}
	return nil, errors.New(
		fmt.Sprintf(
			"Unexpected result for FindByName('%s'): Found %d items",
			scheduleName, len(schedules),
		))
}

func (db MongoDB) FindByStatus(status string) ([]ifc.Schedule, error) {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	sch := sc.DB(db.dbName).C("schedules")
	schedules := make([]ifc.Schedule, 0)
	err := sch.Find(map[string]interface{}{"status": status}).All(&schedules)
	return schedules, err
}

func (db MongoDB) UpdateStatus(scheduleName string, scheduleStatus string) error {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	statusMap := make(map[string]interface{})
	statusMap["status"] = scheduleStatus
	statusMap["lastUpdate"] = time.Now()

	sr := sc.DB(db.dbName).C("schedules")
	return sr.Update(map[string]interface{}{"name": scheduleName}, map[string]interface{}{"$set": statusMap})
}

func (db MongoDB) UpdateStatusAndWorkflowContext(schedule ifc.Schedule) error {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	sch := sc.DB(db.dbName).C("schedules")
	scheduleMap := make(map[string]interface{})
	scheduleMap["status"] = schedule.Status
	scheduleMap["lastUpdate"] = time.Now()
	scheduleMap["workflowContext"] = schedule.WorkflowContext

	return sch.Update(map[string]interface{}{"name": schedule.Name}, map[string]interface{}{"$set": scheduleMap})
}

func (db MongoDB) Insert(schedule ifc.Schedule) error {
	sc := db.mongoSession.Copy()
	defer sc.Close()
	st := sc.DB(db.dbName).C("schedules")
	return st.Insert(schedule)
}

func (db MongoDB) Update(schedule ifc.Schedule) error {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(db.dbName).C("schedules")
	return st.Update(map[string]interface{}{"name": schedule.Name}, map[string]interface{}{"$set": schedule})
}

func (db MongoDB) RemoveByName(scheduleName string) error {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(db.dbName).C("schedules")
	return st.Remove(map[string]interface{}{"name": scheduleName})
}
