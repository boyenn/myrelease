package confluent

import (
	"bytes"
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

type SchemaRegistryClient struct {
	Server    string
	BasicAuth *BasicAuth
}

func CreateClient(base string, auth ...interface{}) *SchemaRegistryClient {
	client := &SchemaRegistryClient{}
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	client.Server = base

	if len(auth) == 2 {
		client.BasicAuth = &BasicAuth{Username: auth[0].(string), Password: auth[1].(string)}
	}
	return client
}

func (s *SchemaRegistryClient) doRequest(req *http.Request) ([]byte, error) {
	req.SetBasicAuth(s.BasicAuth.Username, s.BasicAuth.Password)
	req.Header.Add("content-type", "application/json")
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

func (s *SchemaRegistryClient) DeleteSubject(subject string) (error) {
	url := s.Server + "/subjects" + "/" + subject
	req, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		return err
	}

	_, e := s.doRequest(req)
	if e != nil {
		return e
	}

	return nil
}

type Schema struct {
	Id         int    `db:"id"`
	Definition string `db:"definition"`
	Format     string `db:"format"`
	Subject    string `db:"subject"`
	Version    string `db:"version"`
}

type PostSubjectBody struct {
	Schema string `json:"schema"`
}

type CompatibilitySubjectBody struct {
	Compatibility string `json:"compatibility"`
}

type PostSubjectResponseBody struct {
	Id int `json:"id"`
}

func (s *SchemaRegistryClient) PostSubject(schema Schema) (*PostSubjectResponseBody, error) {
	url := s.Server + "/subjects" + "/" + schema.Subject + "/versions"

	body := new(PostSubjectBody)
	body.Schema = schema.Definition
	marshal, er := json.Marshal(body)
	if er != nil {
		return nil, er
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshal))

	if err != nil {
		return nil, err
	}

	bytes, e := s.doRequest(req)
	if e != nil {
		return nil, e
	}

	var data PostSubjectResponseBody
	err = json.Unmarshal(bytes, &data)

	return &data, err
}

func (s *SchemaRegistryClient) GetSubjectVersions(schema Schema) ([]int, error) {
	url := s.Server + "/subjects" + "/" + schema.Subject + "/versions"

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	bytes, e := s.doRequest(req)
	if e != nil {
		return nil, e
	}

	var data []int
	err = json.Unmarshal(bytes, &data)

	return data, err
}

func (s *SchemaRegistryClient) SetSubjectCompatibility(subject string, compatibility string) error {
	url := s.Server + "/config" + "/" + subject
	body := new(CompatibilitySubjectBody)
	body.Compatibility = compatibility
	marshal, er := json.Marshal(body)

	if er != nil {
		return er
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(marshal))

	if err != nil {
		return err
	}

	_, e := s.doRequest(req)
	if e != nil {
		return e
	}

	return nil
}
