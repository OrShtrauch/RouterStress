package main

import (
	// "RouterStress/conf"
	// "RouterStress/consts"
	// "RouterStress/log"
	// "RouterStress/stress"
	// "os/signal"
	// "os"
	// "fmt"

	// "go.uber.org/zap"
	// "go.uber.org/zap/zapcore"
	//"github.com/go-gota/gota/dataframe"
	"RouterStress/dataprocessing"
	"encoding/json"
	"fmt"
)

func main() {
	var err error
	_, sampleFiles, err := dataprocessing.GetAllCsvFiles(0)

	if err != nil {
		panic(err)
	}

	// err = dataprocessing.PlotRouterCPUData(routerFile)

	// if err != nil {
	// 	panic(err)
	// }

	metrics, err := dataprocessing.GetTestMetrics(sampleFiles, 0)

	if err != nil {
		panic(err)
	}

	//fmt.Println(metrics)

	bytes, err := json.Marshal(metrics)

	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))


	// dataDFs, err := dataprocessing.GetStressDataDataFrames(sampleFiles)
	// if err != nil {
	// 	panic(err)
	// }
	
	// routerDF, err := dataprocessing.GetRouterDF(dataDFs[0], routerFile)

	// if err != nil {
	// 	panic(err)
	// }		
		
	// fmt.Printf("router: %v\n", routerDF)
	// fmt.Println(dataDFs[0])

}

// func main() {
// 	config, err := conf.GetConfig()

// 	if err != nil {
// 		panic(err)
// 	}

// 	InitLogger(&config)

// 	stress, err := stress.NewStress(&config)

// 	if err != nil {
// 		log.Logger.Error(err.Error())
// 		panic(err)
// 	}

// 	channel := make(chan os.Signal)

// 	signal.Notify(channel, os.Interrupt)

// 	go func() {
// 		sig := <-channel 
// 		log.Logger.Info(fmt.Sprintf("Received %s\n", sig))

// 		stress.Cleanup()

// 		if err!= nil {
// 			fmt.Println(err.Error())
// 		}


// 		log.Logger.Debug(fmt.Sprintf("TestID: %v\n", consts.TEST_ID))
		
// 		close(channel)
// 		os.Exit(0)		
// 	}()

// 	log.Logger.Debug("finished setup")

// 	err = stress.Start()

// 	if err != nil {
// 		log.Logger.Error(err.Error())
// 		stress.Cleanup()
// 		panic(err)
// 	}

// 	stress.Cleanup()

// 	log.Logger.Debug("Done.")
// 	log.Logger.Debug(fmt.Sprintf("TestID: %v\n", consts.TEST_ID))
// }

// func InitLogger(config *conf.Config) {
// 	logLvl := zapcore.InfoLevel

// 	if config.Settings.Debug {
// 		logLvl = zapcore.DebugLevel
// 	}

// 	zapConfig := zap.NewProductionEncoderConfig()
// 	zapConfig.EncodeTime = zapcore.ISO8601TimeEncoder
// 	fileEncoder := zapcore.NewConsoleEncoder(zapConfig)
// 	logFile, _ := os.OpenFile("stress.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	writer := zapcore.AddSync(logFile)

// 	core := zapcore.NewTee(
// 		zapcore.NewCore(fileEncoder, writer, logLvl),
// 	)

// 	log.Logger = zap.New(core, zap.AddCaller())
// }
