package goscope2

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/golang/glog"
)

const (
	SEVERITY_INFO    = "INFO"
	SEVERITY_WARNING = "WARNING"
	SEVERITY_ERROR   = "ERROR"
	SEVERITY_FATAL   = "FATAL"
)

var validate = validator.New()

func notify(config Config, severity string, format string, args ...any) {
	if err := validate.Var(severity, "required,oneof=INFO WARNING ERROR FATAL"); err != nil {
		notify(config, SEVERITY_FATAL, "invalid severity on error %w", fmt.Errorf(format, args...))
		return
	}
	message := fmt.Sprintf(format, args...)
	checkAndPurge(config.DB, config.LimitLogs)
	config.DB.Create(&Goscope2Log{
		App:      config.InternalApp,
		Hash:     generateMessageHash(message),
		Severity: severity,
		Message:  message,
	})
}

func Infof(format string, args ...any) {
	go notify(*Goscope2.config, SEVERITY_INFO, format, args...)
	glog.Infof(format, args...)
}
func Warningf(format string, args ...any) {
	go notify(*Goscope2.config, SEVERITY_WARNING, format, args...)
	glog.Warningf(format, args...)
}
func Errorf(format string, args ...any) {
	go notify(*Goscope2.config, SEVERITY_ERROR, format, args...)
	glog.Errorf(format, args...)
}
func Fatalf(format string, args ...any) {
	notify(*Goscope2.config, SEVERITY_FATAL, format, args...)
	glog.Flush()
	glog.Fatalf(format, args...)
}
