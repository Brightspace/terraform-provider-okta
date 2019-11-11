package api

import (
	"encoding/base64"
	"fmt"
)

func NewHTTPSignature(key []byte) (map[string]interface{}, error) {
	headers := make(map[string]interface{})

	contentType := "application/json"
	encodedAuth := base64.StdEncoding.EncodeToString(key)

	headers["Accept"] = contentType
	headers["Content-Type"] = contentType
	headers["Authorization"] = fmt.Sprintf("SSWS %s", encodedAuth)

	return headers, nil
}
