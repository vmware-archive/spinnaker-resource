/*
Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.

This program and the accompanying materials are made available under the terms of the under the Apache License, Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
*/
package spinnaker_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/spinnaker-resource/concourse"
	"github.com/pivotal-cf/spinnaker-resource/spinnaker"
)

const serverCert = `-----BEGIN CERTIFICATE-----
MIIC+DCCAeCgAwIBAgIRAK3uVYcWQA/y8Q8wHWnm0YgwDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0xNjA4MDgyMzExMTFaFw0yNjA4MDYyMzEx
MTFaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQC+qY2Pfr79ltRLudcX45AyUPmOm0DwF2gE8HUihljtQmeWno5Gc2Uc
MqRrs7sfu90geL9ZBU7jYjhFxdlbsIO6710J0+uElLPKgSPI0sJDL3aoIi7jd+mi
qTyQ/OErOxtTOe7V3xUjAS9HrIcqVxKQFGwIic5sIOWhdg5zbVqoCI8eX5QHdxST
zNtoJYeCnCC5P7fdZySZ7lH5Y3HLgQWsVFyqoklKiYVmK1AyOQsZqrfOg1QjkXp9
xKN/Z0EsRsBGItvEnzQdVlaFFdo9yKnuWDzNvdwWJUpH/pdv6SOzvunAhZrNHe8w
DWUeLA6L5E8AvLR9KnT+BBCvygFu8njVAgMBAAGjSTBHMA4GA1UdDwEB/wQEAwIC
pDATBgNVHSUEDDAKBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MA8GA1UdEQQI
MAaHBH8AAAEwDQYJKoZIhvcNAQELBQADggEBAHPBHI8Vx/lD8KIPRBSfY2+XSXKQ
z4dHRFQxC4+hUm5X39Dg++ZgbHf5/Iv3T8466CW3DADCRamEpKmNK0/MAizDRmb2
sQ6qCVO5CrljEPgECY9MIV2MknbRIK6J0WhUEkTNY/RkGyLOkgGFD5Fadorf/b9D
0MKeDOl3xGIoDMz1qGS/ByiUXlu/5Dze3EKigtTI74z8GYIo/eAswfh3sIi0X7KR
vgkHttWh9tkfjV9IxuG/yCAnPTlCN7UI8YTZIH+SPqFakS8cIBzmVlOnZBsH4u2/
wtISX1uF4BH/i+knckhiG5mHNVSOVyUlZvC8lZR2hRMkeXVb/uns66Z/fSE=
-----END CERTIFICATE-----
`

const serverKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAvqmNj36+/ZbUS7nXF+OQMlD5jptA8BdoBPB1IoZY7UJnlp6O
RnNlHDKka7O7H7vdIHi/WQVO42I4RcXZW7CDuu9dCdPrhJSzyoEjyNLCQy92qCIu
43fpoqk8kPzhKzsbUznu1d8VIwEvR6yHKlcSkBRsCInObCDloXYOc21aqAiPHl+U
B3cUk8zbaCWHgpwguT+33Wckme5R+WNxy4EFrFRcqqJJSomFZitQMjkLGaq3zoNU
I5F6fcSjf2dBLEbARiLbxJ80HVZWhRXaPcip7lg8zb3cFiVKR/6Xb+kjs77pwIWa
zR3vMA1lHiwOi+RPALy0fSp0/gQQr8oBbvJ41QIDAQABAoIBAAzM9WQc7lW4Oqia
4YYJETVPmnGomsODzsgGHNckjfPf8XR7ULIKLU+nVsKkXnvS8RWtBavEX3eEsKJ+
lglB4JY8W9K9F6LfGPMPmIdzHvfDyAOhx+QduOHi2t4hHDz6yurbiN1zDMg83B/D
xY9iKSzjMh2gous/iis88dtuDBgb3RV903oiNJmTmHbZiClSEe9r3TWfOlxVMH0B
kFMvsnvRDomDzyfnjDTK+C8fPp07O3/uIM8rbOJaVEBYOVKj/pFlYA0HHY4g+sq5
zYSGzOLJLCVooU8hOYq3DuhYuFliziGDJZx3vg08GKVYwmcBaIlmYxPtFScyliKx
vRTFEgECgYEAwgj0ZUPA/DHyCtydwKUXjCV+j5uQJwDfes2qFDGbhcT6xkoGIM3S
EQl+Lu3NlRXJqZZfyZjurCuK9hOMIWK7Brlm/TyDV5CnOSK86/ez4mOL7mf802zZ
+aMqITebdj1BMLa3IGZhsw4hguLHQ7qelvJpyE/7531OEcyH6AB29MkCgYEA+4zf
BkW1PO7gSAZLU0RA5mkPjLV61OVrL3q7Yfq1sYC+dD/kQ7ug36ElZKLwtnPyPB2D
Yb0fxwDRCAeF2VZE58gVJwC1xtVhVI7DgXRgGdXZZq8EmW5/308mwLov2RfR/4lE
SgQ2gLruVZSt4hqXqmT2CV2UbKwDapDhTEC+Ja0CgYB8s5KWLjguHM9Iycac071R
dZtkIf9AAeCepOTEu6kPDKx6mYJcvMpf5rDw6iYwxWLomdsPzji97/IL+j4aCsDW
LnuRDr3+ndnK75dpM7WpLn71BmHHY3KnbISb+ofwMqfd7d+9c+8gS1mgK60SyzI3
Iq53bWgguzhcWg2SPhI1eQKBgQDJwJODwVb6NxDVU48Iip6O7kaVcVzB8ftEymgN
znn5kquuKyxWEt+VXPbTv0fW3imzg2xDcN9SydndWcNFrEZ5q+UjMhOZFL0Kh7JQ
WtlU/0ptbAQBVzniDeaj/vCvasZ38E1AHB7moobTRvsrdG6eMHmQy2hmvJPE3cyF
TwvyxQKBgEBP37kUkMg2D3JtnKJZX3r0DKf/3fvDXU+nsOipxF+QVi7SMt/30j2d
cGvJLpKX8qu9LOLGoVaB1yxEO8DO5B2YxdG/sjMBLlN+JSMFQ734VEIWD+1LiBvQ
JBf0Z15NGOY3w9KSxeiTPFyXU4/RmymDzyd/VKcnPBMKqTvbqT2G
-----END RSA PRIVATE KEY-----`

var (
	statusCode                        int
	allHandler, pipelineConfigHandler http.HandlerFunc
	applicationName                   string
	pipelineName                      string
	spinnakerServer                   *ghttp.Server
)

var _ = Describe("Spinnaker Client", func() {
	Context("When creating a new spinnaker client", func() {
		JustBeforeEach(func() {
			spinnakerServer = ghttp.NewServer()
			spinnakerServer.AppendHandlers(allHandler, pipelineConfigHandler)
		})

		Context("Given a bad set of certificates", func() {
			It("returns an error", func() {
				source := concourse.Source{
					SpinnakerAPI:         spinnakerServer.URL(),
					SpinnakerApplication: "some_app",
					SpinnakerPipeline:    "some_pipeline",
					X509Cert:             "",
					X509Key:              "",
				}
				_, err := spinnaker.NewClient(source)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("Given an application does not exist", func() {
			BeforeEach(func() {
				applicationName = "nonexistent_app"
				statusCode = 404
				allHandler = ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+applicationName)),
					ghttp.RespondWithJSONEncoded(
						statusCode,
						map[string]interface{}{
							"error":     "Not Found",
							"exception": "com.netflix.spinnaker.kork.web.exceptions.NotFoundException",
							"message":   "Application not found (id: nvidiw)",
							"status":    404,
							"timestamp": 1543336865332,
						},
					),
				)
			})

			It("returns an error indicating an invalid application", func() {
				pipelineName = "nonexistent_pipeline"
				source := concourse.Source{
					SpinnakerAPI:         spinnakerServer.URL(),
					SpinnakerApplication: applicationName,
					SpinnakerPipeline:    pipelineName,
					X509Cert:             serverCert,
					X509Key:              serverKey,
				}
				_, err := spinnaker.NewClient(source)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("spinnaker application " + applicationName + " not found"))
			})
		})

		Context("Given an application exists", func() {
			BeforeEach(func() {
				applicationName = "existent_app"
				statusCode = 200
				allHandler = ghttp.CombineHandlers(
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
					),
				)
			})

			Context("Given an pipeline does not exist", func() {
				BeforeEach(func() {
					pipelineConfigHandler = ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+applicationName+"/pipelineConfigs")),
						ghttp.RespondWithJSONEncoded(
							statusCode,
							[]map[string]string{
								{"name": "existent_pipeline"},
							},
						),
					)
				})
				It("returns an error indicating an invalid pipeline", func() {
					pipelineName = "nonexistent_pipeline"
					source := concourse.Source{
						SpinnakerAPI:         spinnakerServer.URL(),
						SpinnakerApplication: applicationName,
						SpinnakerPipeline:    pipelineName,
						X509Cert:             serverCert,
						X509Key:              serverKey,
					}
					_, err := spinnaker.NewClient(source)

					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("spinnaker pipeline " + pipelineName + " not found"))
				})
			})

			Context("Given an pipeline that exists", func() {
				BeforeEach(func() {
					pipelineConfigHandler = ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", MatchRegexp(".*/applications/"+applicationName+"/pipelineConfigs")),
						ghttp.RespondWithJSONEncoded(
							statusCode,
							[]map[string]interface{}{
								{"name": "existent_pipeline"},
								{"name": "existent_pipeline2"},
							},
						),
					)
				})
				It("returns a new client configured with the pipeline and application", func() {
					pipelineName = "existent_pipeline"
					source := concourse.Source{
						SpinnakerAPI:         spinnakerServer.URL(),
						SpinnakerApplication: applicationName,
						SpinnakerPipeline:    pipelineName,
						X509Cert:             serverCert,
						X509Key:              serverKey,
					}
					_, err := spinnaker.NewClient(source)

					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
	})
})
