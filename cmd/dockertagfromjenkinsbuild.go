package cmd

import (
	"fmt"
	"github.com/boyenn/myrelease/lib/jenkins"
	"os"
	"time"
)

func GetDockerTagFromJenkinsBuild(jobName string, buildNumber string, jenkinsUser string, jenkinsPass string) string {

	j := jenkins.CreateJenkins("https://jenkins2.qone.cloud-apps.tvh.com", jenkinsUser, jenkinsPass)
	job := j.GetJob(jobName)
	buildManager := job.GetJenkinsBuildManager(buildNumber)

	dockerImage, e := buildManager.GetPushedDockerImage()
	duration := time.Second * 3
	totalDuration := time.Duration(0)
	maxDuration := time.Minute * 10

	for dockerImage == "" {

		totalDuration += duration
		if totalDuration > maxDuration {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("Waited for %.1f minutes but could not get pushed docker image from jenkins build %s ", maxDuration.Minutes(), buildManager.Url))
			os.Exit(1);
		}

		if e != nil {
			panic(e)
		}

		dockerImage, e = buildManager.GetPushedDockerImage()
		time.Sleep(duration)
	}

	return dockerImage
}
