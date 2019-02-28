package cloudbuild

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudbuild/v1"
	"io/ioutil"
)

func CreateClient(serviceAccountFileLocation string) (*cloudbuild.Service, error)  {
	data, err := ioutil.ReadFile(serviceAccountFileLocation)
	if err != nil {
		return nil, err
	}
	conf, err := google.JWTConfigFromJSON(data, cloudbuild.CloudPlatformScope)
	if err != nil {
		return nil, err
	}
	client := conf.Client(oauth2.NoContext)
	svc, e := cloudbuild.New(client)
	if e != nil {
		return nil, err
	}
	return svc, nil;
}