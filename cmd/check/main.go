package main

import (
	"sort"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
	"github.com/pivotal-cf/spinnaker-resource/spinnaker"
)

func main() {
	var request concourse.CheckRequest
	concourse.ReadRequest(&request)

	spinClient, err := spinnaker.NewClient(request.Source)
	if err != nil {
		concourse.Fatal("check step failed", err)
	}

	Data, err := spinClient.GetPipelineExecutions()
	if err != nil {
		concourse.Fatal("check step failed", err)
	}

	pipelineExecutions := FilterName(request.Source.SpinnakerPipeline, Data)

	pipelineExecutions = FilterStatus(request.Source.Statuses, pipelineExecutions)

	if len(pipelineExecutions) == 0 {
		concourse.WriteResponse(concourse.CheckResponse{})
	}

	//Sort Data by build time Asc
	sort.Slice(pipelineExecutions, func(i, j int) bool {
		return pipelineExecutions[i].BuildTime < pipelineExecutions[j].BuildTime
	})

	refLoc := len(pipelineExecutions) - 1
	for i, execution := range pipelineExecutions {
		if execution.ID == request.Version.Ref {
			refLoc = i
			break
		}
	}

	//loop from the input execution onwards loop will just use the last element if input execution is not found
	var res concourse.CheckResponse
	for i := refLoc; i < len(pipelineExecutions); i++ {
		res = append(res, concourse.Version{Ref: pipelineExecutions[i].ID})
	}
	concourse.WriteResponse(res)
}

func FilterName(name string, pes []spinnaker.PipelineExecution) []spinnaker.PipelineExecution {
	pe := make([]spinnaker.PipelineExecution, 0)
	for _, pipeExec := range pes {
		if pipeExec.Name == name {
			pe = append(pe, pipeExec)
		}
	}
	return pe
}

func CheckStatus(status string, statuses []string) bool {
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

func FilterStatus(statuses []string, pes []spinnaker.PipelineExecution) []spinnaker.PipelineExecution {
	pe := make([]spinnaker.PipelineExecution, 0)
	for _, pipeExec := range pes {
		if CheckStatus(pipeExec.Status, statuses) {
			pe = append(pe, pipeExec)
		}
	}
	return pe
}
