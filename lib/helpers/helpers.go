package helpers

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"os"
	"strings"
)

func GetFullDirName() (fullDirName string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Could not get current working directory")
	}
	return dir
}

func GetLastDir(fullDirname string) (dirName string) {
	split := strings.Split(fullDirname, "/")
	return split[len(split)-1]
}

func GetCommitHash(fullDirName string) (commitHash string) {
	r, err := git.PlainOpen(fullDirName)
	if err != nil {
		panic(err)
	}
	reference, e := r.Head()
	if e != nil {
		panic(e)
	}
	return reference.Hash().String()
}

func GetBranchName(fullDirName string) (commitHash string) {
	r, err := git.PlainOpen(fullDirName)
	if err != nil {
		panic(err)
	}
	reference, e := r.Head()
	if e != nil {
		panic(e)
	}
	return reference.Name().String()
}

func GetEnv(envName string) string {
	val := os.Getenv(envName)
	if val == "" {
		panic(fmt.Errorf("Environment variable %s not set", envName))
	}
	return val
}
