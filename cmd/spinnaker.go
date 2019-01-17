package cmd

import (
	"github.com/boyenn/myrelease/lib/spinnaker"
	"strings"
)

func PrintSpinnakerInfo(applicationName, regionName, sessionCookie string) {

	spinnakerPkg := spinnaker.CreateSpinnaker("http://spinnaker-api.qone.cloud-apps.tvh.com/", sessionCookie)
	spinnakerApplication := spinnakerPkg.GetApplication(applicationName)

	group, e := spinnakerApplication.GetEnabledServerGroupInRegion(regionName)
	if e != nil {
		panic(e)
	}

	println(strings.Join(group.BuildInfo.Images, " "))

	pipelineConfig, e := spinnakerApplication.GetPipelineConfigForRegion(regionName)
	println(pipelineConfig.Name)
}

func Deploy(applicationName, regionName, repository, tag, sessionCookie string) {
	spinnakerPkg := spinnaker.CreateSpinnaker("http://spinnaker-api.qone.cloud-apps.tvh.com/", sessionCookie)
	spinnakerApplication := spinnakerPkg.GetApplication(applicationName)
	pipelineConfig, e := spinnakerApplication.GetPipelineConfigForRegion(regionName)
	if e != nil {
		panic(e)
	}
	deployRequest := pipelineConfig.CreateDeployRequest(repository, tag)
	err := deployRequest.Execute()
	if err != nil {
		panic(err)
	}

}
