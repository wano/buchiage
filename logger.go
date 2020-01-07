package buchiage

import "github.com/labstack/gommon/log"

type Logger interface {
	Info(...interface{})
	Error(...interface{})
}

type buchiageLogger struct {

}

func (b *buchiageLogger) Info(i ...interface{}) {
	log.Info(i...)
}

func (b *buchiageLogger) Error(i ...interface{}) {
	log.Error(i...)
}

var globalLogger = &buchiageLogger{}