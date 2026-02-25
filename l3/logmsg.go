package l3

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"oss.nandlabs.io/golly/textutils"
)

var logMsgPool = &sync.Pool{
	New: func() interface{} {
		lm := &LogMessage{
			Content: &bytes.Buffer{},
			Buf:     &bytes.Buffer{},
		}
		lm.Content.Grow(1024)
		lm.Buf.Grow(1280)
		return lm
	},
}

// LogMessage struct.
type LogMessage struct {
	Time    time.Time     `json:"timestamp"`
	FnName  string        `json:"function,omitempty"`
	Line    int           `json:"line,omitempty"`
	Content *bytes.Buffer `json:"msg"`
	Level   Level         `json:"level"`
	Buf     *bytes.Buffer
	// SevBytes []byte
}

func getLogMessageF(level Level, f string, v ...interface{}) *LogMessage {
	msg := logMsgPool.Get().(*LogMessage)
	msg.Level = level
	msg.Time = time.Now()
	msg.FnName = textutils.EmptyStr
	msg.Line = 0
	_, _ = fmt.Fprintf(msg.Content, f, v...)
	return msg
}

func getLogMessage(level Level, v ...interface{}) *LogMessage {
	msg := logMsgPool.Get().(*LogMessage)
	msg.Level = level
	msg.Time = time.Now()
	msg.FnName = textutils.EmptyStr
	msg.Line = 0
	_, _ = fmt.Fprint(msg.Content, v...)
	return msg
}

func putLogMessage(logMsg *LogMessage) {
	logMsg.Content.Reset()
	logMsg.Buf.Reset()
	logMsgPool.Put(logMsg)
}
