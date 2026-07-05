package authzen_logger

const (
	defaultFilename   string = ""    // default <processname>-lumberjack.log in os.TempDir().
	defaultMaxSize    int    = 100   // default 100 megabytes.
	defaultMaxAge     int    = 0     // default is not to remove old log files based on age.
	defaultMaxBackups int    = 0     // default is to retain all old log files (though MaxAge may still cause them to get deleted.).
	defaultLocalTime  bool   = false // default is to use UTC time.
	defaultCompress   bool   = false // default is not to perform compression.
)

type Config struct {
	Enabled bool    `json:"enabled"`
	Logger  *Logger `json:"logger"`
}

type Logger struct {
	Filename   string `json:"filename"`
	MaxSize    int    `json:"max_size"`
	MaxAge     int    `json:"max_age"`
	MaxBackups int    `json:"max_backups"`
	LocalTime  bool   `json:"local_time"`
	Compress   bool   `json:"compress"`
}

func defaultConfig() *Config {
	return &Config{
		Enabled: false,
		Logger: &Logger{
			Filename:   defaultFilename,
			MaxSize:    defaultMaxSize,
			MaxBackups: defaultMaxBackups,
			MaxAge:     defaultMaxAge,
			LocalTime:  defaultLocalTime,
			Compress:   defaultCompress,
		},
	}
}
