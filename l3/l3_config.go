package l3

//LogConfig - Configuration & Settings for the logger.
type LogConfig struct {

	//Format of the log. valid values are text,json
	//Default is text
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
	//Async Flag to indicate if the writing of the flag is asynchronous.
	//Default value is false
	Async bool `json:"async,omitempty" yaml:"async,omitempty"`
	//QueueSize to indicate the number log routines that can be queued  to use in background
	//This value is used only if the async value is set to true.
	//Default value for the number items to be in queue 512
	QueueSize int `json:"queue_size,omitempty" yaml:"queueSize,omitempty"`
	//Date - Defaults to  time.RFC3339 pattern
	DatePattern string `json:"datePattern,omitempty" yaml:"datePattern,omitempty"`
	//IncludeFunction will include the calling function name  in the log entries
	//Default value : false
	IncludeFunction bool `json:"includeFunction,omitempty" yaml:"includeFunction,omitempty"`
	//IncludeLineNum ,includes Line number for the log file
	//If IncludeFunction Line is set to false this config is ignored
	IncludeLineNum bool `json:"includeLineNum,omitempty" yaml:"includeLineNum,omitempty"`
	//DefaultLvl that will be used as default
	DefaultLvl string `json:"defaultLvl" yaml:"defaultLvl"`
	//PackageConfig that can be used to
	PkgConfigs []*PackageConfig `json:"pkgConfigs" yaml:"pkgConfigs"`
	//Writers writers for the logger. Need one for all levels
	//If a writer is not found for a specific level it will fallback to os.Stdout if the level is greater then Warn and os.Stderr otherwise
	Writers []*WriterConfig `json:"writers" yaml:"writers"`
}

// PackageConfig configuration
type PackageConfig struct {
	//PackageName
	PackageName string `json:"pkgName" yaml:"pkgName"`
	//Level to be set valid values : OFF,ERROR,WARN,INFO,DEBUG,TRACE
	Level string `json:"level" yaml:"level"`
}

//WriterConfig struct
type WriterConfig struct {
	//File reference. Non mandatory but one of file or console logger is required.
	File *FileConfig `json:"file,omitempty" yaml:"file,omitempty"`
	//Console reference
	Console *ConsoleConfig `json:"console,omitempty" yaml:"console,omitempty"`
}

//FileConfig - Configuration of file based logging
type FileConfig struct {
	//FilePath for the file based log writer
	DefaultPath string `json:"defaultPath" yaml:"defaultPath"`
	ErrorPath   string `json:"errorPath" yaml:"errorPath"`
	WarnPath    string `json:"warnPath" yaml:"warnPath"`
	InfoPath    string `json:"infoPath" yaml:"infoPath"`
	DebugPath   string `json:"debugPath" yaml:"debugPath"`
	TracePath   string `json:"tracePath" yaml:"tracePath"`
	//RollType must indicate one for the following(case sensitive). SIZE,DAILY
	RollType string `json:"rollType" yaml:"rollType"`
	//Max Size of the of the file. Only takes into effect when the RollType="SIZE"
	MaxSize int64 `json:"maxSize" yaml:"maxSize"`
	//CompressOldFile is taken into effect if file rolling is enabled by setting a RollType.
	//Default implementation will just do a GZIP of the file leaving the file with <file_name>.gz
	CompressOldFile bool `json:"compressOldFile" yaml:"compressOldFile"`
}

// ConsoleConfig - Configuration of console based logging. All Log Levels except ERROR and WARN are written to os.Stdout
// The ERROR and WARN log levels can be written  to os.Stdout or os.Stderr, By default they go to os.Stderr
type ConsoleConfig struct {
	//WriteErrToStdOut write error messages to os.Stdout .
	WriteErrToStdOut bool `json:"errToStdOut" yaml:"errToStdOut"`
	//WriteWarnToStdOut write warn messages to os.Stdout .
	WriteWarnToStdOut bool `json:"warnToStdOut" yaml:"warnToStdOut"`
}
