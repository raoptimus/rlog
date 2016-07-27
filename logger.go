package rlog

import (
	"github.com/raoptimus/rlog/mongodb"
	"log"
	"log/syslog"
	"os"
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
		flag uint32
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

const (
	LOG_ALL uint32 = 1 << (32 - 1 - iota)
	LOG_EMERG
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

func NewLogger(t LoggerType, config string, flag uint32) (*Logger, error) {
	return NewLoggerDial(t, "", "", "", flag)
}

func NewLoggerDial(t LoggerType, network, addrOrUrl, tag string, flag uint32) (*Logger, error) {
	switch t {
	case LoggerTypeMongoDb:
		{
			w, err := mongodb.Dial(addrOrUrl, mongodb.LOG_EMERG, tag)
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
			w, err := syslog.Dial(network, addrOrUrl, syslog.LOG_EMERG, tag)
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
			ret := &Logger{
				Logger: log.New(os.Stderr, "", log.LstdFlags),
				writer: &stdWriter{
					flag: flag,
				},
			}
			return ret, nil
		}
	}
}

func (s *stdWriter) Emerg(m string) (err error) {
	if s.flag&LOG_ALL|s.flag&LOG_EMERG != 0 {
		log.Println("emerg->", m)
	}
	return nil
}

func (s *stdWriter) Alert(m string) (err error) {
	if s.flag&LOG_ALL|s.flag&LOG_ALERT != 0 {
		log.Println("alert->", m)
	}
	return nil
}
func (s *stdWriter) Crit(m string) (err error) {
	if s.flag&LOG_ALL|s.flag&LOG_CRIT != 0 {
		log.Println("critical->", m)
	}
	return nil
}
func (s *stdWriter) Err(m string) (err error) {
	if s.flag&LOG_ALL|s.flag&LOG_ERR != 0 {
		log.Println("error->", m)
	}
	return nil
}
func (s *stdWriter) Warning(m string) (err error) {
	if s.flag&LOG_ALL|s.flag&LOG_WARNING != 0 {
		log.Println("warning->", m)
	}
	return nil
}
func (s *stdWriter) Notice(m string) (err error) {
	if s.flag&LOG_ALL|s.flag&LOG_NOTICE != 0 {
		log.Println("notice->", m)
	}
	return nil
}
func (s *stdWriter) Info(m string) (err error) {
	if s.flag&LOG_ALL|s.flag&LOG_INFO != 0 {
		log.Println("info->", m)
	}
	return nil
}
func (s *stdWriter) Debug(m string) (err error) {
	if s.flag&LOG_ALL|s.flag&LOG_DEBUG != 0 {
		log.Println("debug->", m)
	}
	return nil
}
