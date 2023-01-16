package dataprocessing

import (
	"RouterStress/consts"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	//"time"

	//"github.com/go-gota/gota/dataframe"
)

func getAllCsvFiles(runIndex int) (string, []string, error) {
	var routerFile string
	var csvFiles []string
    var err error

	csvFiles = make([]string, 0)

	workingDir, err := os.Getwd()

	if err != nil {
		return routerFile, csvFiles, err
	}

	localPath := fmt.Sprintf("%v/%v/%v", workingDir, consts.RESULTS_DIR, consts.TEST_ID)

	files, err := ioutil.ReadDir(localPath)

	if err!= nil {
        return routerFile, csvFiles, err
	}

	for _, file := range files {
		name := file.Name()

		if strings.HasSuffix(name, ".csv") {
			// if isFileInRun(name, runIndex) {				
			// 	if strings.Contains(name, "router") {
			// 		routerFile = name
			// 	} else {
			// 		csvFiles = append(csvFiles, name)
			// 	}
			// }		
		}
	}

	return routerFile, csvFiles, err
}

func isFileInRun(fileName string, runIndex int) bool {
	fileWOSuffix := strings.Replace(fileName, ".csv", "", -1)

	splitName := strings.Split(fileWOSuffix, "_")

	runIndexStr := splitName[len(splitName) - 1]

	return runIndexStr == fmt.Sprintln(runIndex)
}

// func TrimRouterDF(routerDF dataframe.DataFrame, sampleDF dataframe.DataFrame) dataframe.DataFrame {

// 	//startTime, err := time.Parse(consts.DT_FORMAT, sampleDF.Select([]string{"timestamp"}).Subset([]int{0}))	
// }	