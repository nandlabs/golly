package l3

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"oss.nandlabs.io/golly/config"
	"oss.nandlabs.io/golly/fsutils"
	"oss.nandlabs.io/golly/textutils"
)

const (
	// LogConfigEnvProperty specifies the environment variable that would specify the file location
	LogConfigEnvProperty = "GC_LOG_CONFIG_FILE"
	//DefaultlogFilePath specifies the location where the application should search for log config if the LogConfigEnvProperty is not specified
	DefaultlogFilePath = "./log-config.json"
	//newLineBytes
)

// LogWriter interface
type LogWriter interface {
	InitConfig(w *WriterConfig)
	DoLog(logMsg *LogMessage)
	io.Closer
}

// BaseLogger struct.
type BaseLogger struct {
	level           Level
	pkgName         string
	errorEnabled    bool
	warnEnabled     bool
	infoEnabled     bool
	debugEnabled    bool
	traceEnabled    bool
	includeFunction bool
	includeLine     bool
}

// Map to hold loggers. This is updated in case the log config is reloaded
var loggers = make(map[string]*BaseLogger)

// Writers can be multiple writers
var writers []LogWriter

// Log configuration
var logConfig *LogConfig

// channel of type log message
var logMsgChannel chan *LogMessage

var mutex = &sync.Mutex{}

var newLineBytes = []byte("\n") // TODO Check for windows
var whiteSpaceBytes = []byte(textutils.WhiteSpaceStr)

func init() {
	Configure(loadConfig())
}

// Configure Logging
func Configure(l *LogConfig) {
	mutex.Lock()
	defer mutex.Unlock()
	logConfig = l
	if l.DatePattern == "" {
		l.DatePattern = time.RFC3339
	}
	if l.Async {

		if l.QueueSize == 0 {
			l.QueueSize = 4096
		}
		logMsgChannel = make(chan *LogMessage, l.QueueSize)
		go doAsyncLog()
	}
	if l.Writers != nil {
		for _, w := range l.Writers {
			if w.File != nil {
				fw := &FileWriter{}
				fw.InitConfig(w)
				writers = append(writers, fw)
			} else if w.Console != nil {
				cw := &ConsoleWriter{}
				cw.InitConfig(w)
				writers = append(writers, cw)
			}

		}
	}
}

// Update the flags based on the severity level
func (l *BaseLogger) updateLvlFlags() error {

	if l.level < 0 || l.level > 5 {
		return errors.New("Invalid Log Level  ")
	}
	l.errorEnabled = l.level >= 1
	l.warnEnabled = l.level >= 2
	l.infoEnabled = l.level >= 3
	l.debugEnabled = l.level >= 4
	l.traceEnabled = l.level == 5
	return nil
}

// loadDefaultConfig function with load the default configuration
func loadDefaultConfig() *LogConfig {
	isAsync, _ := config.GetEnvAsBool("GC_LOG_ASYNC", false)
	errToStdOut, _ := config.GetEnvAsBool("GC_LOG_ERR_STDOUT", false)
	warnToStdOut, _ := config.GetEnvAsBool("GC_LOG_WARN_STDOUT", false)

	return &LogConfig{
		Format:      config.GetEnvAsString("GC_LOG_FMT", "text"),
		Async:       isAsync,
		DatePattern: config.GetEnvAsString("GC_LOG_TIME_FMT", time.RFC3339),
		DefaultLvl:  config.GetEnvAsString("GC_LOG_DEF_LEVEL", "INFO"),
		Writers: []*WriterConfig{
			{
				Console: &ConsoleConfig{
					WriteErrToStdOut:  errToStdOut,
					WriteWarnToStdOut: warnToStdOut,
				},
			},
		},
	}
}

// loadConfig function will load the log configuration.
func loadConfig() *LogConfig {
	var logConfig = &LogConfig{}
	fileName := config.GetEnvAsString(LogConfigEnvProperty, DefaultlogFilePath)
	if fsutils.FileExists(fileName) {
		contentType := fsutils.LookupContentType(fileName)
		if contentType == "application/json" {
			logConfigFile, err := os.Open(fileName)
			if err != nil {
				writeLog(os.Stderr, "Unable to open the log config file using default log configuration", err)
				logConfig = loadDefaultConfig()
			} else {
				defer logConfigFile.Close()
				bytes, _ := io.ReadAll(logConfigFile)
				err = json.Unmarshal(bytes, &logConfig)
				if err != nil {
					writeLog(os.Stderr, "Unable to open the log config file using default log config", err)
					logConfig = loadDefaultConfig()
				}
			}
		} else {
			writeLog(os.Stderr, "Invalid file format supported format : application/json . Loading Default configuration")
			logConfig = loadDefaultConfig()
		}
		//TODO Add yaml support once its available
	} else {
		writeLog(os.Stderr, "Log Config file not found. Loading default configuration")
		logConfig = loadDefaultConfig()
	}
	return logConfig
}

// Get function will return the logger object for that package
func Get() Logger {
	mutex.Lock()
	defer mutex.Unlock()
	pc, _, _, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	fnNameSplit := strings.Split(details.Name(), textutils.ForwardSlashStr)
	pkgFnName := strings.Split(fnNameSplit[len(fnNameSplit)-1], textutils.PeriodStr)
	pkgName := pkgFnName[0]

	if _, ok := loggers[pkgName]; !ok {
		Level := logConfig.DefaultLvl

		if logConfig.PkgConfigs != nil && len(logConfig.PkgConfigs) > 0 {
			for _, pkgConfig := range logConfig.PkgConfigs {
				if pkgConfig.PackageName == pkgName {
					Level = pkgConfig.Level
				}
			}
		}

		logger := &BaseLogger{
			level:           LevelsMap[Level],
			pkgName:         pkgName,
			includeFunction: logConfig.IncludeFunction,
			includeLine:     logConfig.IncludeLineNum,
		}
		_ = logger.updateLvlFlags()
		loggers[pkgName] = logger
	}

	return loggers[pkgName]
}

func writeLogMsg(writer io.Writer, logMsg *LogMessage) {
	if logConfig.Format == "json" {
		//TODO update marshalling to direct field access to avoid reflection.
		//This will be based on codec branch.
		data, _ := json.Marshal(logMsg)
		_, _ = writer.Write(data)

	} else if logConfig.Format == "text" {
		buf := bufio.NewWriter(writer)

		if logMsg.FnName != textutils.EmptyStr {

			//writeLog(writer, logMsg.Time.Format(logConfig.DatePattern), Levels[logMsg.Level], logMsg.FnName+":"+strconv.Itoa(logMsg.Line), logMsg.Content.String())

			_, _ = buf.Write(formatTimeToBytes(logMsg.Time, logConfig.DatePattern))
			_, _ = buf.Write(whiteSpaceBytes)
			_, _ = buf.Write(LevelsBytes[logMsg.Level])
			_, _ = buf.Write(whiteSpaceBytes)
			_, _ = buf.WriteString(logMsg.FnName)
			_, _ = buf.WriteString(textutils.ColonStr)
			_, _ = buf.WriteString(strconv.Itoa(logMsg.Line))
			_, _ = buf.Write(whiteSpaceBytes)
			_, _ = buf.Write(logMsg.Content.Bytes())
			_, _ = buf.Write(newLineBytes)

		} else {
			//writeLog(writer, logMsg.Time.Format(logConfig.DatePattern), Levels[logMsg.Level],  logMsg.Content.String())

			_, _ = buf.Write(formatTimeToBytes(logMsg.Time, logConfig.DatePattern))
			_, _ = buf.Write(whiteSpaceBytes)
			_, _ = buf.Write(LevelsBytes[logMsg.Level])
			_, _ = buf.Write(whiteSpaceBytes)
			_, _ = buf.Write(logMsg.Content.Bytes())
			_, _ = buf.Write(newLineBytes)
		}
		_ = buf.Flush()

	}
}

func formatTimeToBytes(t time.Time, layout string) []byte {

	b := make([]byte, 0, len(layout))
	return t.AppendFormat(b, layout)
}

// createLogMessage function creates a new log message with actual content variables
func handleLog(l *BaseLogger, logMsg *LogMessage) {
	if l.includeFunction {
		pc, _, no, _ := runtime.Caller(2)
		details := runtime.FuncForPC(pc)
		fnNameSplit := strings.Split(details.Name(), "/")
		logMsg.FnName = fnNameSplit[len(fnNameSplit)-1]
		if l.includeLine {
			logMsg.Line = no
		}
	}

	if logConfig.Async {
		logMsgChannel <- logMsg
	} else {
		doLog(logMsg)
	}
}

func doLog(logMsg *LogMessage) {
	for _, w := range writers {
		w.DoLog(logMsg)
	}
	putLogMessage(logMsg)
}

func doAsyncLog() {

	for logMsg := range logMsgChannel {

		doLog(logMsg)

	}

}

// writeLog will write to the io.Writer interface
func writeLog(w io.Writer, a ...interface{}) {
	//TODO check error handling here
	_, _ = fmt.Fprintln(w, a...)
}

// String method to get the Severity String
func (sev Level) String() (string, error) {
	if sev < 0 || sev > 5 {
		return "", errors.New("Invalid severity ")
	}
	return Levels[sev], nil
}

// IsEnabled function returns if the current
func (l *BaseLogger) IsEnabled(sev Level) bool {
	return sev <= Trace && sev >= l.level
}

// Error BaseLogger
func (l *BaseLogger) Error(a ...interface{}) {
	if l.errorEnabled && a != nil && len(a) > 0 {
		handleLog(l, getLogMessage(Err, a...))
	}
}

// ErrorF BaseLogger with formatting of the messages
func (l *BaseLogger) ErrorF(f string, a ...interface{}) {
	if l.errorEnabled {
		handleLog(l, getLogMessageF(Err, f, a...))
	}
}

// Warn BaseLogger
func (l *BaseLogger) Warn(a ...interface{}) {
	if l.warnEnabled && a != nil && len(a) > 0 {
		handleLog(l, getLogMessage(Warn, a...))
	}
}

// WarnF BaseLogger with formatting of the messages
func (l *BaseLogger) WarnF(f string, a ...interface{}) {
	if l.warnEnabled {
		handleLog(l, getLogMessageF(Warn, f, a...))

	}
}

// Info BaseLogger
func (l *BaseLogger) Info(a ...interface{}) {
	if l.infoEnabled && a != nil && len(a) > 0 {
		handleLog(l, getLogMessage(Info, a...))
	}
}

// InfoF BaseLogger
func (l *BaseLogger) InfoF(f string, a ...interface{}) {
	if l.infoEnabled {
		handleLog(l, getLogMessageF(Info, f, a...))

	}
}

// Debug BaseLogger
func (l *BaseLogger) Debug(a ...interface{}) {
	if l.debugEnabled && a != nil && len(a) > 0 {
		handleLog(l, getLogMessage(Debug, a...))
	}
}

// DebugF BaseLogger
func (l *BaseLogger) DebugF(f string, a ...interface{}) {
	if l.debugEnabled {
		handleLog(l, getLogMessageF(Debug, f, a...))
	}
}

// Trace BaseLogger
func (l *BaseLogger) Trace(a ...interface{}) {
	if l.traceEnabled && a != nil && len(a) > 0 {
		handleLog(l, getLogMessage(Trace, a...))

	}
}

// TraceF BaseLogger
func (l *BaseLogger) TraceF(f string, a ...interface{}) {
	if l.traceEnabled {
		handleLog(l, getLogMessageF(Trace, f, a...))
	}
}
