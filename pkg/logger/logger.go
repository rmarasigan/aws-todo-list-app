package logger

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	LOG_INFO  = "INFO"
	LOG_DEBUG = "DEBUG"
	LOG_ERROR = "ERROR"
)

type Logs struct {
	Code         string                 `json:"log_code"`
	Message      interface{}            `json:"log_msg"`
	ErrorMessage string                 `json:"log_errmsg,omitempty"`
	Level        string                 `json:"log_level"`
	Keys         map[string]interface{} `json:"log_keys,omitempty"`
	TimeStamp    string                 `json:"log_timestamp"`
}

// Print marshal response JSON to print a string format JSON.
func (log *Logs) Print() {
	encodeJSON, err := json.Marshal(log)
	if err != nil {
		Error(err, &Logs{Code: "EncodeJSONError", Message: "Failed to encode JSON"})
	}

	fmt.Println(string(encodeJSON))
}

// SetKeys checks if Log Keys are empty in order to create an empty map. If it's not empty, set its key-value pair.
func (l *Logs) SetKeys(key string, value interface{}) {
	if l.Keys == nil {
		// Create an empty map
		l.Keys = make(map[string]interface{})
	}
	// Set key-value pairs using typical name[key] = val syntax
	l.Keys[key] = value
}

// SetTimeStamp sets the current timestamp with the ff. format: 2006-01-02 15:04:05.
func (l *Logs) SetTimeStamp() {
	l.TimeStamp = time.Now().Format("2006-01-02 15:04:05")
}

// Info prints a log information
func Info(logs *Logs, kv ...KVP) {
	var entry Logs

	entry.Code = logs.Code
	entry.Level = LOG_INFO
	entry.Message = logs.Message
	entry.SetTimeStamp()

	if len(kv) != 0 {
		for _, kvp := range kv {
			entry.SetKeys(kvp.KeyValue())
		}
	}

	entry.Print()
}

// Debug prints a debug log information
func Debug(logs *Logs, kv ...KVP) {
	var entry Logs

	entry.Code = logs.Code
	entry.Level = LOG_DEBUG
	entry.Message = logs.Message
	entry.SetTimeStamp()

	if len(kv) != 0 {
		for _, kvp := range kv {
			entry.SetKeys(kvp.KeyValue())
		}
	}

	entry.Print()
}

// Error prints an error log information
func Error(err error, logs *Logs, kv ...KVP) {
	var entry Logs

	entry.Level = LOG_ERROR
	entry.Code = logs.Code
	entry.Message = logs.Message
	entry.SetTimeStamp()

	if err != nil {
		entry.ErrorMessage = err.Error()
	}

	if len(kv) != 0 {
		for _, kvp := range kv {
			entry.SetKeys(kvp.KeyValue())
		}
	}

	entry.Print()
}
