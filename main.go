package main

import (
	"RouterStress/conf"
	"RouterStress/consts"
	"RouterStress/dataprocessing"
	"RouterStress/log"
	"syscall"

	"RouterStress/s3"
	"RouterStress/stress"
	"os"
	"os/signal"

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

	channel := make(chan os.Signal, 1)

	signal.Notify(channel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-channel

		log.Logger.Infof("Received %s\n", sig)

		cleanup(stress)

		close(channel)
		os.Exit(0)
	}()

	log.Logger.Debug("finished setup")

	err = stress.Start()

	if err != nil {
		log.Logger.Error(err.Error())
		stress.Cleanup()
		panic(err)
	}

	cleanup(stress)
}

func cleanup(stress stress.Stress) {
	if err := stress.Cleanup(); err != nil {
		log.Logger.Error(err.Error())
		panic(err)
	}

	log.Logger.Debug("Proccessing test Data")

	if err := dataprocessing.Run(&stress, consts.RUN_INDEX); err != nil {
		log.Logger.Error(err.Error())
		panic(err)
	}

	log.Logger.Info("Uploading to S3")

	if err := s3.Upload(); err != nil {
		panic(err)
	}

	log.Logger.Debug("Done.")
	log.Logger.Debugf("TestID: %v\n", consts.TEST_ID)
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

	log.Logger = zap.New(core, zap.AddCaller()).Sugar()
}
