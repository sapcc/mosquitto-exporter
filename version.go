package main

import (
	"fmt"
	"runtime"
)

var (
	Version   = "20170214.01"
	GITCOMMIT = "HEAD"
)

func versionString() string {
	return fmt.Sprintf("%s (%s), %s", Version, GITCOMMIT, runtime.Version())
}
