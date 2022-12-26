package main

import (
	"RouterStress/conf"
	"RouterStress/log"
	"RouterStress/stress"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	config, err := conf.GetConfig()

	if err != nil {
		panic(err)
	}

	InitLogger(&config)


	stress, err := stress.NewStress(&config)

	if err != nil {
		log.Logger.Error(err.Error())
		panic(err)
	}

	err = stress.Start()

	if err != nil {
		log.Logger.Error(err.Error())
		stress.Cleanup()
		panic(err)
	}

	stress.Cleanup()
}

func InitLogger(config *conf.Config) {
	logLvl := zapcore.InfoLevel

	if config.Settings.Debug {
		logLvl = zapcore.DebugLevel
	}

	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewConsoleEncoder(zapConfig)
	logFile, _ := os.OpenFile("stress.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, logLvl),
	)

	log.Logger = zap.New(core, zap.AddCaller())
}
