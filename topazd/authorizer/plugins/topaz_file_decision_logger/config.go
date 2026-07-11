package topaz_file_decision_logger

const (
	default_decision_log_filename string = "decisions.json"
	defaultFilename               string = ""    // default <processname>-lumberjack.log in os.TempDir().
	defaultMaxSize                int    = 50    // default 100 megabytes.
	defaultMaxAge                 int    = 2     // default is not to remove old log files based on age.
	defaultMaxBackups             int    = 2     // default is to retain all old log files (though MaxAge may still cause them to get deleted.).
	defaultLocalTime              bool   = false // default is to use UTC time.
	defaultCompress               bool   = false // default is not to perform compression.
)

type Config struct {
	Enabled    bool       `json:"enabled"`
	Logger     Logger     `json:"logger"`
	PolicyInfo PolicyInfo `json:"policy_info"`
}

type Logger struct {
	Filename   string `json:"filename"`
	MaxSize    int    `json:"max_size"`
	MaxAge     int    `json:"max_age"`
	MaxBackups int    `json:"max_backups"`
	LocalTime  bool   `json:"local_time"`
	Compress   bool   `json:"compress"`
}

type PolicyInfo struct {
	PolicyName      string `json:"policy_name"`      //
	RegistryService string `json:"registry_service"` //
	RegistryImage   string `json:"registry_image"`   //
	RegistryTag     string `json:"registry_tag"`     //
	Digest          string `json:"digest"`           //
}

func defaultConfig() *Config {
	return &Config{
		Enabled: false,
		Logger: Logger{
			Filename:   defaultFilename,
			MaxSize:    defaultMaxSize,
			MaxBackups: defaultMaxBackups,
			MaxAge:     defaultMaxAge,
			LocalTime:  defaultLocalTime,
			Compress:   defaultCompress,
		},
		PolicyInfo: PolicyInfo{
			PolicyName:      "",
			RegistryService: "",
			RegistryImage:   "",
			RegistryTag:     "",
			Digest:          "",
		},
	}
}

func DefaultDecisionLogFilename() string {
	return default_decision_log_filename
}
