package jenkins

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type BasicAuth struct {
	Username string
	Password string
}

type Jenkins struct {
	Server    string
	Version   string
	BasicAuth *BasicAuth
}

type BuildsResponse struct {
	Builds []Build `json:"allBuilds"`
}
type Build struct {
	Number     int    `json:"number"`
	Url        string `json:"url"`
	ChangeSets []struct {
		Items []struct {
			CommitId string `json:"commitId"`
		} `json:"items"`
	} `json:"changeSets"`
	Kind string
}

func CreateJenkins(base string, auth ...interface{}) *Jenkins {
	j := &Jenkins{}
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	j.Server = base

	if len(auth) == 2 {
		j.BasicAuth = &BasicAuth{Username: auth[0].(string), Password: auth[1].(string)}
	}
	return j
}

func (s *Jenkins) doRequest(req *http.Request) ([]byte, error) {
	req.SetBasicAuth(s.BasicAuth.Username, s.BasicAuth.Password)
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
	if 200 != resp.StatusCode {
		return nil, fmt.Errorf("%s", body)
	}
	return body, nil
}

func (s *Jenkins) GetJob(name string) *Job {
	return &Job{
		Jenkins: s,
		name:    name,
		Url:     s.Server + "/job/" + name,
	}
}

type Job struct {
	Jenkins *Jenkins
	Url     string
	name    string
}

func (s *Job) GetAllBuilds() (*BuildsResponse, error) {
	url := s.Jenkins.Server + "/job/" + s.name + "/api/json?tree=allBuilds[changeSets[*,items[commitId]],number,url]"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.Jenkins.doRequest(req)
	if err != nil {
		return nil, err
	}

	var data BuildsResponse
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *Job) GetJenkinsBuildManager(number string) *JenkinsBuildManager {
	return &JenkinsBuildManager{
		Job:    s,
		Number: number,
		Url:    s.Url + "/" + number,
	}
}

type JenkinsBuildManager struct {
	Number string
	Job    *Job
	Url    string
}

func (s *JenkinsBuildManager) GetJobOutput() (string, error) {
	url := s.Job.Jenkins.Server + "/job/" + s.Job.name + "/" + s.Number + "/consoleText"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	bytes, err := s.Job.Jenkins.doRequest(req)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *JenkinsBuildManager) GetPushedDockerImage() (string, error) {
	dockerPushCommand := "+ docker push eu.gcr.io/qone-144607/" + s.Job.name + ":"
	jobOutput, e := s.GetJobOutput()
	if e != nil {
		return "", e
	}
	scanner := bufio.NewScanner(strings.NewReader(jobOutput))
	for scanner.Scan() {
		scanned := scanner.Text()
		if strings.HasPrefix(scanned, dockerPushCommand) {
			tag := "qone-144607/" + s.Job.name + ":" + strings.Split(scanned, dockerPushCommand)[1]
			return tag, nil
		}
	}
	return "", e
}
