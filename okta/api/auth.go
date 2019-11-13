package api

import (
	"fmt"
)

func NewHTTPSignature(key string) (map[string]string, error) {
	headers := make(map[string]string)

	contentType := "application/json"

	headers["Accept"] = contentType
	headers["Content-Type"] = contentType
	headers["Authorization"] = fmt.Sprintf("SSWS %s", key)

	return headers, nil
}
