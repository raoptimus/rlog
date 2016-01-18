package mongodb

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const severityMask = 0x07
const facilityMask = 0xf8

type Priority int

const (
	LOG_EMERG Priority = iota
	LOG_ALERT
	LOG_CRIT
	LOG_ERR
	LOG_WARNING
	LOG_NOTICE
	LOG_INFO
	LOG_DEBUG
)

// A Writer is a connection to a syslog server.
type Writer struct {
	priority Priority
	tag      string
	hostname string
	url      string

	mu   sync.Mutex // guards conn
	conn serverConn
}

type Log struct {
	Id       bson.ObjectId `bson:"_id"`
	Priority int           `bson:"Priority"`
	Time     time.Time     `bson:"Time"`
	Hostname string        `bson:"Hostname"`
	Tag      string        `bson:"Tag"`
	Msg      string        `bson:"Msg"`
	Pid      int           `bson:"Pid"`
}

type serverConn interface {
	writeString(p Priority, hostname, tag, s, nl string) error
	close() error
}

type netConn struct {
	conn *mgo.Session
	col  *mgo.Collection
}

// New establishes a new connection to the system log daemon.  Each
// write to the returned writer sends a log message with the given
// priority and prefix.
func New(priority Priority, tag string) (w *Writer, err error) {
	return Dial("", priority, tag)
}

// If url is empty, Dial will connect to the local mongodb server.
func Dial(url string, priority Priority, tag string) (*Writer, error) {
	if priority < 0 || priority > LOG_DEBUG {
		return nil, errors.New("rlog/mongodblog: invalid priority")
	}

	if tag == "" {
		tag = os.Args[0]
	}
	hostname, _ := os.Hostname()

	w := &Writer{
		priority: priority,
		tag:      tag,
		hostname: hostname,
		url:      url,
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.connect()
	if err != nil {
		return nil, err
	}
	return w, err
}

// connect makes a connection to the mongodb log server.
func (w *Writer) connect() error {
	if w.conn != nil {
		// ignore err from close, it makes sense to continue anyway
		w.conn.close()
		w.conn = nil
	}

	if w.hostname == "" {
		w.hostname = "localhost"
	}
	if w.url == "" {
		w.url = w.hostname + "/rlogs"
	}

	u, err := url.Parse("mongodb://" + w.url)
	if err != nil {
		errors.New("Connection string (" + w.url + ") of mongodb is not correct: " + err.Error())
	}

	options := u.Query()
	options.Del("w")
	options.Del("readPreference")

	log.Println(u.Host + u.Path + "?" + options.Encode())
	session, err := mgo.Dial(u.Host + u.Path + "?" + options.Encode())
	if err != nil {
		return errors.New("Can't connect to mongodb (" + w.url + "): " + err.Error())
	}

	session.SetSafe(&mgo.Safe{
		W: -1,
	})

	switch {
	case options.Get("replicaSet") != "":
		session.SetMode(mgo.Monotonic, true)
	default:
		session.SetMode(mgo.Strong, true)
	}

	c := session.DB("").C("Log")
	err = c.Create(&mgo.CollectionInfo{
		Capped:   true,
		MaxDocs:  10000,
		MaxBytes: 5242880,
	})
	if err != nil {
		if err.Error() != "collection already exists" {
			return errors.New("Can't create the mongo collection Log: " + err.Error())
		}
	}
	ixs, _ := c.Indexes()
	for _, ix := range ixs {
		if ix.ExpireAfter > 0 {
			c.DropIndexName(ix.Name)
			break
		}
	}

	c.EnsureIndex(mgo.Index{
		Key:        []string{"-Time"},
		Background: true,
		//		ExpireAfter: time.Duration(30 * 24 * time.Hour), //do not supported with capped collections
	})
	w.conn = &netConn{conn: session, col: c}
	return nil
}

// Write sends a log message to the syslog daemon.
func (w *Writer) Write(b []byte) (int, error) {
	return w.writeAndRetry(w.priority, string(b))
}

// Close closes a connection to the syslog daemon.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		err := w.conn.close()
		w.conn = nil
		return err
	}
	return nil
}

// Emerg logs a message with severity LOG_EMERG, ignoring the severity
// passed to New.
func (w *Writer) Emerg(m string) (err error) {
	_, err = w.writeAndRetry(LOG_EMERG, m)
	return err
}

// Alert logs a message with severity LOG_ALERT, ignoring the severity
// passed to New.
func (w *Writer) Alert(m string) (err error) {
	_, err = w.writeAndRetry(LOG_ALERT, m)
	return err
}

// Crit logs a message with severity LOG_CRIT, ignoring the severity
// passed to New.
func (w *Writer) Crit(m string) (err error) {
	_, err = w.writeAndRetry(LOG_CRIT, m)
	return err
}

// Err logs a message with severity LOG_ERR, ignoring the severity
// passed to New.
func (w *Writer) Err(m string) (err error) {
	_, err = w.writeAndRetry(LOG_ERR, m)
	return err
}

// Warning logs a message with severity LOG_WARNING, ignoring the
// severity passed to New.
func (w *Writer) Warning(m string) (err error) {
	_, err = w.writeAndRetry(LOG_WARNING, m)
	return err
}

// Notice logs a message with severity LOG_NOTICE, ignoring the
// severity passed to New.
func (w *Writer) Notice(m string) (err error) {
	_, err = w.writeAndRetry(LOG_NOTICE, m)
	return err
}

// Info logs a message with severity LOG_INFO, ignoring the severity
// passed to New.
func (w *Writer) Info(m string) (err error) {
	_, err = w.writeAndRetry(LOG_INFO, m)
	return err
}

// Debug logs a message with severity LOG_DEBUG, ignoring the severity
// passed to New.
func (w *Writer) Debug(m string) (err error) {
	_, err = w.writeAndRetry(LOG_DEBUG, m)
	return err
}

func (w *Writer) writeAndRetry(p Priority, s string) (int, error) {
	pr := (w.priority & facilityMask) | (p & severityMask)

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		if n, err := w.write(pr, s); err == nil {
			return n, err
		}
	}
	if err := w.connect(); err != nil {
		return 0, err
	}
	return w.write(pr, s)
}

// write generates and writes a syslog formatted string. The
// format is as follows: <PRI>TIMESTAMP HOSTNAME TAG[PID]: MSG
func (w *Writer) write(p Priority, msg string) (int, error) {
	// ensure it ends in a \n
	nl := ""
	if !strings.HasSuffix(msg, "\n") {
		nl = "\n"
	}

	err := w.conn.writeString(p, w.hostname, w.tag, msg, nl)
	if err != nil {
		return 0, err
	}
	// Note: return the length of the input, not the number of
	// bytes printed by Fprintf, because this must behave like
	// an io.Writer.
	return len(msg), nil
}

func (n *netConn) writeString(p Priority, hostname, tag, msg, nl string) error {
	log := Log{
		Id:       bson.NewObjectId(),
		Priority: int(p),
		Time:     time.Now().UTC(),
		Hostname: hostname,
		Tag:      tag,
		Msg:      msg,
		Pid:      os.Getpid(),
	}
	return n.col.Insert(log)
}

func (n *netConn) close() error {
	n.conn.Close()
	return nil
}

// NewLogger creates a log.Logger whose output is written to
// the system log service with the specified priority. The logFlag
// argument is the flag set passed through to log.New to create
// the Logger.
func NewLogger(p Priority, logFlag int) (*log.Logger, error) {
	s, err := New(p, "")
	if err != nil {
		return nil, err
	}
	return log.New(s, "", logFlag), nil
}
