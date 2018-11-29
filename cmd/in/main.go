package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
	"github.com/pivotal-cf/spinnaker-resource/spinnaker"
)

func main() {

	var request concourse.InRequest
	concourse.ReadInRequest(&request)

	spinClient, err := spinnaker.NewClient(request.Source)
	concourse.Check("get", err)

	res, err := spinClient.GetPipelineRaw(request.Version.Ref)
	concourse.Check("get", err)

	err = ioutil.WriteFile(filepath.Join(os.Getenv("1"), "metadata.json"), res, 0644)
	concourse.Check("get", err)

	err = ioutil.WriteFile(filepath.Join(os.Getenv("1"), "version"), []byte(request.Version.Ref), 0644)
	concourse.Check("get", err)

	var metaData concourse.IntermediateMetadata
	err = json.Unmarshal(res, &metaData)
	concourse.Check("get", err)

	resArr := []concourse.InResponseMetadataKV{
		concourse.InResponseMetadataKV{
			Name:  "Application Name",
			Value: metaData.ApplicationName,
		},
		concourse.InResponseMetadataKV{
			Name:  "Pipeline Name",
			Value: metaData.PipelineName,
		},
		concourse.InResponseMetadataKV{
			Name:  "Status",
			Value: metaData.Status,
		},
		concourse.InResponseMetadataKV{
			Name:  "Start time",
			Value: time.Unix(metaData.StartTime, 0).Format(time.UnixDate),
		},
		concourse.InResponseMetadataKV{
			Name:  "End time",
			Value: time.Unix(metaData.EndTime, 0).Format(time.UnixDate),
		},
	}

	InResponse := concourse.InResponse{
		Version:  request.Version,
		Metadata: resArr,
	}

	concourse.WriteInResponse(InResponse)

}
