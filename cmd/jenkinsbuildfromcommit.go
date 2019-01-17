package cmd

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/boyenn/myrelease/lib/jenkins"
	"strings"
)

func GetJenkinsBuildFromCommit(jobName string, lookingForRevision string, jenkinsUser string, jenkinsPass string) (*jenkins.Build, error) {
	if len(lookingForRevision) > 8 {
		lookingForRevision = lookingForRevision[:8]
	}

	j := jenkins.CreateJenkins("https://jenkins2.qone.cloud-apps.tvh.com", jenkinsUser, jenkinsPass)

	job := j.GetJob(jobName)

	var buildContainingRevision *jenkins.Build
	e := retry.Do(
		func() error {
			buildsResponse, e := job.GetAllBuilds()
			if e != nil {
				panic(e)
			}
			for _, build := range buildsResponse.Builds {
				if len(build.ChangeSets) > 0 {
					Items := build.ChangeSets[0].Items
					commitHash := Items[len(Items)-1].CommitId
					if strings.HasPrefix(commitHash, lookingForRevision) {
						correctBuild := build
						buildContainingRevision = &correctBuild
						break
					}
				}
			}

			if buildContainingRevision != nil {
				return nil
			} else {
				return fmt.Errorf("No build in %s with revision %s", jobName, lookingForRevision)
			}
		},
	)
	return buildContainingRevision, e
}
