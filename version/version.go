package version

var (
	CurrentCommit string

	BuildVersion = "1.13.0-rc1"

	Version = BuildVersion + CurrentCommit
)
