package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func launchWorkflow(scheduleName string) error {
	logrus.Debugf("startWorkflow scheduleName=%s", scheduleName)

	logrus.Debugf("Loading schedule definitions from DB")

	schedule, err := Configuration.Db.FindByName(scheduleName)
	if err != nil {
		logrus.Errorf("Couldn't find schedule %s", scheduleName)
		return err
	}

	wf := make(map[string]interface{})
	wf["name"] = schedule.WorkflowName
	wf["version"] = schedule.WorkflowVersion
	wf["input"] = schedule.WorkflowContext
	wf["input"].(map[string]interface{})["scheduleName"] = schedule.Name
	if len(schedule.CorrelationID) > 0 {
		wf["correlationId"] = schedule.CorrelationID
	}
	if len(schedule.TaskToDomain) > 0 {
		wf["taskToDomain"] = schedule.TaskToDomain
	}
	wfb, _ := json.Marshal(wf)

	logrus.Debugf("Launching Workflow %s", wf)
	url := fmt.Sprintf("%s/workflow", Configuration.ConductorURL)
	resp, data, err := postHTTP(url, wfb, Configuration.AdminGroups, Configuration.AdminRoles, Configuration.From)
	if err != nil {
		logrus.Errorf("Call to Conductor POST /workflow failed. err=%s", err)
		return err
	}
	if resp.StatusCode != 200 {
		logrus.Warnf("POST /workflow call status!=200. resp=%v", resp)
		return fmt.Errorf("Failed to create new workflow instance. status=%d", resp.StatusCode)
	}
	logrus.Infof("Schedule %s: Workflow %s launched. workflowId=%s", schedule.Name, schedule.WorkflowName, string(data))
	return nil
}

func getWorkflow(name string, version string) (map[string]interface{}, error) {
	logrus.Debugf("getWorkflow %s", name)
	resp, data, err := getHTTP(fmt.Sprintf("%s/metadata/workflow/%s?version=%s", Configuration.ConductorURL, name, version), Configuration.AdminGroups, Configuration.AdminRoles, Configuration.From)
	if err != nil {
		return nil, fmt.Errorf("GET /metadata/workflow/name failed. err=%s", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Couldn't get workflow info. name=%s", name)
	}
	var wfdata map[string]interface{}
	err = json.Unmarshal(data, &wfdata)
	if err != nil {
		logrus.Errorf("Error parsing json. err=%s", err)
		return nil, err
	}
	return wfdata, nil
}

func getWorkflowInstance(workflowID string) (map[string]interface{}, error) {
	logrus.Debugf("getWorkflowInstance %s", workflowID)
	resp, data, err := getHTTP(fmt.Sprintf("%s/workflow/%s?includeTasks=false", Configuration.ConductorURL, workflowID), Configuration.AdminGroups, Configuration.AdminRoles, Configuration.From)
	if err != nil {
		return nil, fmt.Errorf("GET /workflow/%s?includeTasks=false failed. err=%s", err, workflowID)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Couldn't get workflow info. workflowId=%s. status=%d", workflowID, resp.StatusCode)
	}
	var wfdata map[string]interface{}
	err = json.Unmarshal(data, &wfdata)
	if err != nil {
		logrus.Errorf("Error parsing json. err=%s", err)
		return nil, err
	}
	return wfdata, nil
}

func findWorkflows(workflowType string, scheduleName string, running bool) (map[string]interface{}, error) {
	logrus.Debugf("findWorkflows(workflowType=%s,scheduleName=%s,running=%v)", workflowType, scheduleName, running)
	// escape : -> \:

	workflowType = strings.ReplaceAll(workflowType, ":", "\\:")
	scheduleName = strings.ReplaceAll(scheduleName, ":", "\\:")
	runstr := "status IN 'RUNNING'"
	if !running {
		runstr = "status IN 'FAILED,COMPLETED,TERMINATED,TIMED_OUT,PAUSED'"
	}

	query := fmt.Sprintf("workflowType='%s' AND %s", workflowType, runstr)

	logrus.Debugf("search query=%s", query)
	sr := fmt.Sprintf("%s/workflow/search?query=%s&sort=endTime:DESC&size=5", Configuration.ConductorURL, url.PathEscape(query))
	resp, data, err := getHTTP(sr, Configuration.AdminGroups, Configuration.AdminRoles, Configuration.From)

	if err != nil {
		return nil, fmt.Errorf("GET /workflow/search failed. err=%s", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GET /workflow/search failed. status=%d. err=%s", resp.StatusCode, err)
	}

	var wfdata map[string]interface{}
	err = json.Unmarshal(data, &wfdata)
	if err != nil {
		logrus.Errorf("Error parsing json. err=%s", err)
		return nil, err
	}
	return wfdata, nil
}

func postHTTP(url string, data []byte, groupHeader string, roleHeaders string, fromHeader string) (http.Response, []byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		logrus.Errorf("HTTP request creation failed. err=%s", err)
		return http.Response{}, []byte{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-auth-user-groups", groupHeader)
	req.Header.Set("x-auth-user-roles", roleHeaders)
	req.Header.Set("from", fromHeader)

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	logrus.Debugf("POST request=%v", req)
	response, err1 := client.Do(req)
	if err1 != nil {
		logrus.Errorf("HTTP request invocation failed. err=%s", err1)
		return http.Response{}, []byte{}, err1
	}

	logrus.Debugf("Response: %v", response)
	datar, _ := ioutil.ReadAll(response.Body)
	logrus.Debugf("Response body: %s", datar)
	return *response, datar, nil
}

func getHTTP(url0 string, groupHeader string, roleHeaders string, fromHeader string) (http.Response, []byte, error) {
	req, err := http.NewRequest("GET", url0, nil)
	if err != nil {
		logrus.Errorf("HTTP request creation failed. err=%s", err)
		return http.Response{}, []byte{}, err
	}

	req.Header.Set("x-auth-user-groups", groupHeader)
	req.Header.Set("x-auth-user-roles", roleHeaders)
	req.Header.Set("from", fromHeader)

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	logrus.Debugf("GET request=%v", req)
	response, err1 := client.Do(req)
	if err1 != nil {
		logrus.Errorf("HTTP request invocation failed. err=%s", err1)
		return http.Response{}, []byte{}, err1
	}

	// logrus.Debugf("Response: %v", response)
	datar, _ := ioutil.ReadAll(response.Body)
	logrus.Debugf("Response body: %s", datar)
	return *response, datar, nil
}
