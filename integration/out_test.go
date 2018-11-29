package integration_test

import (
	"bytes"
	"encoding/json"
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
