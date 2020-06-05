package models

import (
	"errors"
	"fmt"

	libremotebuild "github.com/JojiiOfficial/LibRemotebuild"
)

var (
	// ErrAURNoRepoFound if no repo name was given
	ErrAURNoRepoFound = errors.New("No AUR repo-name found")
)

// DataManagerArgs data for datamanager
type DataManagerArgs struct {
	Username  string
	Host      string
	Token     string
	Namespace string
}

// ArgParser parse job args
type ArgParser struct {
	args    map[string]string
	JobType libremotebuild.JobType
}

// NewArgParser create new argparser
func NewArgParser(args map[string]string, JobType libremotebuild.JobType) *ArgParser {
	return &ArgParser{
		args:    args,
		JobType: JobType,
	}
}

// ParseEnvars parse args to envars based on the JobType
func (argParser *ArgParser) ParseEnvars() ([]string, error) {
	switch argParser.JobType {
	case libremotebuild.JobAUR:
		return argParser.parseAURArgs()
	}

	return argsToEnvs(argParser.args), nil
}

// Get AUR repo name from args
func (argParser *ArgParser) getAURRepoName() string {
	return argParser.args[libremotebuild.AURPackage]
}

// Parse Args for AUR package builds
func (argParser *ArgParser) parseAURArgs() ([]string, error) {
	repoName := argParser.getAURRepoName()
	if len(repoName) == 0 {
		return nil, ErrAURNoRepoFound
	}

	return []string{fmt.Sprintf("%s=%s", libremotebuild.AURPackage, repoName)}, nil
}

// HasDataManagerArgs return true if DManager data is available
func (argParser *ArgParser) HasDataManagerArgs() bool {
	_, userNameOK := argParser.args[libremotebuild.DMUser]
	_, tokenOK := argParser.args[libremotebuild.DMUser]
	_, hostOK := argParser.args[libremotebuild.DMHost]
	return userNameOK && tokenOK && hostOK
}

// GetDManagerData return Dmanager Data
func (argParser *ArgParser) GetDManagerData() *DataManagerArgs {
	if !argParser.HasDataManagerArgs() {
		return nil
	}

	host := argParser.args[libremotebuild.DMHost]
	username := argParser.args[libremotebuild.DMUser]
	token := argParser.args[libremotebuild.DMToken]
	namespace := argParser.args[libremotebuild.DMNamespace]

	return &DataManagerArgs{
		Host:      host,
		Username:  username,
		Token:     token,
		Namespace: namespace,
	}
}

// Transform all keys in hashmap to envars
func argsToEnvs(args map[string]string) []string {
	var s []string
	for k, v := range args {
		s = append(s, k+"="+v)
	}
	return s
}

// GetDManagerNamespace return the namespace according to the dmanager args
func (argParser *ArgParser) GetDManagerNamespace() string {
	return argParser.args[libremotebuild.DMNamespace]
}

// HasNamespace return true if namespace is set
func (argParser *ArgParser) HasNamespace() bool {
	return len(argParser.GetDManagerNamespace()) > 0
}
