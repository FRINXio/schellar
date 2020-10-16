package main

import (
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/frinx/schellar/ifc"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func startRestAPI() {
	router := mux.NewRouter()

	router.Use(customCorsMiddleware)
	router.HandleFunc("/schedule", createSchedule).Methods("POST", "OPTIONS")
	router.HandleFunc("/schedule", listSchedules).Methods("GET")
	router.HandleFunc("/schedule/{name}", getSchedule).Methods("GET")
	router.HandleFunc("/schedule/{name}", deleteSchedule).Methods("DELETE")
	router.HandleFunc("/schedule/{name}", updateSchedule).Methods("PUT", "OPTIONS")
	router.Handle("/metrics", promhttp.Handler())
	listen := fmt.Sprintf("0.0.0.0:3000")
	logrus.Infof("Listening at %s", listen)
	err := http.ListenAndServe(listen, router)
	if err != nil {
		logrus.Fatalf("Error while listening requests: %s", err)
	}
}

func createSchedule(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("createSchedule r=%v", r)

	decoder := json.NewDecoder(r.Body)
	var schedule ifc.Schedule
	err := decoder.Decode(&schedule)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Error handling post results. err=%s", err.Error()))
		return
	}
	err = schedule.ValidateAndUpdate()
	if err != nil {
		writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Error handling post results. err=%s", err.Error()))
		return
	}

	//check duplicate schedule
	found, err1 := db.FindByName(schedule.Name)
	if err1 != nil {
		writeResponse(w, http.StatusInternalServerError, "Error checking for existing schedule name")
		logrus.Errorf("Error checking for existing schedule name. err=%s", err1)
		return
	}
	if found != nil {
		writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Duplicate schedule name '%s'", schedule.Name))
		return
	}

	logrus.Debugf("Saving schedule %s for workflow %s", schedule.Name, schedule.WorkflowName)
	logrus.Debugf("schedule: %v", schedule)

	err0 := db.Insert(schedule)
	if err0 != nil {
		writeResponse(w, http.StatusInternalServerError, "Error storing schedule.")
		logrus.Errorf("Error storing schedule to the database. err=%s", err0)
		return
	}
	prepareTimers()
	logrus.Debugf("Sending response")
	w.Header().Set("Content-Type", "plain/text")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusCreated)
}

func updateSchedule(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("updateSchedule r=%v", r)
	name := mux.Vars(r)["name"]

	decoder := json.NewDecoder(r.Body)

	var schedule ifc.Schedule
	err := decoder.Decode(&schedule)
	if err != nil {
		writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Error handling post results. err=%s", err.Error()))
		return
	}
	err = schedule.ValidateAndUpdate()
	if err != nil {
		writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Error handling post results. err=%s", err.Error()))
		return
	}

	logrus.Debugf("Updating schedule with %v", schedule)
	found, err1 := db.FindByName(name)
	if err1 != nil {
		writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error updating schedule"))
		logrus.Errorf("Couldn't find schedule name %s. err=%s", name, err1)
		return
	}
	if found == nil {
		writeResponse(w, http.StatusNotFound, fmt.Sprintf("Couldn't find schedule %s", name))
		return
	}

	err = db.Update(schedule)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, "Error updating schedule")
		logrus.Errorf("Error updating schedule %s. err=%s", name, err)
		return
	}
	prepareTimers()
	writeResponse(w, http.StatusOK, fmt.Sprintf("Schedule updated successfully"))
}

func listSchedules(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("listSchedules r=%v", r)

	schedules, err := db.FindAll()
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error listing schedules. err=%s", err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	logrus.Debugf("Schedules=%v", schedules)
	b, err0 := json.Marshal(schedules)
	if err0 != nil {
		writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error listing schedules. err=%s", err.Error()))
		return
	}
	w.Write(b)
	logrus.Debugf("result: %s", string(b))
}

func getSchedule(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("getSchedule r=%v", r)
	name := mux.Vars(r)["name"]

	schedule, err := db.FindByName(name)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error getting schedule. err=%s", err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	b, err0 := json.Marshal(schedule)
	if err0 != nil {
		writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error getting schedule. err=%s", err.Error()))
		return
	}
	w.Write(b)
	logrus.Debugf("result: %s", string(b))
}

func deleteSchedule(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("deleteSchedule r=%v", r)
	name := mux.Vars(r)["name"]

	err := db.RemoveByName(name)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error deleting schedule. err=%s", err.Error()))
		return
	}
	prepareTimers()
	writeResponse(w, http.StatusOK, fmt.Sprintf("Deleted schedule successfully. name=%s", name))
}

func writeResponse(w http.ResponseWriter, statusCode int, message string) {
	msg := make(map[string]string)
	msg["message"] = message
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(msg)
}

func customCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
			return
		}
		next.ServeHTTP(w, r)
	})
}
