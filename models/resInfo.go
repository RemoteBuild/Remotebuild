package models

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// ResInfoFileName name for resInfo file
const ResInfoFileName = "resInfo"

// ErrFileEmpty if file is empty
var ErrFileEmpty = errors.New("file is empty")

// ErrInvalidFormat error if resinfo format is invalid
var ErrInvalidFormat = errors.New("Invalid resInfo format")

// ResInfo containing Information about
// the result of an built package
type ResInfo struct {
	Name    string
	Version string
	File    string
}

// GetResInfoPath return path for resinfo file
func GetResInfoPath(base string) string {
	return filepath.Join(base, ResInfoFileName)
}

// ParseResInfo parses result info from file
func ParseResInfo(file string) (*ResInfo, error) {
	s, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	if s.Size() == 0 {
		return nil, ErrFileEmpty
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Split file to into each line
	sContent := strings.Split(string(content), "\n")

	if len(sContent) < 3 {
		return nil, ErrInvalidFormat
	}

	name := sContent[0]
	version := sContent[1]
	outFile := sContent[2]

	if len(name) == 0 || len(version) == 0 || len(outFile) == 0 {
		return nil, ErrInvalidFormat
	}

	return &ResInfo{
		Name:    name,
		Version: version,
		File:    outFile,
	}, nil
}
