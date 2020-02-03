package datahandler

import (
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger // General logger package chain

func InitLogger(baseLogger *zap.SugaredLogger) {
	// Init package's logger here with distinct name here
	logger = baseLogger.Named("datahandler")
}
