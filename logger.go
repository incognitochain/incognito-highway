package main

import (
	"highway/chain"
	"highway/process"
	"highway/process/topic"
	"highway/route"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

func initLogger() {
	cf := zap.NewDevelopmentConfig()
	cf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	l, _ := cf.Build()
	logger = l.Sugar()

	// Initialize children's loggers
	chain.InitLogger(logger)
	route.InitLogger(logger)
	process.InitLogger(logger)
	topic.InitLogger(logger)
}
