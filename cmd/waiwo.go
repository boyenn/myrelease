package cmd

import (
	"fmt"
	"github.com/boyenn/myrelease/lib/jira"
)

func ListWhatAmIWorkingOn(jiraUser string, jiraPass string) {
	jiraApi := jira.CreateJira("https://jira.qone.mateco.eu/rest/api/latest/", jiraUser, jiraPass);
	issuePage, e := jiraApi.GetIssues(`project=RMS AND assignee="` + jiraUser + `" AND status not in (Done,Resolved)`)
	if e != nil {
		panic(e)
	}
	for _, element := range issuePage.Issues {
		fmt.Printf("%s : %s (%s) https://jira.qone.mateco.eu/browse/%s\n", element.Key, element.Fields.Summary, element.Fields.Status.Name, element.Key)
	}
}
