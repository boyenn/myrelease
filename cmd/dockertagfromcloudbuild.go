package cmd

import (
	"fmt"
	"google.golang.org/api/cloudbuild/v1"
	"os"
	"strings"
	"time"
)

func GetDockerImage(lookingForRevision string, svc *cloudbuild.Service) (*cloudbuild.Build, error) {
	response, e := svc.Projects.Builds.List("mateco-mysite").PageSize(1000).Do()

	if e != nil {
		panic(e)
	}

	var buildContainingRevision *cloudbuild.Build

	for _, build := range response.Builds {
		if build.SourceProvenance != nil {
			commitHash := build.SourceProvenance.ResolvedRepoSource.CommitSha
			if strings.HasPrefix(commitHash, lookingForRevision) {
				buildContainingRevision = build
				break
			}
		}
	}

	if buildContainingRevision == nil {
		_, _ = os.Stderr.WriteString(fmt.Sprintf("no build with revision %s", lookingForRevision))
		os.Exit(1);
	}

	duration := time.Second * 3
	totalDuration := time.Duration(0)
	maxDuration := time.Minute * 10

	for buildContainingRevision.Status == "WORKING" {
		buildContainingRevision, e = svc.Projects.Builds.Get("mateco-mysite", buildContainingRevision.Id).Do()
		totalDuration += duration
		if totalDuration > maxDuration {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("Waited for %.1f minutes but status of build did not change from WORKING", maxDuration.Minutes()))
			os.Exit(1);
		}

		if e != nil {
			panic(e)
		}

		time.Sleep(duration)
	}

	switch buildContainingRevision.Status {
	case "FAILURE":
		_, _ = os.Stderr.WriteString(fmt.Sprintf("BUILD HAD STATUS FAILURE"))
		os.Exit(1)
	}
	return buildContainingRevision, nil
}
