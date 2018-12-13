package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
	"github.com/pivotal-cf/spinnaker-resource/spinnaker"
)

var spinClient spinnaker.SpinClient

const defaultPollingInterval = 30 * time.Second

var triggerParamsBase = map[string]interface{}{"type": "concourse-resource"}

func main() {
	if len(os.Args) < 2 {
		concourse.Fatal(fmt.Sprintf("usage: %s <sources director>\n", os.Args[0]), errors.New("Not enough arguments supplied"))
	}

	var request concourse.OutRequest
	var err error
	concourse.ReadRequest(&request)

	sourcesDir := os.Args[1]

	spinClient, err = spinnaker.NewClient(request.Source)
	if err != nil {
		concourse.Fatal("put step failed", err)
	}

	pipelineExecutionID, err := invokePipeline(sourcesDir, request)
	if err != nil {
		concourse.Fatal("put step failed", err)
	}
	if len(request.Source.Statuses) > 0 {
		err = pollSpinnakerForStatus(request, pipelineExecutionID)
		if err != nil {
			concourse.Fatal("put step failed", err)
		}
		writeSuccessfulResponse(pipelineExecutionID)
	}
	writeSuccessfulResponse(pipelineExecutionID)
}

func invokePipeline(sourcesDir string, request concourse.OutRequest) (string, error) {
	TriggerParamsMap := triggerParamsBase

	triggerParams := map[string]string{}
	if len(request.Params.TriggerParams) > 0 {
		for key, value := range request.Params.TriggerParams {
			triggerParams[key] = os.ExpandEnv(value)
		}
	}
	if len(request.Params.TriggerParamsJSONFilePath) > 0 {
		localPath := filepath.Join(sourcesDir, request.Params.TriggerParamsJSONFilePath)
		dynamicTriggerParams, err := ioutil.ReadFile(localPath)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(dynamicTriggerParams, &triggerParams)
		if err != nil {
			return "", err
		}
	}
	if len(triggerParams) > 0 {
		TriggerParamsMap["parameters"] = triggerParams
	}
	if len(request.Params.Artifacts) > 0 {
		localPath := filepath.Join(sourcesDir, request.Params.Artifacts)
		artifacts, err := ioutil.ReadFile(localPath)
		if err != nil {
			return "", err
		}
		var JSONArtifacts interface{}
		err = json.Unmarshal(artifacts, &JSONArtifacts)
		if err != nil {
			return "", err
		}
		TriggerParamsMap["artifacts"] = JSONArtifacts
	}
	postBody, err := json.Marshal(TriggerParamsMap)
	if err != nil {
		return "", err
	}

	concourse.Sayf("Executing pipeline: '%s/%s'\n", request.Source.SpinnakerApplication, request.Source.SpinnakerPipeline)

	pipelineExecution, err := spinClient.InvokePipelineExecution(postBody)
	if err != nil {
		return "", err
	}
	return pipelineExecution.ID, nil
}

func pollSpinnakerForStatus(request concourse.OutRequest, pipelineExecutionID string) error {
	var statusReached bool
	interval := request.Source.StatusCheckInterval
	timeout := request.Source.StatusCheckTimeout
	if interval == 0*time.Second {
		interval = defaultPollingInterval
	}

	if timeout < interval {
		timeout = interval + 1*time.Second
	}
	concourse.Sayf("Poll Interval: %v, Timeout: %v\n", interval, timeout)
	pollTicker := time.NewTicker(interval)
	timeoutTicker := time.NewTicker(timeout)
	for {
		select {
		case <-pollTicker.C:
			rawPipeline, err := spinClient.GetPipelineExecution(pipelineExecutionID)
			if err != nil {
				concourse.Fatal("put step failed", err)
			}
			statusReached = checkStatus(rawPipeline["status"].(string), request.Source.Statuses)

			//Intermediate statuses
			if statusReached {
				concourse.Sayf("\n")
				return nil
			}
			status := rawPipeline["status"].(string)
			if status != "RUNNING" && status != "NOT_STARTED" && status != "BUFFERED" {
				concourse.Sayf("\n")
				return fmt.Errorf("Pipeline execution reached a final state: %s", status)
			}
			concourse.Sayf(".")

		case <-timeoutTicker.C:
			concourse.Sayf("\n")
			return fmt.Errorf("timed out waiting for configured status(es)")
		}
	}

}

func writeSuccessfulResponse(pipelineExecutionID string) {
	output := concourse.OutResponse{}
	output.Version = concourse.Version{
		Ref: pipelineExecutionID,
	}

	concourse.Sayf("Pipeline executed successfully")

	concourse.WriteResponse(output)
}

func checkStatus(status string, statuses []string) bool {
	if len(statuses) == 0 {
		return true
	}
	for _, currStatus := range statuses {
		if status == currStatus {
			return true
		}
	}
	return false
}
