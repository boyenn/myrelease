package cmd

import (
	"fmt"
	"github.com/boyenn/myrelease/lib/helpers"
	"github.com/boyenn/myrelease/lib/jira"
	"io/ioutil"
	"os/user"
	"strings"
)

func CreateCommitTemplate(jiraUser string, jiraPass string) {
	jiraApi := jira.CreateJira("https://jira.qone.mateco.eu/rest/api/latest/", jiraUser, jiraPass);
	currentDir := helpers.GetFullDirName()
	fullBranchName := helpers.GetBranchName(currentDir)
	split := strings.Split(fullBranchName, "/")
	issueKey := split[len(split)-1]
	issue, e := jiraApi.GetIssue(issueKey)
	if e != nil {
		panic(e)
	}

	commitizenTag := createCommitizenTagFromIssueType(issue.Fields.IssueType.Name)
	template := createTemplate(commitizenTag, issueKey, issue.Fields.Summary)
	fmt.Println(template)

	d1 := []byte(template)

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Println(usr.HomeDir)

	ioErr := ioutil.WriteFile(usr.HomeDir+"/.gitmessage", d1, 0644)
	if ioErr != nil {
		panic(ioErr)
	}
}

func createCommitizenTagFromIssueType(issueType string) string {
	switch issueType {
	case "Bug":
		return "fix"
	default:
		return "feat"
	}
}

func createTemplate(commitizenType string, issueKey string, description string) string {
	return fmt.Sprintf(`%s(%s): %s`, commitizenType, issueKey, description)
}
