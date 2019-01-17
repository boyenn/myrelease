package myrelease

import (
	"github.com/boyenn/myrelease/cmd"
	"github.com/boyenn/myrelease/lib/helpers"
	"github.com/urfave/cli"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	app := cli.NewApp()

	app.Name = "MyRelease"
	app.Usage = "A compilation of commands that make it easier to do releases for the MySite project"
	app.Version = "0.0.1"
	app.EnableBashCompletion = true

	var jenkinsUser string
	var jenkinsPass string
	var jiraPass string
	var jiraUser string
	var spinnakerSessionCookie string

	app.Commands = []cli.Command{
		{
			Name:  "jenkins",
			Usage: "Commands involving jenkins",

			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "jenkins-user",
					Usage:       "Jenkins username for authentication",
					EnvVar:      "JENKINS_USER",
					Destination: &jenkinsUser,
				},

				cli.StringFlag{
					Name:        "jenkins-pass",
					Usage:       "Jenkins password for authentication",
					EnvVar:      "JENKINS_PASS",
					Destination: &jenkinsPass,
				},
			},
			Subcommands: []cli.Command{
				{
					Name:      "dockertag",
					Usage:     "Gets the dockertag from a jenkins build",
					ArgsUsage: "<job name> <build number>",
					Action: func(c *cli.Context) error {
						jobName := c.Args().Get(0)
						buildNumber := c.Args().Get(1)

						dockerTag := cmd.GetDockerTagFromJenkinsBuild(jobName, buildNumber, jenkinsUser, jenkinsPass)
						os.Stdout.WriteString(dockerTag + "\n")
						return nil
					},
				},
				{
					Name:        "buildnumber",
					Usage:       "Gets the jenkins build number from a job name and commit.",
					Description: "Will automatically use current dir and HEAD as reference point if no arguments are provided",
					ArgsUsage:   "<job name> <commit hash>",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "short, s",
							Usage: "Shortens output by printing only the build number",
						},
					},
					Action: func(c *cli.Context) error {
						var jobName string
						var lookingForRevision string
						if len(c.Args()) == 2 {
							jobName = c.Args().Get(0)
							lookingForRevision = c.Args().Get(1)
						} else {
							currentDir := helpers.GetFullDirName()
							lookingForRevision = helpers.GetCommitHash(currentDir)
							jobName = helpers.GetLastDir(currentDir)
						}
						build, e := cmd.GetJenkinsBuildFromCommit(jobName, lookingForRevision, jenkinsUser, jenkinsPass)
						if e != nil {
							return e
						}

						if c.Bool("short") {
							_, _ = os.Stdout.WriteString(strconv.Itoa(build.Number) + "\n")
						} else {
							_, _ = os.Stdout.WriteString(build.Url + "\n")
						}
						return nil
					},
				},
			},
		},
		{
			Name:  "spinnaker",
			Usage: "Interact with spinnaker applications",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "spinnaker-session-cookie",
					Usage:       "Manually stolen spinnaker session cookie for authentication",
					EnvVar:      "SPINNAKER_SESSION_COOKIE",
					Destination: &spinnakerSessionCookie,
				},
			},
			Subcommands: []cli.Command{
				{
					Name:      "print",
					Usage:     "Print information about a spinnaker appication in a given region",
					ArgsUsage: "<application> <region>",
					Action: func(c *cli.Context) error {
						applicationName := c.Args().Get(0)
						region := c.Args().Get(1)

						cmd.PrintSpinnakerInfo(applicationName, region, spinnakerSessionCookie)
						return nil
					},
				},
				{
					Name:      "deploy",
					Usage:     "Deploy tag in a given region",
					ArgsUsage: "<application> <region> <full_tag>",
					Action: func(c *cli.Context) error {
						applicationName := c.Args().Get(0)
						region := c.Args().Get(1)

						fullTagName := c.Args().Get(2)

						repository := strings.Split(fullTagName, ":")[0]
						tag := strings.Split(fullTagName, ":")[1]
						cmd.Deploy(applicationName, region, repository, tag, spinnakerSessionCookie)
						return nil
					},
				},
			},
		},
		{
			Name:  "general",
			Usage: "general helper commands",
			Flags: []cli.Flag{

				cli.StringFlag{
					Name:        "jira-user",
					Usage:       "jira username for authentication",
					EnvVar:      "JIRA_USER",
					Destination: &jiraUser,
				},

				cli.StringFlag{
					Name:        "jira-pass",
					Usage:       "Jira password for authentication",
					EnvVar:      "JIRA_PASS",
					Destination: &jiraPass,
				},
			},
			Subcommands: []cli.Command{
				{
					Name:  "commit",
					Usage: "Creates commit template for current directory",
					Action: func(c *cli.Context) error {
						cmd.CreateCommitTemplate(jiraUser, jiraPass)
						return nil
					},
				},
				{
					Name:  "waiwo",
					Usage: "What am I working on?",
					Action: func(c *cli.Context) error {
						cmd.ListWhatAmIWorkingOn(jiraUser, jiraPass)
						return nil
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
