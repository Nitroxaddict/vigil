package meta

const Name = "Vigil"

var (
	// Version is the compile-time set version
	Version = "v0.0.0-unknown"

	// UserAgent is the http client identifier derived from Version
	UserAgent string
)

func init() {
	UserAgent = Name + "/" + Version
}
