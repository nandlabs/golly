package l3

const (
	//Off - No logging
	Off Level = iota
	//Err - logging only for error level.
	Err
	//Warn - logging turned on for warning & error levels
	Warn
	//Info - logging turned on for Info, Warning and Error levels.
	Info
	//Debug - Logging turned on for Debug, Info, Warning and Error levels.
	Debug
	//Trace - Logging turned on for Trace,Info, Warning and error Levels.
	Trace
)

//Level specifies the log level
type Level int

//Levels of the logging by severity
var Levels = [...]string{
	"OFF",
	"ERROR",
	"WARN",
	"INFO",
	"DEBUG",
	"TRACE",
}

// LevelsBytes  of the logging by Level
var LevelsBytes = [...][]byte{
	[]byte("OFF"),
	[]byte("ERROR"),
	[]byte("WARN"),
	[]byte("INFO"),
	[]byte("DEBUG"),
	[]byte("TRACE"),
}

//LevelsMap of the logging by Level string Level type
var LevelsMap = map[string]Level{
	"OFF":   Off,
	"ERROR": Err,
	"WARN":  Warn,
	"INFO":  Info,
	"DEBUG": Debug,
	"TRACE": Trace,
}

type Logger interface {
	Error(a ...interface{})
	ErrorF(f string, a ...interface{})
	Warn(a ...interface{})
	WarnF(f string, a ...interface{})
	Info(a ...interface{})
	InfoF(f string, a ...interface{})
	Debug(a ...interface{})
	DebugF(f string, a ...interface{})
	Trace(a ...interface{})
	TraceF(f string, a ...interface{})
}
