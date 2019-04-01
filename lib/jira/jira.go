package jira

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type BasicAuth struct {
	Username string
	Password string
}

type Jira struct {
	Server    string
	Version   string
	BasicAuth *BasicAuth
}

func CreateJira(base string, auth ...interface{}) *Jira {
	j := &Jira{}
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	j.Server = base

	if len(auth) == 2 {
		j.BasicAuth = &BasicAuth{Username: auth[0].(string), Password: auth[1].(string)}
	}
	return j
}

func (s *Jira) doRequest(req *http.Request) ([]byte, error) {
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

func (s *Jira) GetIssue(key string) (*Issue, error) {
	url := s.Server + "/issue/" + key
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}

	var data Issue
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *Jira) GetIssues(jql string) (*IssuePage, error) {
	Url, err := url.Parse(s.Server + "/search")
	parameters := url.Values{}
	parameters.Add("jql", jql)
	parameters.Add("fields", "summary,key,status")
	Url.RawQuery = parameters.Encode()
	req, err := http.NewRequest("GET", Url.String(), nil)
	if err != nil {
		return nil, err
	}
	bytes, err := s.doRequest(req)
	if err != nil {
		return nil, err
	}

	var data IssuePage
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

type Issue struct {
	Key    string `json:"key"`
	Self   string `json:self`
	Fields struct {
		IssueType struct {
			Name string `json:"name"`
		} `json:"issueType"`
		Summary string `json:"summary"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
	} `json:"fields"`
}

type IssuePage struct {
	Issues []Issue `json:"issues"`
}
