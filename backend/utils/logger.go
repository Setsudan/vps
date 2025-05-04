package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// GetLogger returns a configured logrus logger.
func GetLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info"
	}
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	return logger
}