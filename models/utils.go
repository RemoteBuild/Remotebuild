package models

import (
	"net/http"
	"strings"
)

func setHeadersFromStr(headers string, header *http.Header) {
	headersrn := strings.Split(headers, "\r\n")
	for _, v := range headersrn {
		if !strings.Contains(v, "=") {
			continue
		}
		kp := strings.Split(v, "=")
		key := kp[0]

		(*header).Set(key, kp[1])
	}
}
