package main

type Configuration struct {
	Storage   storageConfig
	RateLimit rateLimitConfig
	Filter    filterConfig
	Security  securityConfig
	Server    serverConfig
	Redis     redisConfig
	MoreStats bool
}

type storageConfig struct {
	Directory              string
	MaxSize                int64
	IDLen                  int64
	CollisionCheckAttempts int64
}

type rateLimitConfig struct {
	ResetAfter int64
	Upload     int64
	Global     int64
	Bandwidth  rateLimitBandwidthConfig
}

type rateLimitBandwidthConfig struct {
	ResetAfter int64
	Download   int64
	Upload     int64
}

type filterConfig struct {
	Blacklist []string
	Whitelist []string
	Sanitize  []string
}

type securityConfig struct {
	MasterKey string
}

type serverConfig struct {
	Port        int64
	Concurrency int64
}

type redisConfig struct {
	URI      string
	Password string
	Db       int64
}
