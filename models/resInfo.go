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
	JobID   uint
	Name    string
	Version string
	Files   []string
}

// GetResInfoPath return path for resinfo file
func GetResInfoPath(base string) string {
	return filepath.Join(base, ResInfoFileName)
}

// ParseResInfo parses result info from file
func ParseResInfo(dataDir, file string, jobID uint) (*ResInfo, error) {
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
	outFiles := sContent[2:]

	if len(name) == 0 || len(version) == 0 || len(outFiles) == 0 {
		return nil, ErrInvalidFormat
	}

	var newOutfiles []string
	for i := range outFiles {
		if len(strings.TrimSpace(outFiles[i])) == 0 {
			continue
		}

		newOutfiles = append(newOutfiles, filepath.Join(dataDir, "pkgdest", outFiles[i]))
	}

	return &ResInfo{
		Name:    name,
		Version: version,
		Files:   newOutfiles,
		JobID:   jobID,
	}, nil
}
