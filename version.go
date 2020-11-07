package main

import (
	"fmt"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func versionString() string {
	return fmt.Sprintf("mosquitto exporter %s, commit %s, built at %s by %s", version, commit, date, builtBy)
}
