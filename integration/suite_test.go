/*
Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.

This program and the accompanying materials are made available under the terms of the under the Apache License, Version 2.0 (the "License”); you may not use this file except in compliance with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
*/
package integration_test

import (
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
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
	outPath, checkPath, inPath string
	spinnakerServer            *ghttp.Server
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	outBinPath, err := gexec.Build("github.com/pivotal-cf/spinnaker-resource/cmd/out")
	Expect(err).NotTo(HaveOccurred())
	checkBinPath, err := gexec.Build("github.com/pivotal-cf/spinnaker-resource/cmd/check")
	Expect(err).NotTo(HaveOccurred())
	inBinPath, err := gexec.Build("github.com/pivotal-cf/spinnaker-resource/cmd/in")
	Expect(err).NotTo(HaveOccurred())

	return []byte(outBinPath + "," + checkBinPath + "," + inBinPath)
}, func(data []byte) {
	paths := strings.Split(string(data), ",")
	outPath = paths[0]
	checkPath = paths[1]
	inPath = paths[2]

	SetDefaultEventuallyTimeout(10 * time.Second)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	spinnakerServer = ghttp.NewServer()
})

var _ = AfterEach(func() {
	spinnakerServer.Close()
})
