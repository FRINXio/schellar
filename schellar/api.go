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
	router.HandleFunc("/liveness", getLiveness).Methods("GET")

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
		logrus.Debugf("Error creating schedule. err=%v", err)
		writeResponse(w, http.StatusBadRequest, "Error creating schedule")
		return
	}
	err = schedule.ValidateAndUpdate()
	if err != nil {
		logrus.Debugf("Error validating schedule. err=%v", err)
		writeResponse(w, http.StatusBadRequest, "Error validating schedule")
		return
	}

	//check duplicate schedule
	found, err := db.FindByName(schedule.Name)
	if err != nil {
		logrus.Debugf("Error checking for existing schedule name. err=%v", err)
		writeResponse(w, http.StatusInternalServerError, "Error checking for existing schedule name")
		return
	}
	if found != nil {
		logrus.Debugf("Duplicate schedule name '%s'", schedule.Name)
		writeResponse(w, http.StatusBadRequest, "Duplicate schedule name")
		return
	}

	logrus.Debugf("Saving schedule %s for workflow %s", schedule.Name, schedule.WorkflowName)
	logrus.Debugf("schedule: %v", schedule)

	err = db.Insert(schedule)
	if err != nil {
		logrus.Debugf("Error storing schedule to the database. err=%s", err)
		writeResponse(w, http.StatusInternalServerError, "Error storing schedule.")
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
		logrus.Debugf("Error updating schedule. err=%v", err)
		writeResponse(w, http.StatusBadRequest, "Error updating schedule.")
		return
	}
	err = schedule.ValidateAndUpdate()
	if err != nil {
		logrus.Debugf("Error validating updated schedule. err=%v", err)
		writeResponse(w, http.StatusBadRequest, "Error validating updated schedule.")
		return
	}

	logrus.Debugf("Updating schedule with %v", schedule)
	found, err := db.FindByName(name)
	if err != nil {
		logrus.Debugf("Error finding schedule by name '%s'. err=%s", name, err)
		writeResponse(w, http.StatusInternalServerError, "Error finding schedule")
		return
	}
	if found == nil {
		logrus.Debugf("Schedule not found with name '%s'", name)
		writeResponse(w, http.StatusNotFound, "Schedule not found")
		return
	}

	err = db.Update(schedule)
	if err != nil {
		logrus.Debugf("Error updating schedule '%s'. err=%v", name, err)
		writeResponse(w, http.StatusInternalServerError, "Error updating schedule")
		return
	}
	prepareTimers()
	writeResponse(w, http.StatusOK, fmt.Sprintf("Schedule updated successfully"))
	logrus.Debugf("Schedule '%s' updated successfuly", name)
}

func listSchedules(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("listSchedules r=%v", r)

	schedules, err := db.FindAll()
	if err != nil {
		logrus.Debugf("Error listing schedules. err=%v", err)
		writeResponse(w, http.StatusInternalServerError, "Error listing schedules")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	logrus.Debugf("Schedules=%v", schedules)
	b, err := json.Marshal(schedules)
	if err != nil {
		logrus.Debugf("Error serializing schedules. err=%v", err)
		writeResponse(w, http.StatusInternalServerError, "Error listing schedules")
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
		logrus.Debugf("Error getting schedule with name '%s'. err=%v", name, err)
		writeResponse(w, http.StatusInternalServerError, "Error getting schedule")
		return
	}

	if schedule == nil {
		logrus.Debugf("Error getting schedule with name '%s'", name)
		writeResponse(w, http.StatusNotFound, "Not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(schedule)
	if err != nil {
		logrus.Debugf("Error serializing schedules. err=%v", err)
		writeResponse(w, http.StatusInternalServerError, "Error getting schedule. err=%s")
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
		logrus.Debugf("Error deleting schedule. err=%v", err)
		writeResponse(w, http.StatusInternalServerError, "Error deleting schedule")
		return
	}
	prepareTimers()
	writeResponse(w, http.StatusOK, fmt.Sprintf("Deleted schedule successfully. name=%s", name))
}

func getLiveness(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("getLiveness r r=%v", r)

	if r.Method == http.MethodOptions {
		return
	}
	w.Write([]byte("OK"))
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
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")
			return
		}
		next.ServeHTTP(w, r)
	})
}
