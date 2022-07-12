package response

import (
	"encoding/json"

	"github.com/rmarasigan/aws-todo-list-app/pkg/logger"
)

const (
	StatusOK                  = 200
	StatusFound               = 302
	StatusSeeOther            = 303
	StatusTemporaryRedirect   = 307
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusRequestTimeout      = 408
	StatusInternalServerError = 500
	StatusBadGateway          = 502
	StatusServiceUnavailable  = 503
)

var StatusMap = map[int]string{
	StatusOK:                  "OK",
	StatusFound:               "Found",
	StatusSeeOther:            "See Other",
	StatusTemporaryRedirect:   "Temporary Redirect",
	StatusBadRequest:          "Bad Request",
	StatusUnauthorized:        "Unauthorized",
	StatusForbidden:           "Forbidden",
	StatusNotFound:            "Not Found",
	StatusRequestTimeout:      "Request Timeout",
	StatusInternalServerError: "Internal Server Error",
	StatusBadGateway:          "Bad Gateway",
	StatusServiceUnavailable:  "Service Unavailable",
}

// EncodeResponseJSON marshal response JSON to produce a string format JSON.
func EncodeResponseJSON(status int, body interface{}) string {
	encodeJSON, err := json.Marshal(body)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "encodeJSONError", Message: "Failed to encode JSON"},
			logger.KVP{Key: "StatusCode", Value: StatusMap[status]})
	}

	return string(encodeJSON)
}

// EncodeJSON marshal response JSON to produce a string format JSON.
func EncodeJSON(body interface{}) string {
	encodeJSON, err := json.Marshal(body)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "EncodeJSONError", Message: "Failed to encode JSON"})
	}

	return string(encodeJSON)
}

// ParseJSON unmarshal the JSON-encoded data and stores the result in the value pointed to by v.
func ParseJSON(data []byte, v interface{}) error {
	err := json.Unmarshal([]byte(data), v)
	if err != nil {
		logger.Error(err, &logger.Logs{Code: "parseJSONError", Message: "Failed to parse JSON"})
		return err
	}

	return nil
}
