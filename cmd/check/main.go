/*
Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.

This program and the accompanying materials are made available under the terms of the under the Apache License, Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
*/
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

	pipelineExecutions := filterName(request.Source.SpinnakerPipeline, Data)

	pipelineExecutions = filterStatus(request.Source.Statuses, pipelineExecutions)

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
	responseExecutions := pipelineExecutions[refLoc:]
	for _, execution := range responseExecutions {
		res = append(res, concourse.Version{Ref: execution.ID})
	}
	concourse.WriteResponse(res)
}

func filterName(name string, pes []spinnaker.PipelineExecution) []spinnaker.PipelineExecution {
	pe := make([]spinnaker.PipelineExecution, 0)
	for _, pipeExec := range pes {
		if pipeExec.Name == name {
			pe = append(pe, pipeExec)
		}
	}
	return pe
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

func filterStatus(statuses []string, pes []spinnaker.PipelineExecution) []spinnaker.PipelineExecution {
	pe := make([]spinnaker.PipelineExecution, 0)
	for _, pipeExec := range pes {
		if checkStatus(pipeExec.Status, statuses) {
			pe = append(pe, pipeExec)
		}
	}
	return pe
}
