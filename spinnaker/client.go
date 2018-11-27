package spinnaker

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf/spinnaker-resource/concourse"
)

type SpinClient struct {
	sourceConfig concourse.Source
	client       *http.Client
}

func NewClient(source concourse.Source) (SpinClient, error) {

	cert, err := tls.X509KeyPair([]byte(source.X509Cert), []byte(source.X509Key))

	if err != nil {
		return SpinClient{}, err
	}
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		Certificates:             []tls.Certificate{cert},
		InsecureSkipVerify:       true,
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{Transport: tr}

	res, err := client.Get(fmt.Sprintf("%s/applications/%s", source.SpinnakerAPI, source.SpinnakerApplication))
	if err != nil {
		return SpinClient{}, err
	} else if res.StatusCode == 404 {
		err = fmt.Errorf("spinnaker application %s not found", source.SpinnakerApplication)
		return SpinClient{}, err
	} else if res.StatusCode >= 400 {
		body, err := ioutil.ReadAll(res.Body)
		if err == nil {
			err = fmt.Errorf("spinnaker api responded with status code: %d, body: %s", res.StatusCode, string(body))
		}
		return SpinClient{}, err
	}

	res, err = client.Get(fmt.Sprintf("%s/applications/%s/pipelineConfigs", source.SpinnakerAPI, source.SpinnakerApplication))
	if err != nil {
		return SpinClient{}, err
	} else if res.StatusCode >= 400 {
		body, err := ioutil.ReadAll(res.Body)
		if err == nil {
			err = fmt.Errorf("spinnaker api responded with status code: %d, body: %s", res.StatusCode, string(body))
			return SpinClient{}, err
		}
	} else {
		var pipelineConfigs []map[string]interface{}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return SpinClient{}, err
		}

		err = json.Unmarshal(body, &pipelineConfigs)
		if err != nil {
			return SpinClient{}, err
		}

		found := false
		for _, pc := range pipelineConfigs {
			if pc["name"].(string) == source.SpinnakerPipeline {
				found = true
				break
			}
		}
		if !found {
			err = fmt.Errorf("spinnaker pipeline %s not found", source.SpinnakerPipeline)
			return SpinClient{}, err
		}
	}

	spinClient := SpinClient{
		sourceConfig: source,
		client:       client,
	}
	return spinClient, nil
}

//returns the last 25 spinnaker pipeline executions
func (c *SpinClient) GetPipelineExecutions() ([]PipelineExecution, error) {
	var pipelineExecutions []PipelineExecution

	//TODO What does expand do ??
	url := fmt.Sprintf("%s/applications/%s/pipelines?limit=25", c.sourceConfig.SpinnakerAPI, c.sourceConfig.SpinnakerApplication)

	if response, err := c.client.Get(url); err != nil {
		return nil, err
	} else if response.StatusCode >= 400 {
		body, err := ioutil.ReadAll(response.Body)
		if err == nil {
			err = fmt.Errorf("spinnaker api responded with status code: %d, body: %s", response.StatusCode, string(body))
		}
		return nil, err
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(body), &pipelineExecutions)
		if err != nil {
			return nil, err
		}
		return pipelineExecutions, nil
	}
}

func (c *SpinClient) InvokePipelineExecution(body []byte) (PipelineExecution, error) {

	pipelineExecution := PipelineExecution{}

	url := fmt.Sprintf("%s/pipelines/%s/%s", c.sourceConfig.SpinnakerAPI, c.sourceConfig.SpinnakerApplication, c.sourceConfig.SpinnakerPipeline)

	if response, err := c.client.Post(url, "application/json", bytes.NewBuffer(body)); err != nil {
		return pipelineExecution, err
	} else if response.StatusCode >= 400 {
		body, err := ioutil.ReadAll(response.Body)
		if err == nil {
			err = fmt.Errorf("spinnaker api responded with status code: %d, body: %s", response.StatusCode, string(body))
		}
		return pipelineExecution, err
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return pipelineExecution, err
		}
		var Data map[string]interface{}
		err = json.Unmarshal([]byte(body), &Data)
		if err != nil {
			return pipelineExecution, err
		}

		pipelineExecution.ID = strings.Split(Data["ref"].(string), "/")[2]
		return pipelineExecution, nil
	}
}
