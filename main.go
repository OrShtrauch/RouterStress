package main

import (
	"RouterStress/conf"
	"RouterStress/log"
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


	// fmt.Println(ip)
	// fmt.Println(len(server.Used))

	// fmt.Printf("\n\n\n")
	// server.Release(ip)
	// fmt.Println(len(server.Used))
	// var wg sync.WaitGroup
	// for i := 1; i < 5; i++ {
	// 	wg.Add(1)

	// 	i := i
	// 	go func() {
	// 		defer wg.Done()
	// 		worker(i)
	// 	}()

	// }

	// wg.Wait()

	// fmt.Printf("done\n")

	// x := 5

	// y := x + 5

	// fmt.Println(y)
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
