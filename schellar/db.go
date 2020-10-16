package main

import (
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

var (
	mongoSession *mgo.Session
)

func initMongo() error {
	logrus.Debugf("Connecting to MongoDB")
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    strings.Split(mongoAddress, ","),
		Timeout:  2 * time.Second,
		Database: dbName,
		Username: mongoUsername,
		Password: mongoPassword,
	}

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
		return errors.New("Couldn't connect to MongoDB")
	}
	return nil
}

func FindAll() ([]Schedule, error) {
	sc := mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(dbName).C("schedules")
	schedules := make([]Schedule, 0)
	err := st.Find(nil).All(&schedules)
	return schedules, err
}

func FindAllByEnabled(enabled bool) ([]Schedule, error) {
	sc := mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(dbName).C("schedules")
	var activeSchedules []Schedule
	err := st.Find(map[string]interface{}{"enabled": enabled}).All(&activeSchedules)
	return activeSchedules, err
}

func FindByName(scheduleName string) (*Schedule, error) {
	sc := mongoSession.Copy()
	defer sc.Close()

	var schedule Schedule
	st := sc.DB(dbName).C("schedules")
	err := st.Find(map[string]interface{}{"name": scheduleName}).One(&schedule)
	return &schedule, err
}

func FindByStatus(status string) ([]Schedule, error) {
	sc := mongoSession.Copy()
	defer sc.Close()

	sch := sc.DB(dbName).C("schedules")
	schedules := make([]Schedule, 0)
	err := sch.Find(map[string]interface{}{"status": status}).All(&schedules)
	return schedules, err
}

func UpdateStatus(scheduleName string, scheduleStatus string) error {
	sc := mongoSession.Copy()
	defer sc.Close()

	statusMap := make(map[string]interface{})
	statusMap["status"] = scheduleStatus
	statusMap["lastUpdate"] = time.Now()

	sr := sc.DB(dbName).C("schedules")
	return sr.Update(map[string]interface{}{"name": scheduleName}, map[string]interface{}{"$set": statusMap})
}

func UpdateStatusAndWorkflowContext(schedule Schedule) error {
	sc := mongoSession.Copy()
	defer sc.Close()

	sch := sc.DB(dbName).C("schedules")
	scheduleMap := make(map[string]interface{})
	scheduleMap["status"] = schedule.Status
	scheduleMap["lastUpdate"] = time.Now()
	scheduleMap["workflowContext"] = schedule.WorkflowContext

	return sch.Update(map[string]interface{}{"name": schedule.Name}, map[string]interface{}{"$set": scheduleMap})
}

func Insert(schedule Schedule) error {
	sc := mongoSession.Copy()
	defer sc.Close()
	st := sc.DB(dbName).C("schedules")
	return st.Insert(schedule)
}

func Update(schedule Schedule) error {
	sc := mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(dbName).C("schedules")
	return st.Update(map[string]interface{}{"name": schedule.Name}, map[string]interface{}{"$set": schedule})
}

func RemoveByName(scheduleName string) error {
	sc := mongoSession.Copy()
	defer sc.Close()

	st := sc.DB(dbName).C("schedules")
	return st.Remove(map[string]interface{}{"name": scheduleName})
}
