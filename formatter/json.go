package formatter

import (
	"regexp"
	"time"

	jsoniter "github.com/json-iterator/go"
)

// JSON is a formatter for outputting json logs
type JSON struct{}

var _ Formatter = &JSON{}

// filter matches ASCII color code sequences.
// See https://stackoverflow.com/questions/4842424/list-of-ansi-color-escape-sequences
var filter = regexp.MustCompile(`\x1b\[[0-9;]+m`)

// Format formats the log event data into bytes
func (j *JSON) Format(event LogEvent) ([]byte, error) {
	msg, ok := event["msg"]
	if !ok {
		return nil, nil
	}
	event["timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05-0700")
	event["msg"] = filter.ReplaceAllString(msg, "")
	return jsoniter.Marshal(event)
}
