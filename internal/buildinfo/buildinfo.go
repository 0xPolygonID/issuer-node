package buildinfo

import (
	"runtime/debug"
)

const length = 7

// Revision returns the revision of the current build.
func Revision() (rev string) {
	rev = get("vcs.revision")
	if len(rev) > length {
		rev = rev[:length]
	}
	return
}

func get(key string) string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == key {
				return setting.Value
			}
		}
	}
	return ""
}
