package dataprocessing

import (
	"RouterStress/consts"
	"RouterStress/log"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
)

const (
	SCENARIOS = "scenarios"
)

func Run(runIndex int) error {
	var err error
	routerFile, sampleFiles, err := GetAllCsvFiles(runIndex)

	if err != nil {
		return err
	}

	metrics, err := GetTestMetrics(sampleFiles, runIndex)

	if err != nil {
		return err
	}

	err = saveJson(*metrics)

	if err != nil {
		return err
	}

	dataRow, err := fileToStruct[[]DataRow](sampleFiles[0])

	if err != nil {
		return err
	}

	return PlotRouterCPUData(routerFile, dataRow[0])
}

// func GetStressDataDataFrames(sampleFiles []string) ([]dataframe.DataFrame, error) {
// 	var dfs []dataframe.DataFrame
// 	var err error

// 	for _, file := range sampleFiles {
// 		df, err := GetStressDataDateFrame(file)

// 		if err != nil {
// 			return dfs, err
// 		}

// 		dfs = append(dfs, df)
// 	}

// 	return dfs, err
// }

// func GetStressDataDateFrame(file string) (dataframe.DataFrame, error) {
// 	var df dataframe.DataFrame
// 	var err error

// 	text, err := os.ReadFile(file)
// 	if err != nil {
// 		return df, err
// 	}

// 	df = dataframe.ReadCSV(strings.NewReader(string(text)))

// 	return df, err
// }

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
			metrics.MinElapsed = &row.Elasped
		}
	}

	if metrics.MaxElapsed == nil {
		metrics.MaxElapsed = &row.Elasped
	} else {
		if row.Elasped > *metrics.MaxElapsed {
			metrics.MaxElapsed = &row.Elasped
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
	fileRI, err := strconv.Atoi(splitName[2])

	if err != nil {
		return false, err
	}

	return fileRI == runIndex, err
}

func getFileMap(files []string) map[string][]string {
	modes := make(map[string][]string)

	for _, file := range files {
		split_name := strings.Split(file, "_")
		mode := split_name[len(split_name)-4]

		if modes[mode] == nil {
			modes[mode] = make([]string, 0)
		}

		modes[mode] = append(modes[mode], file)

	}

	return modes
}

func PlotRouterCPUData(routerFile string, sampleDataRow DataRow) error {
	var err error

	p := plot.New()

	p.Title.Text = "CPU Usage Over Time(%)"
	p.X.Label.Text = "Time"
	p.Y.Label.Text = "Usage"

	routerXYS, err := getRouterXYs(routerFile, sampleDataRow)

	if err != nil {
		return err
	}

	line, err := plotter.NewLine(routerXYS)

	if err != nil {
		return err
	}

	p.X.Tick.Marker = plot.TimeTicks{Format: consts.PLOT_DT_LAYOUT}
	p.Y.Label.Text = "CPU(%)"
	p.X.Label.Padding = 10
	p.X.Label.Text = "Time"

	p.Add(line)

	p.Title.Padding = 15
	p.X.Padding = 10
	p.Y.Padding = 10

	width := (10*routerXYS.Len() + 150)

	testDir, err := getTestDir()

	if err != nil {
		return err
	}

	localPath := fmt.Sprintf("%v/%v", testDir, "cpu_usage.png")

	err = p.Save(font.Length(width), 500, localPath)

	return err
}

func getRouterXYs(routerFile string, sampleDataRow DataRow) (plotter.XYs, error) {
	var xys plotter.XYs
	var err error
	routerData, err := fileToStruct[[]RouterDataRow](routerFile)
	if err != nil {
		return xys, err
	}

	err = AdjustRouterData(&routerData, sampleDataRow)

	if err != nil {
		return xys, err
	}

	xys = make(plotter.XYs, len(routerData))

	for i, row := range routerData {
		timestamp, err := parseTimestamp(row.Timestamp, false)

		if err != nil {
			return xys, err
		}

		xys[i].X = float64(timestamp.Unix())
		xys[i].Y = float64(row.Cpu)
	}
	log.Logger.Debug(fmt.Sprintf("xys: %v\n", xys))
	return xys, err
}

func AdjustRouterData(routerData *[]RouterDataRow, sampleDataRow DataRow) error {
	var err error

	if len(*routerData) > 0 {
		return err
	}

	sampleDT, err := parseTimestamp(sampleDataRow.Timestamp, false)

	if err != nil {
		return err
	}

	routerSampleDT, err := parseTimestamp((*routerData)[0].Timestamp, false)

	if err != nil {
		return err
	}

	diff := int(sampleDT.Sub(routerSampleDT).Seconds() / 3600)

	for _, row := range *routerData {
		dt, err := parseTimestamp(row.Timestamp, false)

		if err != nil {
			return err
		}

		row.Timestamp = dt.Add(time.Hour * time.Duration(diff)).Format(consts.DT_LAYOUT)
	}

	return err
}

// 	diff := int(sample_dt.Sub(router_dt).Seconds() / 3600)

// 	records := routerDF.Records()

// 	for i, row := range records {
// 		if i == 0 {
// 			continue
// 		}

// 		dt, err := time.Parse(consts.DT_LAYOUT, row[0])

// 		if err != nil {
// 			panic(err)
// 		}

// 		row[0] = dt.Add(time.Hour * time.Duration(diff)).Format(consts.DT_LAYOUT)
// 	}

// 	return dataframe.LoadRecords(records), err

// func GetJsonDataFile(sampleFiles []string, testID string) error {
// 	totalDataEntries := make([]DataRow, 0)
// 	scenarios := GetModesFromFiles(sampleFiles) // dict {"http": [file1, file2]}

// 	scenariosEntries := make(map[string][]DataRow)

// 	for scenario, files := range scenarios {
// 		scenrioEntries := make([]DataRow, 0)

// 		for index, file := range files {
// 			workingDir, err := os.Getwd()

// 			if err != nil {
// 				return err
// 			}

// 			localPath := fmt.Sprintf("%v/%v/%v/%v", workingDir, consts.RESULTS_DIR, testID, file)

// 			text, err := os.ReadFile(localPath)

// 			if err != nil {
// 				return err
// 			}

// 			df := dataframe.ReadCSV(strings.NewReader(string(text)))

// 			if index == 0 {
// 				continue
// 			}
// 		}
// 	}

// 	data := make(map[string]map[string][]Pair, 0)
// 	summedData :=  make(map[string]map[string]map[string]metrics, 0)
// 	totalData := make([]Pair, 0)

// 	for c, caseFiles := range cases {
// 		if data[c] == nil {
// 			data[c] = make(map[string][]Pair)
// 		}

// 		if summedData[SCENARIOS][c] == nil {
// 			summedData[SCENARIOS][c] = make(map[string]Pair)
// 		}

// 		caseTotal := make([]Pair, 0)

// 		for _, file := range caseFiles {
// 			workingDir, err := os.Getwd()

// 			if err != nil {
// 				return err
// 			}

// 			localPath := fmt.Sprintf("%v/%v/%v/%v", workingDir, consts.RESULTS_DIR, testID, file)

// 			text, err := os.ReadFile(localPath)

// 			if err != nil {
// 				return err
// 			}

// 			df := dataframe.ReadCSV(strings.NewReader(string(text))).Records()

// 			for i, row := range df {
// 				if i == 0 {
// 					continue
// 				}

// 				date, err := time.Parse(consts.DT_LAYOUT, row[0])

// 				if err != nil {
// 					return err
// 				}

// 				// changing the seconds to 0
// 				dateStr := time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, date.Location()).String()

// 				elapsed, err :=  strconv.ParseFloat(row[1], 32)

// 				if err != nil {
// 					return err
// 				}

// 				status, err :=  strconv.ParseFloat(row[2], 32)

// 				if err != nil {
// 					return err
// 				}

// 				pair := Pair{elasped: elapsed, status: status}

// 				caseTotal = append(caseTotal, pair)
// 				totalData = append(totalData, pair)

// 				if data[c][dateStr] == nil {
// 					data[c][dateStr] = make([]Pair, 0)
// 				}

// 				data[c][dateStr] = append(data[c][dateStr], pair)
// 			}

// 			metrics := GetCaseMetrics(caseTotal)

// 		}
// 	}
// 	return nil
// }

// func GetCaseMetrics(caseTotal []Pair) CaseMetrics {
// 	len := len(caseTotal)

// 	if len == 0 {
// 		return CaseMetrics{}
// 	}

// 	var errorCount int

// 	var sum float32
// 	min := caseTotal[0].elasped
// 	max := caseTotal[0].elasped

// 	for index, c := range caseTotal {
// 		if c.status == 1 {
// 			errorCount += 1
// 		}

// 		sum += c.elasped

// 		if index != 0 {
// 			if c.elasped < min {
// 				min = c.elasped
// 			}

// 			if c.elasped > max {
// 				max = c.elasped
// 			}
// 		}

// 	}

// 	return CaseMetrics{
// 		errorRate:  float32(errorCount) / float32(len),
// 		minElapsed: min,
// 		maxElapsed: max,
// 		avgElapsed: sum / float32(len),
// 	}
// }

// func GetFileMap(files []string) map[string][]string {
// 	modes := make(map[string][]string)

// 	for _, file := range files {
// 		mode := strings.Split(file, "_")[0]

// 		if modes[mode] == nil {
// 			modes[mode] = make([]string, len(files))
// 		}

// 		modes[mode] = append(modes[mode], file)

// 	}

// 	return modes
// }

// func GetRouterDF(sampleDF dataframe.DataFrame, routerFile string) (dataframe.DataFrame, error) {
// 	var err error

// 	routerDF, err := GetTimeAdjustedRouterDF(sampleDF, routerFile)

// 	if err != nil {
// 		return routerDF, err
// 	}

// 	routerDF, err = TrimRouterDF(routerDF, sampleDF)

// 	return routerDF, err

// }

// func TrimRouterDF(routerDF dataframe.DataFrame, sampleDF dataframe.DataFrame) (dataframe.DataFrame, error) {
// 	var adjutedRouterDF dataframe.DataFrame
// 	var err error

// 	start_time_str := sampleDF.Col("timestamp").Elem(0).String()
// 	end_time_str := sampleDF.Col("timestamp").Elem(sampleDF.Nrow() - 1).String()

// 	startTime, err := time.Parse(consts.DT_LAYOUT, start_time_str)

// 	if err != nil {
// 		return adjutedRouterDF, err
// 	}

// 	endTime, err := time.Parse(consts.DT_LAYOUT, end_time_str)
// 	fmt.Printf("start time: %v\n, end time: %v\n", startTime, endTime)

// 	if err != nil {
// 		return adjutedRouterDF, err
// 	}

// 	records := routerDF.Records()
// 	var timestamp time.Time

// 	var adjustedRecords [][]string

// 	for i, row := range records {
// 		if i == 0 {
// 			continue
// 		}

// 		timestamp, err = time.Parse(consts.DT_LAYOUT, row[0])

// 		if err != nil {
// 			return adjutedRouterDF, err
// 		}

// 		if !(timestamp.After(endTime) || timestamp.Before(startTime)) {
// 			adjustedRecords = append(adjustedRecords, row)
// 		}
// 	}

// 	adjutedRouterDF = dataframe.LoadRecords(adjustedRecords)

// 	return adjutedRouterDF, err
// }

// func GetTimeAdjustedRouterDF(sampleDF dataframe.DataFrame, routerFile string) (dataframe.DataFrame, error) {
// 	var adjutedRouterDF dataframe.DataFrame

// 	text, err := os.ReadFile(routerFile)

// 	if err != nil {
// 		return adjutedRouterDF, err
// 	}

// 	routerDF := dataframe.ReadCSV(strings.NewReader(string(text)))

// 	sample_dt, err := time.Parse(consts.DT_LAYOUT, sampleDF.Col("timestamp").Elem(0).String())

// 	if err != nil {
// 		return adjutedRouterDF, err
// 	}

// 	router_dt, err := time.Parse(consts.DT_LAYOUT, routerDF.Col("timestamp").Elem(0).String())

// 	if err != nil {
// 		return adjutedRouterDF, err
// 	}

// 	diff := int(sample_dt.Sub(router_dt).Seconds() / 3600)

// 	records := routerDF.Records()

// 	for i, row := range records {
// 		if i == 0 {
// 			continue
// 		}

// 		dt, err := time.Parse(consts.DT_LAYOUT, row[0])

// 		if err != nil {
// 			panic(err)
// 		}

// 		row[0] = dt.Add(time.Hour * time.Duration(diff)).Format(consts.DT_LAYOUT)
// 	}

// 	return dataframe.LoadRecords(records), err
// }
