package logger

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

var log = logrus.New()

func Setup() {
	log.SetFormatter(&logrus.JSONFormatter{})
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func WithRequest(id string, req *http.Request) *logrus.Entry {
	fs := logrus.Fields{
		"requestId":  id,
		"uri":        req.RequestURI,
		"method":     req.Method,
		"remoteaddr": req.RemoteAddr,
	}
	return log.WithFields(fs)
}

func WithFields(fields map[string]interface{}) *logrus.Entry {
	return log.WithFields(fields)
}
