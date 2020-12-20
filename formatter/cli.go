package formatter

import (
	"bytes"

	"github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/gologger/levels"
)

// CLI is a formatter for outputting CLI logs
type CLI struct {
	NoUseColors bool
}

var _ Formatter = &CLI{}

// Format formats the log event data into bytes
func (c *CLI) Format(event *LogEvent) ([]byte, error) {
	c.colorizeLable(event)

	buffer := &bytes.Buffer{}
	buffer.Grow(len(event.Message))

	label, ok := event.Metadata["label"]
	if label != "" && ok {
		buffer.WriteRune('[')
		buffer.WriteString(label)
		buffer.WriteRune(']')
		buffer.WriteRune(' ')
	}
	buffer.WriteString(event.Message)

	for k, v := range event.Metadata {
		buffer.WriteRune(' ')
		buffer.WriteString(c.colorizeKey(k))
		buffer.WriteRune('=')
		buffer.WriteString(v)
	}
	buffer.WriteRune('\n')
	data := buffer.Bytes()
	return data, nil
}

// colorizeKey colorizes the metadata key if enabled
func (c *CLI) colorizeKey(key string) string {
	if c.NoUseColors {
		return key
	}
	return aurora.Bold(key).String()
}

// colorizeLable colorizes the label if their exists one and colors are enabled
func (c *CLI) colorizeLable(event *LogEvent) {
	lable := event.Metadata["lable"]
	if lable == "" || c.NoUseColors {
		return
	}
	switch event.Level {
	case levels.LevelSilent:
		return
	case levels.LevelInfo, levels.LevelVerbose:
		event.Metadata["lable"] = aurora.Blue(lable).String()
	case levels.LevelFatal:
		event.Metadata["lable"] = aurora.Bold(aurora.Red(lable)).String()
	case levels.LevelError:
		event.Metadata["lable"] = aurora.Red(lable).String()
	case levels.LevelDebug:
		event.Metadata["lable"] = aurora.Magenta(lable).String()
	case levels.LevelWarning:
		event.Metadata["lable"] = aurora.Yellow(lable).String()
	}
}
