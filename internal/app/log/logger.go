package log

//go:generate go run genlog.go Print Fatal Info Error Warn

import (
	"fmt"
	defaultLog "log"
	"os"

	"github.com/sirupsen/logrus"
)

// logFormatter - кастомный формат для логгера logrus.
type logFormatter struct {
}

// Format реализует кастомный вывод сообщения.
func (f *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(
		fmt.Sprintf(
			"%v [%v] %v\n",
			entry.Time.Format("2006/01/02 03:04:05"),
			entry.Level,
			entry.Message),
	), nil
}

// init настраивает поток и формат вывода для логгера.
func init() {
	logrus.SetOutput(os.Stdout)
	defaultLog.SetOutput(os.Stdout)
	logrus.SetFormatter(new(logFormatter))
	logrus.SetLevel(logrus.DebugLevel)
}
