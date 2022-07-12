package trail

import (
	"fmt"
	"time"

	"github.com/rmarasigan/aws-todo-list-app/pkg/response"
)

type Trail struct {
	Level     string `json:"trail_level"`
	Message   string `json:"trail_message"`
	TimeStamp string `json:"timestamp"`
}

const (
	OK    = 0
	INFO  = 1
	DEBUG = 2
	ERROR = 3
)

var trailLevel = map[int]string{
	OK:    "OK",
	INFO:  "INFO",
	DEBUG: "DEBUG",
	ERROR: "ERROR",
}

// SetTimeStamp sets the current timestamp with the ff. format: 2006-01-02 15:04:05.
func (t *Trail) SetTimeStamp() {
	t.TimeStamp = time.Now().Format("2006-01-02 15:04:05")
}

// Print accepts a level parameter and formats according to a format.
//
// level accepts OK, INFOR, DEBUG and ERROR.
func Print(level int, msg interface{}, i ...interface{}) {
	message := fmt.Sprint(msg)

	entry := new(Trail)
	entry.Level = trailLevel[level]
	entry.Message = fmt.Sprintf(message, i...)
	entry.SetTimeStamp()

	fmt.Println(response.EncodeJSON(entry))
}
