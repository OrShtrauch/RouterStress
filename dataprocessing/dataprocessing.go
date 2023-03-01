// Package responsible for processing the raw test results to a JSON file and a graph.
package dataprocessing

import (
	"RouterStress/consts"
	"RouterStress/log"
	"RouterStress/stress"
	"RouterStress/errors"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"golang.org/x/sync/errgroup"
)

const (
	SCENARIOS = "scenarios"
)

func Run(stress *stress.Stress, runIndex int) error {
	var eg errgroup.Group

	eg.Go(func() error {
		var err error
		_, sampleFiles, err := GetAllCsvFiles(runIndex)

		if err != nil {
			return err
		}

		if len(sampleFiles) == 0 {
			return errors.NoFilesFound{}
		}

		metrics, err := GetTestMetrics(sampleFiles, runIndex)

		if err != nil {
			return err
		}
		log.Logger.Sugar().Debugf("metrics: %v", *metrics)
		return saveJson(*metrics)
	})

	eg.Go(func() error {
		c, err := stress.Docker.StartPlotterContainer()

		if err != nil {
			return err
		}

		return stress.Docker.WaitForContainerToDie(c)
	})

	return eg.Wait()
}

func saveJson(metrics TestMetrics) error {
	dir, err := getTestDir()

	if err != nil {
		return err
	}

	path := fmt.Sprintf("%v/%v", dir, "results.json")
	f, err := os.Create(path)

	if err != nil {
		return err
	}

	defer f.Close()

	bytes, err := json.Marshal(metrics)

	if err != nil {
		return err
	}

	_, err = f.WriteString(string(bytes))

	return err
}

func GetTestMetrics(files []string, runIndex int) (*TestMetrics, error) {
	var err error
	testTotalMetrics := &TestMetrics{
		Total:     &Metrics{},
		Scenarios: make([]*Scenario, 0),
	}

	testTotalMetrics.Total.RunIndex = &runIndex
	testTotalMetrics.Total.Clients = new(int)
	testTotalMetrics.Total.Requests = new(int)

	dataMap, err := getDataMap(files)

	if err != nil {
		return testTotalMetrics, err
	}

	for scenarioName, tuples := range dataMap {
		scenario := &Scenario{
			Name:    scenarioName,
			Total:   &Metrics{},
			Minutes: make(map[string]*Metrics),
		}

		scenario.Total.Clients = &tuples.Clients
		*testTotalMetrics.Total.Clients += tuples.Clients

		for _, row := range tuples.DataRows {
			// apending data to total test metrics
			appendDataToMetrics(testTotalMetrics.Total, row)
			// apending data to scenario metrics
			appendDataToMetrics(scenario.Total, row)

			// apending data to minute-per-scenaio metrics
			if err = appendDataToMinuteMetrics(scenario, row); err != nil {
				return testTotalMetrics, err
			}
		}

		testTotalMetrics.Scenarios = append(testTotalMetrics.Scenarios, scenario)
	}

	CalcAvgAndErrorRate(testTotalMetrics)

	return testTotalMetrics, err
}

func CalcAvgAndErrorRate(testTotalMetrics *TestMetrics) {
	if testTotalMetrics.Total.Requests == nil {
		log.Logger.Debug("req is nil")
	}
	if testTotalMetrics.Total.AvgElapsed == nil {
		log.Logger.Debug("avg is nil")
	}
	*testTotalMetrics.Total.AvgElapsed /= float32(*testTotalMetrics.Total.Requests)
	*testTotalMetrics.Total.ErrorRate /= float32(*testTotalMetrics.Total.Requests)

	for _, scenario := range testTotalMetrics.Scenarios {
		*scenario.Total.AvgElapsed /= float32(*scenario.Total.Requests)
		*scenario.Total.ErrorRate /= float32(*scenario.Total.Requests)

		for _, minute := range scenario.Minutes {
			*minute.AvgElapsed /= float32(*minute.Requests)
			*minute.ErrorRate /= float32(*minute.Requests)
		}
	}
}

func appendDataToMinuteMetrics(scenario *Scenario, row DataRow) error {
	var err error

	// apending data to minute-per-scenaio metrics
	minute, err := parseTimestamp(row.Timestamp, true)

	if err != nil {
		return err
	}

	minute_str := minute.Format(consts.DT_LAYOUT)

	if scenario.Minutes[minute_str] == nil {
		scenario.Minutes[minute_str] = &Metrics{}
	}

	appendDataToMetrics(scenario.Minutes[minute_str], row)

	return err
}

func appendDataToMetrics(metrics *Metrics, row DataRow) {
	if metrics.MinElapsed == nil {
		metrics.MinElapsed = &row.Elasped
	} else {
		if row.Elasped < *metrics.MinElapsed {
			*metrics.MinElapsed = row.Elasped
		}
	}

	if metrics.MaxElapsed == nil {
		metrics.MaxElapsed = &row.Elasped
	} else {
		if row.Elasped > *metrics.MaxElapsed {
			*metrics.MaxElapsed = row.Elasped
		}
	}

	if metrics.AvgElapsed == nil {
		metrics.AvgElapsed = new(float32)
	}

	*metrics.AvgElapsed += row.Elasped

	if metrics.Requests == nil {
		metrics.Requests = new(int)
	}
	*metrics.Requests += 1

	if metrics.ErrorRate == nil {
		metrics.ErrorRate = new(float32)
	}

	if row.Status == 1 {
		*metrics.ErrorRate += 1
	}
}

func getDataMap(files []string) (map[string]*Tuple, error) {
	dataMap := make(map[string]*Tuple)
	var err error

	filesMap := getFileMap(files)

	for scenario, files := range filesMap {
		dataRows := make([]DataRow, 0)

		for _, file := range files {
			rows, err := fileToStruct[[]DataRow](file)

			if err != nil {
				return dataMap, err
			}

			dataRows = append(dataRows, rows...)
		}

		dataMap[scenario] = &Tuple{
			DataRows: dataRows,
			Clients:  len(files),
		}
	}

	return dataMap, err

}

func parseTimestamp(datetime string, zeroSeconds bool) (time.Time, error) {
	var date time.Time
	var err error

	date, err = time.Parse(consts.DT_LAYOUT, datetime)

	if err != nil {
		return date, err
	}

	if zeroSeconds {
		// changing the seconds to 0
		date = time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, date.Location())
	}

	return date, err
}

func fileToStruct[T any](filename string) (T, error) {
	var t T
	bytes, err := os.ReadFile(filename)

	if err != nil {
		return t, err
	}

	gocsv.UnmarshalBytes(bytes, &t)
	return t, err
}

func getTestDir() (string, error) {
	var path string
	var err error

	workingDir, err := os.Getwd()

	if err != nil {
		return path, err
	}

	tesdID := consts.TEST_ID //"Edison_Puma5_e78a4eb7-1610-4aa3-ba06-59f5cabf16f2" //consts.TEST_ID
	return fmt.Sprintf("%v/%v%v", workingDir, consts.RESULTS_DIR, tesdID), err
}

func GetAllCsvFiles(runIndex int) (string, []string, error) {
	var routerFile string
	var err error

	csvFiles := make([]string, 0)

	path, err := getTestDir()

	if err != nil {
		return routerFile, csvFiles, err
	}

	files, err := os.ReadDir(path)

	if err != nil {
		return routerFile, csvFiles, err
	}

	for _, file := range files {
		name := file.Name()
		filePath := fmt.Sprintf("%v/%v", path, name)

		if strings.HasSuffix(name, ".csv") {
			if strings.Contains(name, "router") {
				routerFile = filePath
			} else {
				if inRun, err := isFileInRun(name, runIndex); err == nil {
					if inRun {
						csvFiles = append(csvFiles, filePath)
					}
				} else {
					return routerFile, csvFiles, err
				}
			}
		}
	}

	return routerFile, csvFiles, err
}

func isFileInRun(fileName string, runIndex int) (bool, error) {
	var err error

	fileWOSuffix := strings.Replace(fileName, ".csv", "", -1)

	splitName := strings.Split(fileWOSuffix, "_")
	fileRI, err := strconv.Atoi(splitName[3])

	if err != nil {
		return false, err
	}

	return fileRI == runIndex, err
}

func getFileMap(files []string) map[string][]string {
	modes := make(map[string][]string)

	for _, file := range files {
		split_name := strings.Split(file, "_")
		mode := split_name[len(split_name)-5]

		if modes[mode] == nil {
			modes[mode] = make([]string, 0)
		}

		modes[mode] = append(modes[mode], file)

	}

	return modes
}
