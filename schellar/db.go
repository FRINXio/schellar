package main

import (
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

type MongoDB struct {
	mongoSession *mgo.Session
}

func InitDB() (DB, error) {
	logrus.Debugf("Connecting to MongoDB")
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    strings.Split(mongoAddress, ","),
		Timeout:  2 * time.Second,
		Database: dbName,
		Username: mongoUsername,
		Password: mongoPassword,
	}

	var mongoSession *mgo.Session
	for i := 0; i < 30; i++ {
		ms, err := mgo.DialWithInfo(mongoDBDialInfo)
		if err != nil {
			logrus.Infof("Couldn't connect to mongoDB. err=%s", err)
			time.Sleep(1 * time.Second)
			logrus.Infof("Retrying...")
			continue
		}
		mongoSession = ms
		logrus.Infof("Connected to MongoDB successfully")
		break
	}

	if mongoSession == nil {
		return nil, errors.New("Couldn't connect to MongoDB")
	}
	return MongoDB{mongoSession}, nil
}

func (db MongoDB) FindAll() ([]Schedule, error) {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(dbName).C("schedules")
	schedules := make([]Schedule, 0)
	err := st.Find(nil).All(&schedules)
	return schedules, err
}

func (db MongoDB) FindAllByEnabled(enabled bool) ([]Schedule, error) {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(dbName).C("schedules")
	var activeSchedules []Schedule
	err := st.Find(map[string]interface{}{"enabled": enabled}).All(&activeSchedules)
	return activeSchedules, err
}

func (db MongoDB) FindByName(scheduleName string) (*Schedule, error) {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	var schedule Schedule
	st := sc.DB(dbName).C("schedules")
	err := st.Find(map[string]interface{}{"name": scheduleName}).One(&schedule)
	return &schedule, err
}

func (db MongoDB) FindByStatus(status string) ([]Schedule, error) {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	sch := sc.DB(dbName).C("schedules")
	schedules := make([]Schedule, 0)
	err := sch.Find(map[string]interface{}{"status": status}).All(&schedules)
	return schedules, err
}

func (db MongoDB) UpdateStatus(scheduleName string, scheduleStatus string) error {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	statusMap := make(map[string]interface{})
	statusMap["status"] = scheduleStatus
	statusMap["lastUpdate"] = time.Now()

	sr := sc.DB(dbName).C("schedules")
	return sr.Update(map[string]interface{}{"name": scheduleName}, map[string]interface{}{"$set": statusMap})
}

func (db MongoDB) UpdateStatusAndWorkflowContext(schedule Schedule) error {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	sch := sc.DB(dbName).C("schedules")
	scheduleMap := make(map[string]interface{})
	scheduleMap["status"] = schedule.Status
	scheduleMap["lastUpdate"] = time.Now()
	scheduleMap["workflowContext"] = schedule.WorkflowContext

	return sch.Update(map[string]interface{}{"name": schedule.Name}, map[string]interface{}{"$set": scheduleMap})
}

func (db MongoDB) Insert(schedule Schedule) error {
	sc := db.mongoSession.Copy()
	defer sc.Close()
	st := sc.DB(dbName).C("schedules")
	return st.Insert(schedule)
}

func (db MongoDB) Update(schedule Schedule) error {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(dbName).C("schedules")
	return st.Update(map[string]interface{}{"name": schedule.Name}, map[string]interface{}{"$set": schedule})
}

func (db MongoDB) RemoveByName(scheduleName string) error {
	sc := db.mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(dbName).C("schedules")
	return st.Remove(map[string]interface{}{"name": scheduleName})
}
