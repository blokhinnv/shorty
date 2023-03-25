package log

//go:generate go run genlog.go Print Fatal Info Error Warn

import (
	"fmt"
	defaultLog "log"
	"os"

	"github.com/sirupsen/logrus"
)

// logFormatter - custom format for the logrus logger.
type logFormatter struct {
}

// Format implements custom message output.
func (f *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(
		fmt.Sprintf(
			"%v [%v] %v\n",
			entry.Time.Format("2006/01/02 03:04:05"),
			entry.Level,
			entry.Message),
	), nil
}

// init sets up the stream and output format for the logger.
func init() {
	logrus.SetOutput(os.Stdout)
	defaultLog.SetOutput(os.Stdout)
	logrus.SetFormatter(new(logFormatter))
	logrus.SetLevel(logrus.DebugLevel)
}
