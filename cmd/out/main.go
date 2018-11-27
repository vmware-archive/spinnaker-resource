package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
	"github.com/pivotal-cf/spinnaker-resource/spinnaker"
)

var (
	TriggerParams []byte
	Params        string
	Data          map[string]interface{}
)

func main() {
	var request concourse.OutRequest
	concourse.ReadRequest(&request)

	spinClient, err := spinnaker.NewClient(request.Source)
	if err != nil {
		concourse.Fatal("put step failed ", err)
	}

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

	output := concourse.OutResponse{}
	pipelineExecution, err := spinClient.InvokePipelineExecution(TriggerParams)
	if err != nil {
		concourse.Fatal("put step failed ", err)
	}
	output.Version = concourse.Version{
		Ref: pipelineExecution.ID,
	}

	concourse.Sayf("Pipeline executed successfully")

	concourse.WriteResponse(output)
}
