package loggy

import (
	"os"
)

func Info(args ...any) {
	instance.Info(args...)
}

func Debug(args ...any) {
	instance.Debug(args...)
}

func Error(args ...any) {
	instance.Error(args...)
}

func Warn(args ...any) {
	instance.Warn(args...)
}

func Infof(template string, args ...any) {
	instance.Infof(template, args...)
}

func Debugf(template string, args ...any) {
	instance.Debugf(template, args...)
}

func Errorf(template string, args ...any) {
	instance.Errorf(template, args...)
}

func Warnf(template string, args ...any) {
	instance.Warnf(template, args...)
}

func Infoln(args ...any) {
	instance.Infoln(args...)
}

func Debugln(args ...any) {
	instance.Debugln(args...)
}

func Errorln(args ...any) {
	instance.Errorln(args...)
}

func Warnln(args ...any) {
	instance.Warnln(args...)
}

func Fatal(args ...any) {
	Errorln(args...)
	os.Exit(1)
}

func Fatalf(template string, args ...any) {
	Errorf(template, args...)
	os.Exit(1)
}
