package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
	"github.com/pivotal-cf/spinnaker-resource/spinnaker"
)

func main() {

	if len(os.Args) < 2 {
		concourse.Fatal("get step failed", fmt.Errorf("destination path not specified"))
	}

	var request concourse.InRequest
	concourse.ReadRequest(&request)

	spinClient, err := spinnaker.NewClient(request.Source)
	if err != nil {
		concourse.Fatal("get step failed", err)
	}

	res, err := spinClient.GetPipelineExecutionRaw(request.Version.Ref)
	if err != nil {
		concourse.Fatal("get step failed", err)
	}

	dest := os.Args[1]

	err = ioutil.WriteFile(filepath.Join(dest, "metadata.json"), res, 0644)
	if err != nil {
		concourse.Fatal("get step failed", err)
	}

	err = ioutil.WriteFile(filepath.Join(dest, "version"), []byte(request.Version.Ref), 0644)
	if err != nil {
		concourse.Fatal("get step failed", err)
	}

	var metaData concourse.IntermediateMetadata
	err = json.Unmarshal(res, &metaData)
	if err != nil {
		concourse.Fatal("get step failed", err)
	}

	resArr := []concourse.InResponseMetadata{
		concourse.InResponseMetadata{
			Name:  "Application Name",
			Value: metaData.ApplicationName,
		},
		concourse.InResponseMetadata{
			Name:  "Pipeline Name",
			Value: metaData.PipelineName,
		},
		concourse.InResponseMetadata{
			Name:  "Status",
			Value: metaData.Status,
		},
		concourse.InResponseMetadata{
			Name:  "Start time",
			Value: time.Unix(metaData.StartTime/1000, 0).Format(time.UnixDate),
		},
		concourse.InResponseMetadata{
			Name:  "End time",
			Value: time.Unix(metaData.EndTime/1000, 0).Format(time.UnixDate),
		},
	}

	InResponse := concourse.InResponse{
		Version:  request.Version,
		Metadata: resArr,
	}

	concourse.WriteResponse(InResponse)

}
