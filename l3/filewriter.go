package l3

import (
	"io"
	"os"
	"sync"

	"oss.nandlabs.io/golly/textutils"
)

// FileWriter struct
type FileWriter struct {
	mu                                                            sync.Mutex
	errorWriter, warnWriter, infoWriter, debugWriter, traceWriter *os.File
}

// InitConfig FileWriter
func (fw *FileWriter) InitConfig(w *WriterConfig) {

	var defaultWriter *os.File
	var err error
	if w.File.DefaultPath != textutils.EmptyStr {
		defaultWriter, err = os.OpenFile(w.File.DefaultPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			writeLog(os.Stderr, "l3: unable to open default log file:", w.File.DefaultPath, err)
		}
	}
	if w.File.ErrorPath != textutils.EmptyStr {
		fw.errorWriter, err = os.OpenFile(w.File.ErrorPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			writeLog(os.Stderr, "l3: unable to open error log file:", w.File.ErrorPath, err)
		}
	}
	if w.File.WarnPath != textutils.EmptyStr {
		fw.warnWriter, err = os.OpenFile(w.File.WarnPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			writeLog(os.Stderr, "l3: unable to open warn log file:", w.File.WarnPath, err)
		}
	}
	if w.File.InfoPath != textutils.EmptyStr {
		fw.infoWriter, err = os.OpenFile(w.File.InfoPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			writeLog(os.Stderr, "l3: unable to open info log file:", w.File.InfoPath, err)
		}
	}
	if w.File.DebugPath != textutils.EmptyStr {
		fw.debugWriter, err = os.OpenFile(w.File.DebugPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			writeLog(os.Stderr, "l3: unable to open debug log file:", w.File.DebugPath, err)
		}
	}
	if w.File.TracePath != textutils.EmptyStr {
		fw.traceWriter, err = os.OpenFile(w.File.TracePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			writeLog(os.Stderr, "l3: unable to open trace log file:", w.File.TracePath, err)
		}
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
		fw.mu.Lock()
		writeLogMsg(writer, logMsg)
		fw.mu.Unlock()
	}
}

// Close closes all open file handles.
func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	// Deduplicate: multiple levels may share the same file.
	closed := make(map[*os.File]struct{})
	for _, f := range []*os.File{fw.errorWriter, fw.warnWriter, fw.infoWriter, fw.debugWriter, fw.traceWriter} {
		if f == nil {
			continue
		}
		if _, ok := closed[f]; ok {
			continue
		}
		closed[f] = struct{}{}
		_ = f.Close()
	}
	return nil
}
