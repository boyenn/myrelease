package spinnaker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Spinnaker struct {
	Server        string
	SessionCookie string
}

func CreateSpinnaker(base string, sessionCookie string) *Spinnaker {
	j := &Spinnaker{}
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	j.Server = base
	j.SessionCookie = sessionCookie
	return j
}

func (s *Spinnaker) doRequest(req *http.Request) ([]byte, error) {

	cookie := http.Cookie{Name: "SESSION", Value: s.SessionCookie}
	req.AddCookie(&cookie)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(strconv.Itoa(resp.StatusCode), "2") {
		return nil, fmt.Errorf("%s", body)
	}
	return body, nil
}

func (s *Spinnaker) GetApplication(name string) *SpinnakerApplication {
	return &SpinnakerApplication{
		Spinnaker: s,
		name:      name,
	}
}

type SpinnakerApplication struct {
	Spinnaker *Spinnaker
	name      string
}

func (s *SpinnakerApplication) GetDetails() error {
	url := s.Spinnaker.Server + "/applications/" + s.name
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	bytes, err := s.Spinnaker.doRequest(req)
	if err != nil {
		return err
	}
	println(string(bytes), err)
	return nil
}

type ServerGroup struct {
	Account   string `json:"account"`
	BuildInfo struct {
		Images []string `json:"images"`
	} `json:"buildInfo"`
	IsDisabled bool   `json:"isDisabled"`
	Name       string `json:"name"`
	Region     string `json:"region"`
	Instances  []struct {
		HealthState string `json:"healthState"`
	} `json:"instances"`
}

func (s *SpinnakerApplication) GetServerGroups() ([]ServerGroup, error) {
	url := s.Spinnaker.Server + "/applications/" + s.name + "/serverGroups"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	bytes, err := s.Spinnaker.doRequest(req)
	if err != nil {
		return nil, err
	}

	var data []ServerGroup
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *SpinnakerApplication) GetEnabledServerGroupInRegion(region string) (*ServerGroup, error) {
	groups, e := s.GetServerGroups()
	if e != nil {
		return nil, e
	}
	for _, group := range groups {
		if !group.IsDisabled && group.Instances[0].HealthState == "Up" && group.Region == region {
			return &group, nil
		}
	}

	return nil, fmt.Errorf("Server group not found")
}

func (s *SpinnakerApplication) GetPipelineConfigs() ([]PipelineConfig, error) {
	url := s.Spinnaker.Server + "/applications/" + s.name + "/pipelineConfigs"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	bytes, err := s.Spinnaker.doRequest(req)
	if err != nil {
		return nil, err
	}

	var data []PipelineConfig
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	for i := range data {
		data[i].SpinnakerApplication = s
	}

	return data, nil
}

func (s *SpinnakerApplication) GetPipelineConfigForRegion(region string) (*PipelineConfig, error) {
	configs, e := s.GetPipelineConfigs()
	if e != nil {
		return nil, e
	}
	for _, pipelineConfig := range configs {
		deployRegion, err := pipelineConfig.GetFirstDeployRegion()
		if err != nil {
			continue
		}
		if deployRegion == region {
			return &pipelineConfig, nil
		}
	}
	return nil, fmt.Errorf("No pipeline found in the specified region")
}

func (pc *PipelineConfig) GetFirstDeployRegion() (string, error) {
	for _, stage := range pc.Stages {
		if stage.Type == "deploy" {
			return stage.Clusters[0].Region, nil
		}
	}
	return "", fmt.Errorf("Deploy region not found")
}

type PipelineConfig struct {
	Application string `json:"application"`
	Name        string `json:"name"`
	Stages      []struct {
		Type     string `json:"type"`
		Clusters []struct {
			Region string `json:"region"`
		} `json:"clusters"`
	} `json:"stages"`
	SpinnakerApplication *SpinnakerApplication
}

func (p *PipelineConfig) CreateDeployRequest(repository string, tag string) *DeployRequest {
	return &DeployRequest{
		Account:        "my-docker-registry-account",
		Application:    p.SpinnakerApplication.name,
		Description:    "",
		Enabled:        false,
		Organization:   "qone-144607",
		Registry:       "eu.gcr.io",
		Repository:     repository,
		Tag:            tag,
		Type:           "manual",
		User:           "boyen.vaesen@tvh.com",
		PipelineConfig: p,
	}
}

func (d *DeployRequest) Execute() error {
	url := d.PipelineConfig.SpinnakerApplication.Spinnaker.Server + "/pipelines/" +
		d.PipelineConfig.SpinnakerApplication.name + "/" +
		d.PipelineConfig.Name

	marshal, e := json.Marshal(d)
	if e != nil {
		return e
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshal))
	req.Header.Set("Content-Type", "application/json")

	if err != nil {
		return err
	}
	_, err = d.PipelineConfig.SpinnakerApplication.Spinnaker.doRequest(req)
	if err != nil {
		return err
	}
	return nil
}

type DeployRequest struct {
	Account        string   `json:"account"`
	Application    string   `json:"application"`
	Description    string   `json:"description"`
	Enabled        bool     `json:"enabled"`
	Organization   string   `json:"organization"`
	Registry       string   `json:"registry"`
	Repository     string   `json:"repository"`
	Status         []string `json:"status"`
	Tag            string   `json:"tag"`
	Type           string   `json:"type"`
	User           string   `json:"user"`
	PipelineConfig *PipelineConfig
}
