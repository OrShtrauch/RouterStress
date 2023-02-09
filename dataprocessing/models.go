package dataprocessing

type Tuple struct {
	DataRows []DataRow
	Clients  int
}

type DataRow struct {
	Timestamp string  `csv:"timestamp"`
	Elasped   float32 `csv:"elapsed"`
	Status    float32 `csv:"status"`
}

type RouterDataRow struct {
	Timestamp string  `csv:"timestamp"`
	Cpu       float32 `csv:"cpu"`
	Mem       float32 `csv:"mem"`
}

type Metrics struct {
	Clients    *int     `json:"clients,omitempty"`
	Requests   *int     `json:"requests,omitempty"`
	RunIndex   *int     `json:"run_index,omitempty"`
	SpeedRatio *float32 `json:"speed_ratio,omitempty"`
	ErrorRate  *float32 `json:"error_rate,omitempty"`
	MinElapsed *float32 `json:"min_elapsed,omitempty"`
	MaxElapsed *float32 `json:"max_elapsed,omitempty"`
	AvgElapsed *float32 `json:"avg_elapsed,omitempty"`
}

type Scenario struct {
	Name    string              `json:"name"`
	Total   *Metrics            `json:"total"`
	Minutes map[string]*Metrics `json:"minutes"`
}

type TestMetrics struct {
	Total     *Metrics    `json:"total"`
	Scenarios []*Scenario `json:"scenarios"`
}

type NoFilesFound struct{}

func (n NoFilesFound) Error() string {
	return "No Files Found With given file Index"
}
