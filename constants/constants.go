package constants

const (
	// Version is the current version of the server.
	Version = "1.3.4"

	// ExtensionLengthLimit means that file extensions cannot have more characters than the number specified.
	// file.png has an extension of 4.
	ExtensionLengthLimit = 12
)

// PathType is an integer representation of what path is currently being handled.
// Used mainly by constants.LimitPath.
type PathType int

const (
	// LimitUploadPath represents /upload.
	LimitUploadPath PathType = iota

	// LimitGeneralPath represents all paths which aren't handled individually (like LimitUploadPath).
	LimitGeneralPath
)

const (
	RateLimitBandwidthDownload = "bw_dn_"
	RateLimitBandwidthUpload = "bw_up_"
)

const (
	RequestMaxBodySizePadding = 2048
)

// sus imposter
var PathLengthLimitBytes int
