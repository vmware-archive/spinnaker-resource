package integration_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
)

var _ = Describe("Out", func() {
	var (
		applicationName, pipelineName string
		pipelineExecutionID           string
		responseMap                   map[string]string
		input                         concourse.OutRequest
		marshalledInput               []byte
		err                           error
		statusCode                    int
		outResponse                   concourse.OutResponse
	)

	JustBeforeEach(func() {
		spinnakerServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+applicationName)),
				ghttp.RespondWithJSONEncoded(
					200,
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
					200,
					[]map[string]string{
						{"name": pipelineName},
					},
				)),
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", MatchRegexp(".*/pipelines/"+applicationName+"/"+pipelineName+".*")),
				ghttp.RespondWithJSONEncoded(
					statusCode,
					responseMap,
				),
			),
		)
		input = concourse.OutRequest{
			Source: concourse.Source{
				SpinnakerAPI:         spinnakerServer.URL(),
				SpinnakerApplication: applicationName,
				SpinnakerPipeline:    pipelineName,
				X509Cert:             serverCert,
				X509Key:              serverKey,
			},
		}
		marshalledInput, err = json.Marshal(input)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when Spinnaker responds with an accepted pipeline execution", func() {
		BeforeEach(func() {
			pipelineName = "foo"
			applicationName = "bar"
			pipelineExecutionID = "ABC123"
			responseMap = map[string]string{
				"ref": "/pipelines/" + pipelineExecutionID,
			}
			statusCode = 202
		})

		It("returns the pipeline execution id", func() {
			cmd := exec.Command(outPath)
			cmd.Stdin = bytes.NewBuffer(marshalledInput)
			outSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			<-outSess.Exited
			Expect(outSess.ExitCode()).To(Equal(0))

			err = json.Unmarshal(outSess.Out.Contents(), &outResponse)
			Expect(err).ToNot(HaveOccurred())
			Expect(outResponse.Version.Ref).To(Equal(pipelineExecutionID))

		})
		//TODO needs drying up
		Context("when a status and timeout are specified and Spinnaker pipeline doesn't reach the desired state within the timeout duration", func() {
			JustBeforeEach(func() {
				input = concourse.OutRequest{
					Source: concourse.Source{
						SpinnakerAPI:         spinnakerServer.URL(),
						SpinnakerApplication: applicationName,
						SpinnakerPipeline:    pipelineName,
						X509Cert:             serverCert,
						X509Key:              serverKey,
						Statuses:             []string{"SUCCEEDED"},
						StatusCheckTimeout:   500 * time.Millisecond,
						StatusCheckInterval:  200 * time.Millisecond,
					},
				}
				runningHandler := ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineExecutionID+".*")),
					ghttp.RespondWithJSONEncoded(
						statusCode,
						map[string]string{
							"id":     pipelineExecutionID,
							"status": "RUNNING",
						},
					),
				)
				marshalledInput, err = json.Marshal(input)
				Expect(err).ToNot(HaveOccurred())
				spinnakerServer.AppendHandlers(
					runningHandler,
					runningHandler,
					runningHandler,
				)
			})

			It("times out and exits with a non zero status and prints an error message", func() {
				cmd := exec.Command(outPath)
				cmd.Stdin = bytes.NewBuffer(marshalledInput)
				outSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(outSess.Exited).Should(BeClosed())
				// Expect(spinnakerServer.ReceivedRequests()).Should(HaveLen(6))
				Expect(outSess.ExitCode()).To(Equal(1))

				Expect(outSess.Err).To(gbytes.Say("error put step failed: "))
				Expect(outSess.Err).To(gbytes.Say("timed out waiting for configured status\\(es\\)"))
			})
		})
		//TODO needs drying up
		Context("when a status is specified, and an unexpected final status reached", func() {
			JustBeforeEach(func() {
				input = concourse.OutRequest{
					Source: concourse.Source{
						SpinnakerAPI:         spinnakerServer.URL(),
						SpinnakerApplication: applicationName,
						SpinnakerPipeline:    pipelineName,
						X509Cert:             serverCert,
						X509Key:              serverKey,
						Statuses:             []string{"SUCCEEDED"},
						StatusCheckInterval:  200 * time.Millisecond,
					},
				}
				marshalledInput, err = json.Marshal(input)
				Expect(err).ToNot(HaveOccurred())
				spinnakerServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineExecutionID+".*")),
						ghttp.RespondWithJSONEncoded(
							statusCode,
							map[string]string{
								"id":     pipelineExecutionID,
								"status": "RUNNING",
							},
						),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineExecutionID+".*")),
						ghttp.RespondWithJSONEncoded(
							statusCode,
							map[string]string{
								"id":     pipelineExecutionID,
								"status": "TERMINAL",
							},
						),
					),
				)
			})

			It("exits with non zero code and prints an error message", func() {
				cmd := exec.Command(outPath)
				cmd.Stdin = bytes.NewBuffer(marshalledInput)
				outSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				<-outSess.Exited
				Expect(spinnakerServer.ReceivedRequests()).Should(HaveLen(5))
				Expect(outSess.ExitCode()).To(Equal(1))

				Expect(outSess.Err).To(gbytes.Say("error put step failed:"))
				Expect(outSess.Err).To(gbytes.Say("Pipeline execution reached a final state: TERMINAL"))
			})
		})

		//TODO needs drying up
		Context("when a status is specified, and reached", func() {
			JustBeforeEach(func() {
				input = concourse.OutRequest{
					Source: concourse.Source{
						SpinnakerAPI:         spinnakerServer.URL(),
						SpinnakerApplication: applicationName,
						SpinnakerPipeline:    pipelineName,
						X509Cert:             serverCert,
						X509Key:              serverKey,
						Statuses:             []string{"SUCCEEDED"},
						StatusCheckInterval:  200 * time.Millisecond,
					},
				}
				marshalledInput, err = json.Marshal(input)
				Expect(err).ToNot(HaveOccurred())
				spinnakerServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineExecutionID+".*")),
						ghttp.RespondWithJSONEncoded(
							statusCode,
							map[string]string{
								"id":     pipelineExecutionID,
								"status": "RUNNING",
							},
						),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineExecutionID+".*")),
						ghttp.RespondWithJSONEncoded(
							statusCode,
							map[string]string{
								"id":     pipelineExecutionID,
								"status": "SUCCEEDED",
							},
						),
					),
				)
			})
			It("waits till the pipeline execution status is satisfied and returns the pipeline execution id", func() {
				cmd := exec.Command(outPath)
				cmd.Stdin = bytes.NewBuffer(marshalledInput)
				outSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				<-outSess.Exited
				Expect(spinnakerServer.ReceivedRequests()).Should(HaveLen(5))
				Expect(outSess.ExitCode()).To(Equal(0))

				err = json.Unmarshal(outSess.Out.Contents(), &outResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(outResponse.Version.Ref).To(Equal(pipelineExecutionID))
			})
		})
	})

	Context("when spinnaker responds with 4xx", func() {
		BeforeEach(func() {
			pipelineName = "foo"
			applicationName = "bar"
			responseMap = map[string]string{
				"message": "500 ",
			}
			statusCode = 422
		})

		It("prints the status code, response body and exits with exit code 1", func() {
			cmd := exec.Command(outPath)
			cmd.Stdin = bytes.NewBuffer(marshalledInput)
			outSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			<-outSess.Exited
			Expect(outSess.ExitCode()).To(Equal(1))

			Expect(outSess.Err).To(gbytes.Say("error put step failed:"))
			Expect(outSess.Err).To(gbytes.Say("spinnaker api responded with status code: " + strconv.Itoa(statusCode)))
			responseString, err := json.Marshal(responseMap)
			Expect(err).ToNot(HaveOccurred())
			Expect(outSess.Err).To(gbytes.Say("body: " + string(responseString)))
		})
	})
})
