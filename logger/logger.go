package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	// DEBUG debug level 0
	DEBUG = iota
	// INFO info level 1
	INFO
	// WARN warn level 2
	WARN
	// ERROR error level 4
	ERROR
)

const (
	infoStr  string = "INFO "
	debugStr string = "DEBUG "
	warnStr  string = "WARN "
	errorStr string = "ERROR "
)

// Ailog just std logger wrapper.
type Ailog struct {
	logger *log.Logger
	level  int
}

// NewLogger create a new ailog.
func NewLogger(writer io.Writer) *Ailog {
	return &Ailog{logger: log.New(writer, "", log.LstdFlags), level: DEBUG}
}

// GetInstance return pub instance
func GetInstance() *Ailog {
	return ailog
}

// SetPrefix set  log prefix
// The prefix will appears at the beginning of each generated log line.
func SetPrefix(prefix string) {
	ailog.logger.SetPrefix(prefix + " ")
}

// SetPrefix set pub ailog log prefix
// The prefix will appears at the beginning of each generated log line.
func (ailog *Ailog) SetPrefix(prefix string) {
	ailog.logger.SetPrefix(prefix + " ")
}

var ailog = NewLogger(os.Stdout)

// SetLevel set ailog level.
// if log level small than set level, the log will not output to output destination.
func (ailog *Ailog) SetLevel(l int) {
	ailog.level = l
}

// SetLevel set pub ailog level.
// if log level small than set level, the log will not output to output destination.
func SetLevel(l int) {
	ailog.level = l
}

// Info print info msg
func (ailog *Ailog) Info(msg ...interface{}) {
	if ailog.level > INFO {
		return
	}
	ailog.logger.Print(infoStr, fmt.Sprint(msg...))
}

// Infof print format info msg
func (ailog *Ailog) Infof(formater string, msg ...interface{}) {
	if ailog.level > INFO {
		return
	}

	ailog.logger.Printf(fmt.Sprint(infoStr, formater), msg...)
}

// Info pub ailog print info msg
func Info(msg ...interface{}) {
	if ailog.level > INFO {
		return
	}
	ailog.logger.Print(infoStr, fmt.Sprint(msg...))
}

// Infof pub ailog print format info msg
func Infof(formater string, msg ...interface{}) {
	if ailog.level > INFO {
		return
	}
	ailog.logger.Printf(fmt.Sprint(infoStr, formater), msg...)
}

// Debug print debug msg
func (ailog *Ailog) Debug(msg ...interface{}) {
	if ailog.level > DEBUG {
		return
	}
	ailog.logger.Print(debugStr, fmt.Sprint(msg...))
}

// Debugf print format debug msg
func (ailog *Ailog) Debugf(formater string, msg ...interface{}) {
	if ailog.level > DEBUG {
		return
	}

	ailog.logger.Printf(fmt.Sprint(debugStr, formater), msg...)
}

// Debug pub ailog print debug msg
func Debug(msg ...interface{}) {
	if ailog.level > DEBUG {
		return
	}
	ailog.logger.Print(debugStr, fmt.Sprint(msg...))
}

// Debugf pub ailog print format debug msg
func Debugf(formater string, msg ...interface{}) {
	if ailog.level > DEBUG {
		return
	}

	ailog.logger.Printf(fmt.Sprint(debugStr, formater), msg...)
}

// Warn print warn msg
func (ailog *Ailog) Warn(msg ...interface{}) {
	if ailog.level > WARN {
		return
	}
	ailog.logger.Print(warnStr, fmt.Sprint(msg...))
}

// Warnf print format warn msg
func (ailog *Ailog) Warnf(formater string, msg ...interface{}) {
	if ailog.level > WARN {
		return
	}

	ailog.logger.Printf(fmt.Sprint(warnStr, formater), msg...)
}

// Warn pub ailog print warn msg
func Warn(msg ...interface{}) {
	if ailog.level > WARN {
		return
	}
	ailog.logger.Print(warnStr, fmt.Sprint(msg...))
}

// Warnf pub ailog print format warn msg
func Warnf(formater string, msg ...interface{}) {
	if ailog.level > WARN {
		return
	}

	ailog.logger.Printf(fmt.Sprint(warnStr, formater), msg...)
}

// Error print error msg
func (ailog *Ailog) Error(msg ...interface{}) {
	if ailog.level > ERROR {
		return
	}
	ailog.logger.Print(errorStr, fmt.Sprint(msg...))
}

// Errorf print format error msg
func (ailog *Ailog) Errorf(formater string, msg ...interface{}) {
	if ailog.level > ERROR {
		return
	}

	ailog.logger.Printf(fmt.Sprint(errorStr, formater), msg...)
}

// Error pub ailog print error msg
func Error(msg ...interface{}) {
	if ailog.level > ERROR {
		return
	}
	ailog.logger.Print(errorStr, fmt.Sprint(msg...))
}

// Errorf pub ailog print format error msg
func Errorf(formater string, msg ...interface{}) {
	if ailog.level > ERROR {
		return
	}

	ailog.logger.Printf(fmt.Sprint(errorStr, formater), msg...)
}
