package client

import "log"

var Debug bool
var Verbose bool

func LogDebug(format string, v ...interface{}) {
	if Debug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func LogVerbose(format string, v ...interface{}) {
	if Verbose {
		log.Printf("[VERBOSE] "+format, v...)
	}
}
