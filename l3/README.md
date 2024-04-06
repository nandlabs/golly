# go-l3
A Lightweight Levelled Logger for Go

# Features
* Multiple logging levels ```OFF,ERROR,INFO,DEBUG,TRACE```
* Console and File based writers
* Rolling file support ([WIP]
* Ability to specify log levels for a specific package
* Internationalisation (i18n) support ([WIP])
* Async logging support
* Configuration can be done using either a file,env variables,Struct values at runtime.

## Usage

### Simple Usage
The simplest example is  as shown below.
```
    import (
	"errors"
	
	"oss.nandlabs.io/golly/l3"
)

//logger Package Level Logger
var logger l3.Logger = l3.Get()

func main() {
	logger.Info("This is an info log msg")
	logger.Warn("This is an warning msg")
	logger.Error("This is an error log msg")
	logger.Error("This is an error log msg with optional error", errors.New("Some error happened here "))
	logger.Warn("Message with any level can also have an error", errors.New("Some error happened here but want it as info"))

}

    
   ```
This will create a logger with default configuration

* Logs get written to console
* Log levels  ```ERROR, WARN``` get written to stderr and remaining levels to stdout
* The default log level is ```INFO``` this can be overwritten using an env variable or log config file.
  See Log [Configuration](#Log Configuration) section for more details.

###




# Log Configuration
The below table specifies the configuration parameters for logging
The log can be configured in the following ways.

### 1. File Based Configuration
The file based configuration allows a file with log configuration to be specified. Here is a sample file configuration
based on ```json```.
```
{
     "format": "json",
     "async": false,
     "defaultLvl": "INFO",
     "includeFunction": true,
     "includeLineNum": true,
     "pkgConfigs": [
       {
         "pkgName": "main",
         "level": "INFO"
       }
     ],
     "writers": [
       {
         "console": {
           "errToStdOut": false,
           "warnToStdOut": false
         }
       }
       {
         "file": {
           "defaultPath": "/tmp/default.log"
         }
       },
       
     ]
   }
```

following table specifies the field values
|Field Name   | Type    | Description   | Default Value|
|:-|:-|:-|:-:|
|format|String| The output format of the log message. The valid values are `text` or `json`| `text` |
|async|Boolean| Determines if the message is to be written to the destination asynchronously. If set to `true` then the LogMessage is prepared synchronously.However, it is written to destination in a async fashion.|`false`|
|defaultLvl| String|Sets the default Logging level for the Logger. This is a global value. For overriding the log levels for a specific packages use `pkgConfigs`. The valid values are `OFF,ERROR,INFO,DEBUG,TRACE`| `INFO`|
|datePattern|String| The timestamp format for the log entries.Valid values are the ones acceptable by`time.Format(<pattern>)` function.|As defined by `time.RFC3339` |
|includeFunction| Boolean| Determines if the Function Name needs to be printed in logs. |`false`|
|includeLineNum|Boolean| Determines if the line number needs to be printed in logs. This config takes into effect only if `includeFunction=true`|`false`|
|pkgConfigs   |Array|This field consists array of package specific configuration.<br>`{"pkgName": "<packageName>","level": "<Level>"}`|`null`|
|writers| Array|Array of writers either `file` or `console` based writer. Console Writer Has the following has the following configuration `{"console": {"errToStdOut": false,"warnToStdOut": false}}`.Log levels except `ERROR and WARN` are written to `os.Stdout`. The Entries for the `ERROR,WARN` can be written to either os.StdErr or os.Stdout <br> For a file based log destination,paths for each level can be specified as follows.<br> `{"file": { "defaultPath": "<file Path>","errorPath": "<file Path>","warnPath": "<file Path>","infoPath": "<file Path>","debugPath": "<file Path>","tracePath": "<file Path>" }`<br> If any of the `errorPath,warnPath,infoPath,debugPath,tracePath` is not specified then default path for that level is applied. If all of the level specific paths are specified then the `defaultPath` value is ignored.| N/A|

By Default the logging framework looks for a file named `log-config.json` in the same directory if the application.
This is the default location.This location can be overridden using an environment variable `GC_LOG_CONFIG_FILE`.
If the framework cannot resolve the configuration file either in the default location or at the location specified by
environment  variable, then the framework loads a default configuration as described below.


### 2. Default Log Config- With ENV variables override
The default log configuration will write the log entries to the console and the framework default log  level is  `INFO`.
Few fields can be overwritten using environment variables. The following table shows those env variables
|Environment Var | Converted Type  | Description   | Default Value|
|:-|:-|:-|:-:|
|GC_LOG_ASYNC|Boolean|Determines if the message is to be written to the destination asynchronously. If set to `true` then the LogMessage is prepared synchronously.However, it is written to destination in a async fashion.|`false`|
|GC_LOG_FMT|String| The output format of the log message. The valid values are `text` or `json`| `text` |
|GC_LOG_DEF_LEVEL| String|Sets the default Logging level for the Logger. This is a global value. For overriding the log levels for a specific packages use `pkgConfigs`. The valid values are `OFF,ERROR,INFO,DEBUG,TRACE`| `INFO`|
|GC_LOG_TIME_FMT|String| The timestamp format for the log entries.Valid values are the ones acceptable by`time.Format(<pattern>)` function.|As defined by `time.RFC3339`|
|GC_LOG_WARN_STDOUT|Boolean|Indicates whether the log entries with `WARN` level needs to be written to `os.Stdout` instead of `os.Stderr`|`false`|
|GC_LOG_ERR_STDOUT|Boolean|Indicates whether the log entries with `ERROR` level needs to be written to `os.Stdout` instead of `os.Stderr`|`false`|

### 3. Configuring logging during runtime.

The log configuration can be set at runtime using the `l3.Configure(l *LogConfig)`. This is a global level log configuration and will impact all instances of logger
