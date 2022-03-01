package log

import (
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
)

func Info(msg ...interface{}) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	log.WithFields(log.Fields{
		"line": numLine,
		"file": srcFile,
	}).Info(msg)
}

func Warning(msg ...interface{}) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	log.WithFields(log.Fields{
		"line": numLine,
		"file": srcFile,
	}).Warning(msg)
}

func Error(err ...interface{}) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	log.WithFields(log.Fields{
		"line": numLine,
		"file": srcFile,
	}).Error(err)
}

func Debug(value ...interface{}) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	log.WithFields(log.Fields{
		"line": numLine,
		"file": srcFile,
	}).Debug(value)
}

func Fatal(value ...interface{}) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	log.WithFields(log.Fields{
		"line": numLine,
		"file": srcFile,
	}).Fatal(value)
}

func Println(value ...interface{}) {
	_, path, numLine, _ := runtime.Caller(1)
	srcFile := filepath.Base(path)
	log.WithFields(log.Fields{
		"line": numLine,
		"file": srcFile,
	}).Println(value)
}
