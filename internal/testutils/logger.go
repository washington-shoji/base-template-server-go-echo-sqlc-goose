package testutils

import (
	"go-echo-server-template/internal/logger"
)

func init() {
	// Initialize logger for tests, using "DEBUG" as the default log level for tests.
	if err := logger.Initialize("test", "DEBUG"); err != nil {
		panic("Failed to initialize logger for tests: " + err.Error())
	}
}
