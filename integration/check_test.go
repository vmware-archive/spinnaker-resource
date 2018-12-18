/*
Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.

This program and the accompanying materials are made available under the terms of the under the Apache License, Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
*/
package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
)

var _ = Describe("Check", func() {
	var (
		applicationName, pipelineName string
		responseMap                   []map[string]interface{}
		input                         concourse.CheckRequest
		marshalledInput               []byte
		err                           error
		statusCode                    int
		pipelineExecutions            []map[string]interface{}
		checkResponse                 []concourse.Version
		allHandler                    http.HandlerFunc
		inputRef                      string
		checkSess                     *gexec.Session
		statuses                      []string
	)
	pipelineName = "foo"
	applicationName = "bar"
	pipelineExecutions = []map[string]interface{}{
		map[string]interface{}{
			"id":        "EX1",
			"name":      pipelineName,
			"buildTime": 1543244670,
			"status":    "SUCCEEDED",
		},
		map[string]interface{}{
			"id":        "EX2",
			"name":      pipelineName,
			"buildTime": 1543244680,
			"status":    "SUCCEEDED",
		},
		map[string]interface{}{
			"id":        "EX3",
			"name":      pipelineName,
			"buildTime": 1543244690,
			"status":    "TERMINAL",
		},
		map[string]interface{}{
			"id":        "EX4",
			"name":      "other-pipeline",
			"buildTime": 1543244690,
			"status":    "SUCCEEDED",
		},
		map[string]interface{}{
			"id":        "EX5",
			"name":      "other-pipeline",
			"buildTime": 1543244690,
			"status":    "SUCCEEDED",
		},
	}
	JustBeforeEach(func() {
		spinnakerServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+applicationName)),
				ghttp.RespondWithJSONEncoded(
					statusCode,
					map[string]interface{}{
						"attributes": map[string]interface{}{
							"accounts": nil,
							"name":     applicationName,
						},
						"clusters": nil,
						"name":     applicationName,
					},
				)),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+applicationName+"/pipelineConfigs")),
				ghttp.RespondWithJSONEncoded(
					statusCode,
					[]map[string]string{
						{"name": pipelineName},
					},
				)),
			allHandler,
		)
		input = concourse.CheckRequest{
			Source: concourse.Source{
				SpinnakerAPI:         spinnakerServer.URL(),
				SpinnakerApplication: applicationName,
				SpinnakerPipeline:    pipelineName,
				Statuses:             statuses,
				X509Cert:             serverCert,
				X509Key:              serverKey,
			},
			Version: concourse.Version{
				Ref: inputRef,
			},
		}
		marshalledInput, err = json.Marshal(input)
		Expect(err).ToNot(HaveOccurred())
		cmd := exec.Command(checkPath)
		cmd.Stdin = bytes.NewBuffer(marshalledInput)
		checkSess, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		<-checkSess.Exited
	})
	Context("when input version is not empty", func() {
		BeforeEach(func() {
			statusCode = 200
			allHandler = ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+applicationName+"/pipelines"), "limit=25"),
				ghttp.RespondWithJSONEncoded(
					statusCode,
					pipelineExecutions,
				),
			)
		})
		Context("when statuses are specified in the resource params", func() {
			BeforeEach(func() {
				inputRef = pipelineExecutions[0]["id"].(string)
				statuses = []string{"SUCCEEDED"}
			})

			It("returns the versions that match the provided statuses", func() {
				Expect(checkSess.ExitCode()).To(Equal(0))

				err = json.Unmarshal(checkSess.Out.Contents(), &checkResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(checkResponse)).To(Equal(2))
				Expect(checkResponse[0].Ref).To(Equal(pipelineExecutions[0]["id"].(string)))
				Expect(checkResponse[1].Ref).To(Equal(pipelineExecutions[1]["id"].(string)))
			})
		})
		Context("when statuses are not specified in the resource params", func() {
			Context("when input version exists but not the latest version", func() {
				BeforeEach(func() {
					inputRef = pipelineExecutions[1]["id"].(string)
					statuses = []string{}
				})
				It("returns the input version and every version that follows", func() {
					Expect(checkSess.ExitCode()).To(Equal(0))

					err = json.Unmarshal(checkSess.Out.Contents(), &checkResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(checkResponse)).To(Equal(2))
					Expect(checkResponse[0].Ref).To(Equal(pipelineExecutions[1]["id"].(string)))
					Expect(checkResponse[1].Ref).To(Equal(pipelineExecutions[2]["id"].(string)))
				})
			})

			Context("when input version is the latest version", func() {
				BeforeEach(func() {
					inputRef = pipelineExecutions[2]["id"].(string)
					statuses = []string{}
				})
				It("returns the only the input version", func() {
					Expect(checkSess.ExitCode()).To(Equal(0))

					err = json.Unmarshal(checkSess.Out.Contents(), &checkResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(checkResponse)).To(Equal(1))
					Expect(checkResponse[0].Ref).To(Equal(pipelineExecutions[2]["id"].(string)))
				})
			})
			Context("when input version doesn't exist anymore", func() {
				BeforeEach(func() {
					responseMap = []map[string]interface{}{
						pipelineExecutions[1],
						pipelineExecutions[2],
						pipelineExecutions[3],
					}
					allHandler = ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+applicationName+"/pipelines"), "limit=25"),
						ghttp.RespondWithJSONEncoded(
							statusCode,
							responseMap,
						),
					)
					inputRef = pipelineExecutions[0]["id"].(string)
					statuses = []string{}
				})
				It("returns the only the input version", func() {
					Expect(checkSess.ExitCode()).To(Equal(0))

					err = json.Unmarshal(checkSess.Out.Contents(), &checkResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(checkResponse)).To(Equal(1))
					Expect(checkResponse[0].Ref).To(Equal(pipelineExecutions[2]["id"].(string)))
				})
			})
		})
	})
	Context("when input version is empty", func() {
		BeforeEach(func() {
			inputRef = ""
			pipelineName = "foo"
			applicationName = "bar"
			responseMap = []map[string]interface{}{
				pipelineExecutions[0],
				pipelineExecutions[1],
				pipelineExecutions[2],
				pipelineExecutions[3],
			}
			statuses = []string{}
			statusCode = 200
			allHandler = ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+applicationName+"/pipelines"), "limit=25"),
				ghttp.RespondWithJSONEncoded(
					statusCode,
					responseMap,
				),
			)
		})
		Context("when statuses are specified", func() {
			BeforeEach(func() {
				statuses = []string{"SUCCEEDED"}

			})
			It("returns the only the latest version that mathces the specified statuses to stdout", func() {
				Expect(checkSess.ExitCode()).To(Equal(0))

				err = json.Unmarshal(checkSess.Out.Contents(), &checkResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(checkResponse)).To(Equal(1))
				Expect(checkResponse[0].Ref).To(Equal(pipelineExecutions[1]["id"].(string)))
			})

			Context("when pipeline executions does not have the status we are looking for", func() {
				BeforeEach(func() {
					statuses = []string{"FAILED"}
				})

				It("returns no versions", func() {
					Expect(checkSess.ExitCode()).To(Equal(0))

					err = json.Unmarshal(checkSess.Out.Contents(), &checkResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(checkResponse)).To(Equal(0))
				})
			})
		})
		Context("when statuses are not specified", func() {
			BeforeEach(func() {
				statuses = []string{}
			})

			Context("when pipeline executions does not have the status we are looking for", func() {
				BeforeEach(func() {
					statusCode = 200

					allHandler = ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+applicationName+"/pipelines"), "limit=25"),
						ghttp.RespondWithJSONEncoded(
							statusCode,
							[]map[string]interface{}{
								pipelineExecutions[3],
								pipelineExecutions[4],
							},
						),
					)
				})

				It("returns no versions", func() {
					Expect(checkSess.ExitCode()).To(Equal(0))

					err = json.Unmarshal(checkSess.Out.Contents(), &checkResponse)
					Expect(err).ToNot(HaveOccurred())
					Expect(len(checkResponse)).To(Equal(0))
				})
			})

			It("returns the only the latest version to stdout", func() {
				Expect(checkSess.ExitCode()).To(Equal(0))

				err = json.Unmarshal(checkSess.Out.Contents(), &checkResponse)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(checkResponse)).To(Equal(1))
				Expect(checkResponse[0].Ref).To(Equal(pipelineExecutions[2]["id"].(string)))
			})
		})
	})
})
