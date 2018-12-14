package integration_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"

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
		outResponse                   concourse.OutResponse
		inputSource                   concourse.Source
		inputParams                   concourse.OutParams
	)
	BeforeEach(func() {
		pipelineName = "foo"
		applicationName = "bar"
		inputSource = concourse.Source{
			SpinnakerAPI:         spinnakerServer.URL(),
			SpinnakerApplication: applicationName,
			SpinnakerPipeline:    pipelineName,
			X509Cert:             serverCert,
			X509Key:              serverKey,
		}
		pipelineExecutionID = "ABC123"

		spinnakerServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+inputSource.SpinnakerApplication)),
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
				ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+inputSource.SpinnakerApplication+"/pipelineConfigs")),
				ghttp.RespondWithJSONEncoded(
					200,
					[]map[string]string{
						{"name": pipelineName},
					},
				)),
		)
	})
	JustBeforeEach(func() {
		input = concourse.OutRequest{
			Source: inputSource,
			Params: inputParams,
		}
		marshalledInput, err = json.Marshal(input)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("when Spinnaker responds with a status code 202 accepted pipeline execution", func() {
		var httpPOSTSuccessHandler http.HandlerFunc
		BeforeEach(func() {
			httpPOSTSuccessHandler = ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", MatchRegexp(".*/pipelines/"+inputSource.SpinnakerApplication+"/"+pipelineName+".*")),
				ghttp.RespondWithJSONEncoded(
					202,
					map[string]string{
						"ref": "/pipelines/" + pipelineExecutionID,
					},
				),
			)
			// spinnakerServer.AppendHandlers(httpPOSTSuccessHandler)
		})

		Context("when no concourse params are defined", func() {
			BeforeEach(func() {
				spinnakerServer.AppendHandlers(httpPOSTSuccessHandler)
			})
			It("returns the pipeline execution id", func() {
				cmd := exec.Command(outPath, "")
				cmd.Stdin = bytes.NewBuffer(marshalledInput)
				outSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				<-outSess.Exited
				Expect(outSess.ExitCode()).To(Equal(0))

				err = json.Unmarshal(outSess.Out.Contents(), &outResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(outResponse.Version.Ref).To(Equal(pipelineExecutionID))
			})
		})

		Context("when artifacts are defined", func() {
			BeforeEach(func() {
				postBody := `{"type":"concourse-resource","artifacts":[{"foo":"bar"}]}`
				httpPOSTSuccessHandler = ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", MatchRegexp(".*/pipelines/"+inputSource.SpinnakerApplication+"/"+pipelineName+".*")),
					ghttp.VerifyJSON(postBody),
					ghttp.RespondWithJSONEncoded(
						202,
						map[string]string{
							"ref": "/pipelines/" + pipelineExecutionID,
						},
					),
				)
				spinnakerServer.AppendHandlers(httpPOSTSuccessHandler)

				dir, err := ioutil.TempDir("", "location_for_artifact")
				Expect(err).ToNot(HaveOccurred())

				stringArtifact := `[{"foo":"bar"}]`
				err = ioutil.WriteFile(dir+"/my-artifact.json", []byte(stringArtifact), 0644)
				Expect(err).ToNot(HaveOccurred())

				inputParams = concourse.OutParams{
					Artifacts: dir + "/my-artifact.json",
				}
			})

			It("calls Spinnaker API with the artifacts in the post body", func() {

				cmd := exec.Command(outPath, "")
				cmd.Stdin = bytes.NewBuffer(marshalledInput)
				outSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				<-outSess.Exited
				Expect(outSess.ExitCode()).To(Equal(0))

				err = json.Unmarshal(outSess.Out.Contents(), &outResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(outResponse.Version.Ref).To(Equal(pipelineExecutionID))

			})
		})

		Context("when json file trigger params are defined", func() {
			BeforeEach(func() {
				postBody := `{"type":"concourse-resource","parameters":{"foo":"bar", "foobar": "bazbar"}}`
				httpPOSTSuccessHandler = ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", MatchRegexp(".*/pipelines/"+inputSource.SpinnakerApplication+"/"+pipelineName+".*")),
					ghttp.VerifyJSON(postBody),
					ghttp.RespondWithJSONEncoded(
						202,
						map[string]string{
							"ref": "/pipelines/" + pipelineExecutionID,
						},
					),
				)
				spinnakerServer.AppendHandlers(httpPOSTSuccessHandler)

				dir, err := ioutil.TempDir("", "location_for_params")
				Expect(err).ToNot(HaveOccurred())

				stringArtifact := `{"foo":"bar", "foobar": "bazbar"}`
				err = ioutil.WriteFile(dir+"/my-trigger-params.json", []byte(stringArtifact), 0644)
				Expect(err).ToNot(HaveOccurred())

				inputParams = concourse.OutParams{
					TriggerParamsJSONFilePath: dir + "/my-trigger-params.json",
				}
			})

			It("calls Spinnaker API with the contents of the json file as trigger params in the post body", func() {
				cmd := exec.Command(outPath, "")
				cmd.Stdin = bytes.NewBuffer(marshalledInput)
				outSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				<-outSess.Exited
				Expect(outSess.ExitCode()).To(Equal(0))

				err = json.Unmarshal(outSess.Out.Contents(), &outResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(outResponse.Version.Ref).To(Equal(pipelineExecutionID))

			})
		})

		Context("when trigger params are defined", func() {
			BeforeEach(func() {
				postBody := `{"type":"concourse-resource","parameters":{"foo":"bar", "foobar": "bazbar"}}`
				httpPOSTSuccessHandler = ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", MatchRegexp(".*/pipelines/"+inputSource.SpinnakerApplication+"/"+pipelineName+".*")),
					ghttp.VerifyJSON(postBody),
					ghttp.RespondWithJSONEncoded(
						202,
						map[string]string{
							"ref": "/pipelines/" + pipelineExecutionID,
						},
					),
				)
				spinnakerServer.AppendHandlers(httpPOSTSuccessHandler)

				inputParams = concourse.OutParams{
					TriggerParams: map[string]string{
						"foo":    "bar",
						"foobar": "$BAZ",
					},
				}
			})

			It("calls Spinnaker API with the trigger params in the post body", func() {
				cmd := exec.Command(outPath, "")
				cmd.Env = []string{"BAZ=bazbar"}
				cmd.Stdin = bytes.NewBuffer(marshalledInput)
				outSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				<-outSess.Exited
				Expect(outSess.ExitCode()).To(Equal(0))

				err = json.Unmarshal(outSess.Out.Contents(), &outResponse)
				Expect(err).ToNot(HaveOccurred())
				Expect(outResponse.Version.Ref).To(Equal(pipelineExecutionID))

			})
		})

		Context("when status is defined", func() {
			BeforeEach(func() {
				inputSource.Statuses = []string{"SUCCEEDED"}
				inputSource.StatusCheckInterval = "200ms"
				spinnakerServer.AppendHandlers(httpPOSTSuccessHandler)
			})

			Context("when a timeout is specified and Spinnaker pipeline doesn't reach the desired state within the timeout duration", func() {
				BeforeEach(func() {
					inputSource.StatusCheckTimeout = "500ms"

					runningHandler := ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineExecutionID+".*")),
						ghttp.RespondWithJSONEncoded(
							200,
							map[string]string{
								"id":     pipelineExecutionID,
								"status": "RUNNING",
							},
						),
					)
					spinnakerServer.AppendHandlers(
						runningHandler,
						runningHandler,
						runningHandler,
					)
				})

				It("print a '.' for every check and times out and exits with a non zero status and prints an error message", func() {
					cmd := exec.Command(outPath, "")
					cmd.Stdin = bytes.NewBuffer(marshalledInput)
					outSess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).ToNot(HaveOccurred())
					Eventually(outSess.Exited).Should(BeClosed())
					Expect(outSess.ExitCode()).To(Equal(1))

					Expect(outSess.Err).To(gbytes.Say("\\.\\.\n"))
					Expect(outSess.Err).To(gbytes.Say("error put step failed: "))
					Expect(outSess.Err).To(gbytes.Say("timed out waiting for configured status\\(es\\)"))
				})
			})

			Context("when a status is specified, and an unexpected final status reached", func() {
				BeforeEach(func() {
					statusHandlers := []http.HandlerFunc{
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineExecutionID+".*")),
							ghttp.RespondWithJSONEncoded(
								200,
								map[string]string{
									"id":     pipelineExecutionID,
									"status": "RUNNING",
								},
							),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineExecutionID+".*")),
							ghttp.RespondWithJSONEncoded(
								200,
								map[string]string{
									"id":     pipelineExecutionID,
									"status": "TERMINAL",
								},
							),
						),
					}
					spinnakerServer.AppendHandlers(statusHandlers...)
				})

				It("exits with non zero code and prints an error message", func() {
					cmd := exec.Command(outPath, "")
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

			Context("when a status is specified, and reached", func() {
				BeforeEach(func() {
					spinnakerServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineExecutionID+".*")),
							ghttp.RespondWithJSONEncoded(
								200,
								map[string]string{
									"id":     pipelineExecutionID,
									"status": "RUNNING",
								},
							),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineExecutionID+".*")),
							ghttp.RespondWithJSONEncoded(
								200,
								map[string]string{
									"id":     pipelineExecutionID,
									"status": "SUCCEEDED",
								},
							),
						),
					)
				})
				It("waits till the pipeline execution status is satisfied and returns the pipeline execution id", func() {
					cmd := exec.Command(outPath, "")
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
	})

	Context("when Spinnaker responds with status code 4xx on a POST for a pipeline execution", func() {
		var statusCode int
		BeforeEach(func() {

			statusCode = 422
			responseMap = map[string]string{
				"message": "500 ",
			}
			httpPOSTFailureHandler := ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", MatchRegexp(".*/pipelines/"+inputSource.SpinnakerApplication+"/"+pipelineName+".*")),
				ghttp.RespondWithJSONEncoded(
					statusCode,
					responseMap,
				),
			)
			spinnakerServer.AppendHandlers(httpPOSTFailureHandler)

		})

		It("prints the status code, response body and exits with exit code 1", func() {
			cmd := exec.Command(outPath, "")
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
