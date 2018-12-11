package integration_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
)

var _ = Describe("In", func() {
	var (
		applicationName, pipelineName string
		input                         concourse.InRequest
		marshalledInput               []byte
		err                           error
		statusCode                    int
		pipelineID                    string
		allHandler                    http.HandlerFunc
		inSess                        *gexec.Session
		dir                           string
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
			allHandler,
		)
		input = concourse.InRequest{
			Source: concourse.Source{
				SpinnakerAPI:         spinnakerServer.URL(),
				SpinnakerApplication: applicationName,
				SpinnakerPipeline:    pipelineName,
				X509Cert:             serverCert,
				X509Key:              serverKey,
			},
			Version: concourse.Version{
				Ref: pipelineID,
			},
		}

		marshalledInput, err = json.Marshal(input)
		Expect(err).ToNot(HaveOccurred())

		dir, err = ioutil.TempDir("", "location_for_metadata")
		Expect(err).ToNot(HaveOccurred())
		cmd := exec.Command(inPath, dir)
		cmd.Stdin = bytes.NewBuffer(marshalledInput)
		inSess, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		<-inSess.Exited
	})

	Context("when the pipeline exists and spinnaker all response data", func() {
		var (
			mappedRes map[string]interface{}
		)

		BeforeEach(func() {
			statusCode = 200
			pipelineID = "goodID"

			expectedResBytes, err := ioutil.ReadFile("./fixtures/get_pipelines_response.json")
			Expect(err).ToNot(HaveOccurred())

			json.Unmarshal(expectedResBytes, &mappedRes)

			allHandler = ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineID)),
				ghttp.RespondWithJSONEncoded(
					statusCode,
					mappedRes,
				),
			)
		})

		It("stores the metadata into a JSON file, in the resource's volume", func() {
			defer os.RemoveAll(dir)

			Expect(inSess.ExitCode()).To(Equal(0))

			Expect(filepath.Join(dir, "metadata.json")).To(BeAnExistingFile())
			Expect(filepath.Join(dir, "version")).To(BeAnExistingFile())

			actualResBytes, err := ioutil.ReadFile(filepath.Join(dir, "metadata.json"))

			var actualResMap map[string]interface{}
			err = json.Unmarshal([]byte(actualResBytes), &actualResMap)
			Expect(err).ToNot(HaveOccurred())

			Expect(actualResMap).To(Equal(mappedRes))

			actualVersionBytes, err := ioutil.ReadFile(filepath.Join(dir, "version"))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(actualVersionBytes)).To(Equal(pipelineID))
		})

		It("returns the version and metadata", func() {
			defer os.RemoveAll(dir)

			Expect(inSess.ExitCode()).To(Equal(0))

			//Values here match the fixture file
			expArr := []concourse.InResponseMetadataKV{
				concourse.InResponseMetadataKV{
					Name:  "Application Name",
					Value: "nvidia",
				},
				concourse.InResponseMetadataKV{
					Name:  "Pipeline Name",
					Value: "bar",
				},
				concourse.InResponseMetadataKV{
					Name:  "Status",
					Value: "SUCCEEDED",
				},
				concourse.InResponseMetadataKV{
					Name:  "Start time",
					Value: time.Unix(1543414041364/1000, 0).Format(time.UnixDate),
				},
				concourse.InResponseMetadataKV{
					Name:  "End time",
					Value: time.Unix(1543414041439/1000, 0).Format(time.UnixDate),
				},
			}

			var inResponse concourse.InResponse
			expectedResponse := concourse.InResponse{
				Version: concourse.Version{
					Ref: pipelineID,
				},
				Metadata: expArr,
			}
			err = json.Unmarshal(inSess.Out.Contents(), &inResponse)
			Expect(err).ToNot(HaveOccurred())
			Expect(inResponse).To(Equal(expectedResponse))
		})
	})

	Context("when the status code > 400", func() {
		Context("when the status code is not 404", func() {
			BeforeEach(func() {
				statusCode = 500
				pipelineID = "badID"

				allHandler = ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineID)),
					ghttp.RespondWithJSONEncoded(
						statusCode,
						map[string]interface{}{
							"error":     "Internal Server Error",
							"exception": "org.springframework.web.method.annotation.MethodArgumentTypeMismatchException",
							"message":   "something bad happend",
							"status":    statusCode,
							"timestamp": 1543422807176,
						},
					),
				)
			})

			It("errors with a pipeline not found error, exits with exit code 1", func() {
				Expect(inSess.ExitCode()).To(Equal(1))

				Expect(inSess.Err).Should(gbytes.Say("error get step failed:"))
				Expect(inSess.Err).Should(gbytes.Say("spinnaker api responded with status code: " + strconv.Itoa(statusCode)))
			})
		})

		Context("when the status code is 404", func() {
			BeforeEach(func() {
				statusCode = 404
				pipelineID = "badID"

				allHandler = ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", MatchRegexp(".*/pipelines/"+pipelineID)),
					ghttp.RespondWithJSONEncoded(
						statusCode,
						map[string]interface{}{
							"error":     "Not Found",
							"exception": "com.netflix.spinnaker.kork.web.exceptions.NotFoundException",
							"message":   "Pipeline not found (id: " + pipelineID + ")",
							"status":    statusCode,
							"timestamp": 1543336865332,
						},
					),
				)
			})
			It("prints the status code, response body and exits with exit code 1", func() {
				Expect(inSess.ExitCode()).To(Equal(1))

				Expect(inSess.Err).Should(gbytes.Say("error get step failed: "))
				Expect(inSess.Err).Should(gbytes.Say("pipeline execution ID not found \\(ID: " + pipelineID + "\\)\n"))
			})
		})
	})
})
