package l3

import (
	"bufio"
	"io"
	"os"
	"sync"
)

// ConsoleWriter struct
type ConsoleWriter struct {
	mu                                                            sync.Mutex
	errorWriter, warnWriter, infoWriter, debugWriter, traceWriter io.Writer
}

// InitConfig ConsoleWriter
func (cw *ConsoleWriter) InitConfig(w *WriterConfig) {
	if w.Console.WriteErrToStdOut {
		cw.errorWriter = bufio.NewWriter(os.Stdout)
	} else {
		cw.errorWriter = bufio.NewWriter(os.Stderr)
	}
	if w.Console.WriteWarnToStdOut {
		cw.warnWriter = bufio.NewWriter(os.Stdout)
	} else {
		cw.warnWriter = bufio.NewWriter(os.Stderr)
	}

	cw.infoWriter = bufio.NewWriter(os.Stdout)
	cw.debugWriter = bufio.NewWriter(os.Stdout)
	cw.traceWriter = bufio.NewWriter(os.Stdout)

}

// DoLog consoleWriter
func (cw *ConsoleWriter) DoLog(logMsg *LogMessage) {
	var writer io.Writer

	switch logMsg.Level {
	case Off:
		break
	case Err:
		writer = cw.errorWriter
	case Warn:
		writer = cw.warnWriter
	case Info:
		writer = cw.infoWriter
	case Debug:
		writer = cw.debugWriter
	case Trace:
		writer = cw.traceWriter
	}

	if writer != nil {
		cw.mu.Lock()
		writeLogMsg(writer, logMsg)
		cw.mu.Unlock()
	}
}

// Close flushes and closes the ConsoleWriter.
func (cw *ConsoleWriter) Close() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	for _, w := range []io.Writer{cw.errorWriter, cw.warnWriter, cw.infoWriter, cw.debugWriter, cw.traceWriter} {
		if bw, ok := w.(*bufio.Writer); ok {
			_ = bw.Flush()
		}
	}
	return nil
}
