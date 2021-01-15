package main

type Configuration struct {
	Security SecurityConfig
	Server   ServerConfig
	Net      NetConfig
}

type SecurityConfig struct {
	MasterKey      string
	MaxSizeBytes   int
	RateLimit      RateLimitConfig
	BandwidthLimit BandwidthLimitConfig
	PublicMode     bool
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
	Port          string
	Concurrency   int
	MaxConnsPerIP int
	IDLen         int
}

type NetConfig struct {
	Redis RedisConfig
	GCS   GCSConfig
}

type RedisConfig struct {
	URI      string
	Password string
	Db       int
}

type GCSConfig struct {
	BucketName string
	SecretKey  string
}
