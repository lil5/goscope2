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

func notify(gs *GoScope2, severity string, format string, args ...any) {
	if err := validate.Var(severity, "required,oneof=INFO WARNING ERROR FATAL"); err != nil {
		notify(gs, SEVERITY_FATAL, "invalid severity on error %w", fmt.Errorf(format, args...))
		return
	}
	message := fmt.Sprintf(format, args...)
	log := &Goscope2Log{
		Type:     TYPE_LOG,
		Severity: severity,
		Message:  message,
	}
	go func() {
		log.GenerateHash()
		gs.DB.Create(log)
	}()
}

func (gs *GoScope2) Infof(format string, args ...any) {
	go notify(gs, SEVERITY_INFO, format, args...)
	glog.Infof(format, args...)
}
func (gs *GoScope2) Warningf(format string, args ...any) {
	go notify(gs, SEVERITY_WARNING, format, args...)
	glog.Warningf(format, args...)
}
func (gs *GoScope2) Errorf(format string, args ...any) {
	go notify(gs, SEVERITY_ERROR, format, args...)
	glog.Errorf(format, args...)
}
func (gs *GoScope2) Fatalf(format string, args ...any) {
	notify(gs, SEVERITY_FATAL, format, args...)
	glog.Flush()
	glog.Fatalf(format, args...)
}
