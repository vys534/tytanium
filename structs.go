package main

type Configuration struct {
	Security SecurityConfig
	Server   ServerConfig
	Net      NetConfig
	Storage  StorageConfig
}

type StorageConfig struct {
	Directory string
}

type SecurityConfig struct {
	MasterKey      string
	MaxSizeBytes   int
	RateLimit      RateLimitConfig
	BandwidthLimit BandwidthLimitConfig
	Filter         FilterConfig
}

type BandwidthLimitConfig struct {
	ResetAfter int64
	Download   int64
	Upload     int64
}

type RateLimitConfig struct {
	ResetAfter int64
	Upload     int64
	Global     int64
}

type FilterConfig struct {
	Blacklist []string
	Whitelist []string
	Sanitize  []string
}

type ServerConfig struct {
	Port                   string
	Concurrency            int
	IDLen                  int
	CollisionCheckAttempts int
}

type NetConfig struct {
	Redis RedisConfig
}

type RedisConfig struct {
	URI      string
	Password string
	Db       int
}
