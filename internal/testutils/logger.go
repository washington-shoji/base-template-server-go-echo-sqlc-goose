package testutils

import (
	"go-echo-server-template/internal/logger"
)

func init() {
	// Initialize logger for tests
	if err := logger.Initialize("test"); err != nil {
		panic("Failed to initialize logger for tests: " + err.Error())
	}
}
