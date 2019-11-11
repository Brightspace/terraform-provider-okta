package api

import (
	"fmt"
)

func NewHTTPSignature(key string) (map[string]interface{}, error) {
	headers := make(map[string]interface{})

	contentType := "application/json"

	headers["Accept"] = contentType
	headers["Content-Type"] = contentType
	headers["Authorization"] = fmt.Sprintf("SSWS %s", key)

	return headers, nil
}
