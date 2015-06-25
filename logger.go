package rlog

import (
	"log"
	"log/syslog"
	"os"
	"raoptimus/rlog/mongodb"
)

type LoggerType int

const (
	LoggerTypeStd LoggerType = iota
	LoggerTypeFile
	LoggerTypeMongoDb
	LoggerTypeSyslog
)

type (
	stdWriter struct {
	}
	writer interface {

		// Emerg logs a message with severity LOG_EMERG, ignoring the severity
		// passed to New.
		Emerg(m string) (err error)

		// Alert logs a message with severity LOG_ALERT, ignoring the severity
		// passed to New.
		Alert(m string) (err error)

		// Crit logs a message with severity LOG_CRIT, ignoring the severity
		// passed to New.
		Crit(m string) (err error)

		// Err logs a message with severity LOG_ERR, ignoring the severity
		// passed to New.
		Err(m string) (err error)

		// Warning logs a message with severity LOG_WARNING, ignoring the
		// severity passed to New.
		Warning(m string) (err error)

		// Notice logs a message with severity LOG_NOTICE, ignoring the
		// severity passed to New.
		Notice(m string) (err error)

		// Info logs a message with severity LOG_INFO, ignoring the severity
		// passed to New.
		Info(m string) (err error)

		// Debug logs a message with severity LOG_DEBUG, ignoring the severity
		// passed to New.
		Debug(m string) (err error)
	}

	Logger struct {
		*log.Logger
		writer
	}
)

func NewLogger(t LoggerType, config string) (*Logger, error) {
	return NewLoggerDial(t, "", "", "")
}

func NewLoggerDial(t LoggerType, network, raddrOrUrl, tag string) (*Logger, error) {
	switch t {
	case LoggerTypeMongoDb:
		{
			w, err := mongodb.Dial(raddrOrUrl, mongodb.LOG_EMERG, tag)
			if err != nil {
				return nil, err
			}
			lg := log.New(w, "", log.LstdFlags)
			ret := &Logger{}
			ret.writer = w
			ret.Logger = lg
			return ret, nil
		}
	case LoggerTypeSyslog:
		{
			w, err := syslog.Dial(network, raddrOrUrl, syslog.LOG_EMERG, tag)
			if err != nil {
				return nil, err
			}
			lg := log.New(w, "", log.LstdFlags)
			ret := &Logger{}
			ret.writer = w
			ret.Logger = lg
			return ret, nil
		}
	default:
		{
			ret := &Logger{}
			ret.Logger = log.New(os.Stderr, "", log.LstdFlags)
			ret.writer = &stdWriter{}
			return ret, nil
		}
	}
}

func (s *stdWriter) Emerg(m string) (err error) {
	log.Println("emerg->", m)
	return nil
}

func (s *stdWriter) Alert(m string) (err error) {
	log.Println("alert->", m)
	return nil
}
func (s *stdWriter) Crit(m string) (err error) {
	log.Println("critical->", m)
	return nil
}
func (s *stdWriter) Err(m string) (err error) {
	log.Println("error->", m)
	return nil
}
func (s *stdWriter) Warning(m string) (err error) {
	log.Println("warning->", m)
	return nil
}
func (s *stdWriter) Notice(m string) (err error) {
	log.Println("notice->", m)
	return nil
}
func (s *stdWriter) Info(m string) (err error) {
	log.Println("info->", m)
	return nil
}
func (s *stdWriter) Debug(m string) (err error) {
	log.Println("debug->", m)
	return nil
}
