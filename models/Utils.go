package models

import (
	"encoding/base64"
)

func base64Decode(inp string) (string, error) {
	dec, err := base64.RawStdEncoding.DecodeString(inp)
	if err != nil {
		return "", err
	}
	return string(dec), nil
}
