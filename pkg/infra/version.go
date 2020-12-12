package infra

import (
	"fmt"
	"runtime"
)

const (
	programName = "tape"
	version     = "0.0.2"
)

// GetVersionInfo return version information
// TODO add commit hash, Built info
func GetVersionInfo() string {
	return fmt.Sprintf(
		"%s:\n Version: %s\n Go version: %s\n OS/Arch: %s\n",
		programName,
		version,
		runtime.Version(),
		fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	)
}
