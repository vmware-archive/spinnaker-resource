/*
Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.

This program and the accompanying materials are made available under the terms of the under the Apache License, Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
*/
package concourse

type Source struct {
	SpinnakerAPI         string   `json:"spinnaker_api"`
	SpinnakerApplication string   `json:"spinnaker_application"`
	SpinnakerPipeline    string   `json:"spinnaker_pipeline"`
	Statuses             []string `json:"statuses"`
	StatusCheckTimeout   string   `json:"status_check_timeout"`
	StatusCheckInterval  string   `json:"status_check_interval"`
	X509Cert             string   `json:"spinnaker_x509_cert"`
	X509Key              string   `json:"spinnaker_x509_key"`
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

type InResponseMetadata struct {
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
	Metadata []InResponseMetadata `json:"metadata"`
}
