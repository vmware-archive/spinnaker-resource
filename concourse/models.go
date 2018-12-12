package concourse

import "time"

type Source struct {
	SpinnakerAPI         string        `json:"spinnaker_api"`
	SpinnakerApplication string        `json:"spinnaker_application"`
	SpinnakerPipeline    string        `json:"spinnaker_pipeline"`
	Statuses             []string      `json:"statuses"`
	StatusCheckTimeout   time.Duration `json:"status-check-timeout"`
	StatusCheckInterval  time.Duration `json:"status-check-interval"`
	X509Cert             string        `json:"spinnaker_x509_cert"`
	X509Key              string        `json:"spinnaker_x509_key"`
}

type Version struct {
	Ref string `json:"ref"`
}

type MetadataPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type OutParams struct {
	TriggerParams             map[string]string `json:"trigger_params,omitempty"`      // optional
	Artifacts                 string            `json:"artifacts"`                     // optional
	TriggerParamsJSONFilePath string            `json:"trigger_params_json_file_path"` //optional
}

type CheckRequest struct {
	Source  Source `json:"source"`
	Version `json:"version"`
}
type InRequest struct {
	Source  Source    `json:"source"`
	Version Version   `json:"version"`
	Params  OutParams `json:"params"`
}
type OutRequest struct {
	Source Source    `json:"source"`
	Params OutParams `json:"params"`
}

type OutResponse struct {
	Version  Version        `json:"version"`
	Metadata []MetadataPair `json:"metadata"`
}

type CheckResponse []Version

type InResponseMetadataKV struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type IntermediateMetadata struct {
	PipelineName    string `json:"name"`
	ApplicationName string `json:"application"`
	StartTime       int64  `json:"startTime"`
	EndTime         int64  `json:"endTime"`
	Status          string `json:"status"`
}

type InResponse struct {
	Version  `json:"version"`
	Metadata []InResponseMetadataKV `json:"metadata"`
}
