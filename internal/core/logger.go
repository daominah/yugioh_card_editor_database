package core

import (
	"fmt"
	"log"
	"time"
)

var VietnamTimezone *time.Location

// init initializes the standard logger with millisecond precision.
// It sets log flags to include short file name and microseconds, and configures
// a custom output writer that adds Hanoi timezone timestamp with millisecond precision to each log line.
// This runs automatically when the core package is imported.
func init() {
	log.SetFlags(log.Lshortfile) // time format will be defined in customLogger
	vnTimezone, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		// Fallback to UTC if timezone loading fails
		vnTimezone = time.UTC
	}
	vnLogger := customLogger{timezone: vnTimezone}
	log.SetOutput(vnLogger)

	VietnamTimezone = vnTimezone
}

// customLogger adds time to the beginning of each log line, write to stdout
type customLogger struct {
	timezone *time.Location
}

func (writer customLogger) Write(bytes []byte) (int, error) {
	return fmt.Printf("%v %s", time.Now().In(writer.timezone).Format("2006-01-02T15:04:05.000Z07:00"), bytes)
}
