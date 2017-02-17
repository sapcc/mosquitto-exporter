package main

import (
	"fmt"
	"runtime"
)

var (
	Version   = "develop"
	GITCOMMIT = "HEAD"
)

func versionString() string {
	return fmt.Sprintf("%s (%s), %s", Version, GITCOMMIT, runtime.Version())
}
