package main

import (
	"sort"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
	"github.com/pivotal-cf/spinnaker-resource/spinnaker"
)

func main() {

	var request concourse.CheckRequest
	concourse.ReadCheckRequest(&request)

	spinClient, err := spinnaker.NewClient(request.Source)
	concourse.Check("check", err)

	Data, err := spinClient.GetPipelineExecutions()
	concourse.Check("check", err)

	pipelineExecutions := make([]spinnaker.PipelineExecution, 1)
	for _, execution := range Data {
		if execution.Name == request.Source.SpinnakerPipeline {
			pipelineExecutions = append(pipelineExecutions, execution)
		}
	}
	//Sort Data by build time Asc
	sort.Slice(pipelineExecutions, func(i, j int) bool {
		return pipelineExecutions[i].BuildTime < pipelineExecutions[j].BuildTime
	})

	//find the input ExecutionID
	refLoc := sort.Search(len(pipelineExecutions), func(i int) bool {
		return pipelineExecutions[i].ID == request.Version.Ref
	})

	// i is the latest element unless the executionId exists
	i := len(pipelineExecutions) - 1
	if refLoc < len(pipelineExecutions) {
		i = refLoc
	}

	//loop from the input execution onwards loop will just use the last element if input execution is not found
	var res concourse.CheckResponse
	for ; i < len(pipelineExecutions); i++ {
		res = append(res, concourse.Version{Ref: pipelineExecutions[i].ID})
	}
	concourse.WriteCheckResponse(res)
}
