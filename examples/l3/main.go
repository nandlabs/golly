// Package main demonstrates the l3 (multi-level logging) package.
package main

import (
	"oss.nandlabs.io/golly/l3"
)

func main() {
	// Get the default logger
	logger := l3.Get()

	// Log at different levels
	logger.Info("Application started")
	logger.InfoF("Server listening on port %d", 8080)

	logger.Debug("Debug information: initializing subsystems")
	logger.DebugF("Loading config from %s", "/etc/app/config.yaml")

	logger.Warn("Disk usage above 80%")
	logger.WarnF("Connection pool at %d%% capacity", 85)

	logger.Trace("Entering function main()")
	logger.TraceF("Processing item %d of %d", 42, 100)

	// Error level logging
	logger.Error("Failed to connect to database")
	logger.ErrorF("Request failed with status %d: %s", 500, "Internal Server Error")

	// Custom configuration
	cfg := &l3.LogConfig{
		PkgConfigs: []*l3.PackageConfig{
			{
				PackageName: "main",
				Level:       "debug",
			},
		},
	}
	l3.Configure(cfg)

	// After configuration, log again
	logger = l3.Get()
	logger.Info("Logger reconfigured")
}
