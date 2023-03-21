package main

import (
	"fmt"
	"log"
)

func logDebugf(format string, v ...interface{}) {
	log.Printf(logPrefixFormat("DEBUG", format), v...)
}

func logErrorf(format string, v ...interface{}) {
	log.Printf(logPrefixFormat("ERROR", format), v...)
}

func logFatalf(format string, v ...interface{}) {
	log.Fatalf(logPrefixFormat("FATAL", format), v...)
}

func logInfof(format string, v ...interface{}) {
	log.Printf(logPrefixFormat("INFO", format), v...)
}

func logWarningf(format string, v ...interface{}) {
	log.Printf(logPrefixFormat("WARNING", format), v...)
}

func logPrefixFormat(level string, format string) string {
	return fmt.Sprintf("%-7s | %s", level, format)
}
