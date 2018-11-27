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
	if err != nil {
		concourse.Fatal("check step failed ", err)
	}

	Data, err := spinClient.GetPipelineExecutions()
	if err != nil {
		concourse.Fatal("check step failed ", err)
	}

	//Sort Data by build time Asc
	sort.Slice(Data, func(i, j int) bool {
		return Data[i].BuildTime < Data[j].BuildTime
	})

	//find the input ExecutionID
	refLoc := sort.Search(len(Data), func(i int) bool {
		return Data[i].ID == request.Version.Ref
	})

	// i is the latest element unless the executionId exists
	i := len(Data) - 1
	if refLoc < len(Data) {
		i = refLoc
	}

	//loop from the input execution onwards
	//loop will just use the last element if input execution is not found
	var res concourse.CheckResponse
	for ; i < len(Data); i++ {
		if Data[i].Name == request.Source.SpinnakerPipeline {
			res = append(res, concourse.Version{Ref: Data[i].ID})
		}
	}
	concourse.WriteCheckResponse(res)
}
