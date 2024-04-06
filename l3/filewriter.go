package l3

import (
	"io"
	"os"

	"oss.nandlabs.io/golly/textutils"
)

// FileWriter struct
type FileWriter struct {
	errorWriter, warnWriter, infoWriter, debugWriter, traceWriter *os.File
}

// InitConfig FileWriter
func (fw *FileWriter) InitConfig(w *WriterConfig) {

	var defaultWriter *os.File
	if w.File.DefaultPath != textutils.EmptyStr {
		defaultWriter, _ = os.OpenFile(w.File.DefaultPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	}
	if w.File.ErrorPath != textutils.EmptyStr {
		writeLog(os.Stderr, w.File.ErrorPath)

		fw.errorWriter, _ = os.OpenFile(w.File.ErrorPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	}
	if w.File.WarnPath != textutils.EmptyStr {
		fw.warnWriter, _ = os.OpenFile(w.File.WarnPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	}
	if w.File.InfoPath != textutils.EmptyStr {
		fw.infoWriter, _ = os.OpenFile(w.File.InfoPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	}
	if w.File.DebugPath != textutils.EmptyStr {
		fw.debugWriter, _ = os.OpenFile(w.File.DebugPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	}
	if w.File.TracePath != textutils.EmptyStr {
		fw.traceWriter, _ = os.OpenFile(w.File.TracePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	}
	if defaultWriter != nil {
		if fw.errorWriter == nil {
			fw.errorWriter = defaultWriter
		}
		if fw.warnWriter == nil {
			fw.warnWriter = defaultWriter
		}
		if fw.infoWriter == nil {
			fw.infoWriter = defaultWriter
		}
		if fw.debugWriter == nil {
			fw.debugWriter = defaultWriter
		}
		if fw.traceWriter == nil {
			fw.traceWriter = defaultWriter
		}
	}
}

// DoLog FileWriter
func (fw *FileWriter) DoLog(logMsg *LogMessage) {
	var writer io.Writer
	switch logMsg.Level {
	case Off:
		return
	case Err:
		writer = fw.errorWriter
	case Warn:
		writer = fw.warnWriter
	case Info:
		writer = fw.infoWriter
	case Debug:
		writer = fw.debugWriter
	case Trace:
		writer = fw.traceWriter
	}

	if writer != nil {
		writeLogMsg(writer, logMsg)
	}
}

// Close stream
func (fw *FileWriter) Close() error {
	return fw.debugWriter.Close()
}
