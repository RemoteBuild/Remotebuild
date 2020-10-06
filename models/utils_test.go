package models

import "testing"

func TestGetCopyDest1(t *testing.T) {
	src := "/tmp/remotebuild_CHewqdNt0DqIrkRgTot9cmIF67oqTS/glogg/glogg-1.1.4.tar.gz"
	dst := "/home/jojii/programming/go/src/RemoteBuild/output"

	if out := getCopyDestFile(src, dst); out != "/home/jojii/programming/go/src/RemoteBuild/output/glogg-1.1.4.tar.gz" {
		t.Errorf("Expected another output. Got: %s", out)
	}
}

func TestGetCopyDest2(t *testing.T) {
	src := "glogg-1.1.4.tar.gz"
	dst := "/home/jojii/programming/go/src/RemoteBuild/output"

	if out := getCopyDestFile(src, dst); out != "/home/jojii/programming/go/src/RemoteBuild/output/glogg-1.1.4.tar.gz" {
		t.Errorf("Expected another output. Got: %s", out)
	}
}
