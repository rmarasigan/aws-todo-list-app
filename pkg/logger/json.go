package logger

import (
	"encoding/json"
	"errors"
)

func returnJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		errors.New("failed to marshal data")
		return ""
	}

	return string(data)
}
