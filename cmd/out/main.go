package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
	"github.com/pivotal-cf/spinnaker-resource/spinnaker"
)

var (
	TriggerParams []byte
	Params        string
	Data          map[string]interface{}
	spinClient    spinnaker.SpinClient
)

func main() {
	var request concourse.OutRequest
	var err error
	concourse.ReadRequest(&request)

	spinClient, err = spinnaker.NewClient(request.Source)
	if err != nil {
		concourse.Fatal("put step failed", err)
	}

	pipelineExecutionID, err := invokePipeline(request)
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

func invokePipeline(request concourse.OutRequest) (string, error) {
	//Can be refactored
	if len(request.Params.TriggerParams) == 0 {
		TriggerParams = []byte(`{"type": "concourse-resource"}`)
	} else {
		for key, value := range request.Params.TriggerParams {
			Params = Params + fmt.Sprintf("\"%s\":\"%s\",", key, value)
		}
		Params = strings.TrimSuffix(Params, ",")
		// expand any variables
		paramsText := os.ExpandEnv(Params)
		TriggerParams = []byte(`{"type": "concourse-resource", "parameters": {` + paramsText + `}}`)
	}

	concourse.Sayf("Executing pipeline: '%s/%s'\n", request.Source.SpinnakerApplication, request.Source.SpinnakerPipeline)

	pipelineExecution, err := spinClient.InvokePipelineExecution(TriggerParams)
	if err != nil {
		return "", err
	}
	return pipelineExecution.ID, nil
}

func pollSpinnakerForStatus(request concourse.OutRequest, pipelineExecutionID string) error {
	statusReached := false
	interval := request.Source.StatusCheckInterval
	timeout := request.Source.StatusCheckTimeout
	if interval == 0*time.Second {
		interval = 30 * time.Second
	}

	if timeout < interval {
		timeout = interval + 1*time.Second
	}
	concourse.Sayf("Interval: %v, Timeout: %v\n", interval, timeout)
	pollTicker := time.NewTicker(interval)
	timeoutTicker := time.NewTicker(timeout)
	for {
		select {
		case <-pollTicker.C:
			rawPipeline, err := spinClient.GetPipeline(pipelineExecutionID)
			concourse.Check("put", err)
			statusReached = checkStatus(rawPipeline["status"].(string), request.Source.Statuses)

			//Intermediate statuses
			if statusReached {
				return nil
			}
			status := rawPipeline["status"].(string)
			if status != "RUNNING" && status != "NOT_STARTED" && status != "BUFFERED" {
				return fmt.Errorf("Pipeline execution reached a final state: %s", status)
			}

		case <-timeoutTicker.C:
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
