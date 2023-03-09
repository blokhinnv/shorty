// Code generated by go generate; DO NOT EDIT.
// This file was generated by genlog.go

package log

import "github.com/sirupsen/logrus"

func SetLevel(level logrus.Level) {
	logrus.SetLevel(level)
}

const (
	FatalLevel logrus.Level = logrus.FatalLevel

	InfoLevel logrus.Level = logrus.InfoLevel

	ErrorLevel logrus.Level = logrus.ErrorLevel

	WarnLevel logrus.Level = logrus.WarnLevel
)

func Print(args ...interface{}) {
	logrus.Print(args...)
}

func Printf(format string, args ...interface{}) {
	logrus.Printf(format, args...)
}

func Println(args ...interface{}) {
	logrus.Println(args...)
}

func Fatal(args ...interface{}) {
	logrus.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	logrus.Fatalf(format, args...)
}

func Fatalln(args ...interface{}) {
	logrus.Fatalln(args...)
}

func Info(args ...interface{}) {
	logrus.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Infoln(args ...interface{}) {
	logrus.Infoln(args...)
}

func Error(args ...interface{}) {
	logrus.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Errorln(args ...interface{}) {
	logrus.Errorln(args...)
}

func Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Warnln(args ...interface{}) {
	logrus.Warnln(args...)
}
