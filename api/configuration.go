package api

// Configuration is the configuration structure used by the program.
type Configuration struct {
	Storage                 storageConfig
	RateLimit               rateLimitConfig
	Filter                  filterConfig
	Security                securityConfig
	Server                  serverConfig
	Redis                   redisConfig
	MoreStats               bool
	ForceZeroWidth          bool
	StatsCollectionInterval int
	Logging                 loggingConfig
}

type loggingConfig struct {
	Enabled bool
	LogFile string
}

type storageConfig struct {
	Directory              string
	MaxSize                int
	IDLength               int
	CollisionCheckAttempts int
}

type rateLimitConfig struct {
	ResetAfter int
	Path       struct {
		Upload int
		Global int
	}
	Bandwidth rateLimitBandwidthConfig
}

type rateLimitBandwidthConfig struct {
	ResetAfter int
	Download   int
	Upload     int
}

type filterConfig struct {
	Blacklist []string
	Whitelist []string
	Sanitize  []string
}

type securityConfig struct {
	MasterKey                    string
	DisableEmptyMasterKeyWarning bool
}

type serverConfig struct {
	Port         int
	Concurrency  int
	ReadTimeout  int
	WriteTimeout int
}

type redisConfig struct {
	URI      string
	Password string
	DB       int
}
