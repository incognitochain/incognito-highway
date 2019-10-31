package customizelog

//variant version of rlog https://github.com/romana/rlog
const (
	notATrace     = -1
	noTraceOutput = -1
)

// The known log levels
const (
	levelNone = iota
	levelCrit
	levelErr
	levelWarn
	levelInfo
	levelDebug
	levelTrace
)

// Translation map from level to string representation
var levelStrings = map[int]string{
	levelTrace: "TRACE",
	levelDebug: "DEBUG",
	levelInfo:  "INFO",
	levelWarn:  "WARN",
	levelErr:   "ERROR",
	levelCrit:  "CRITICAL",
	levelNone:  "NONE",
}

// Translation from level string to number.
var levelNumbers = map[string]int{
	"TRACE":    levelTrace,
	"DEBUG":    levelDebug,
	"INFO":     levelInfo,
	"WARN":     levelWarn,
	"ERROR":    levelErr,
	"CRITICAL": levelCrit,
	"NONE":     levelNone,
}

// filterSpec holds a list of filters. These are applied to the 'caller'
// information of a log message (calling module and file) to see if this
// message should be logged. Different log or trace levels per file can
// therefore be maintained. For log messages this is the log level, for trace
// messages this is going to be the trace level.
type filterSpec struct {
	filters []filter
}

// filter holds filename and level to match logs against log messages.
type filter struct {
	Pattern string
	Level   int
}

// LogConfig captures the entire configuration of log, as supplied by a user
// via environment variables and/or config files. This still requires checking
// and translation into more easily used config items. All values therefore are
// stored as simple strings here.
type LogConfig struct {
	LogLevel        string `json:"loglvl"`     // What log level. String, since filters are allowed
	TraceLevel      string `json:"tracelvl"`   // What trace level. String, since filters are allowed
	LogTimeFormat   string `json:"timeformat"` // The time format spec for date/time stamps in output
	LogFile         string `json:"logfile"`    // Name of logfile
	ConfFile        string `json:"configfile"` // Name of config file
	LogStream       string // Name of logstream: stdout, stderr or NONE
	LogNoTime       string // Flag to determine if date/time is logged at all
	ShowCallerInfo  string // Flag to determine if caller info is logged
	ShowGoroutineID string // Flag to determine if goroute ID shows in caller info
	ConfCheckInterv string // Interval in seconds for checking config file
}
