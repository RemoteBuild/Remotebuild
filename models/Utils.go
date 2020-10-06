package models

import (
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
)

func base64Decode(inp string) (string, error) {
	dec, err := base64.RawStdEncoding.DecodeString(inp)
	if err != nil {
		return "", err
	}
	return string(dec), nil
}

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	dst = getCopyDestFile(src, dst)
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	buff := make([]byte, 1024*1024)

	_, err = io.CopyBuffer(out, in, buff)
	if err != nil {
		return err
	}

	return out.Close()
}

func getCopyDestFile(src, dst string) string {
	s, err := os.Stat(dst)
	if err == nil && s != nil && s.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}

	return dst
}
