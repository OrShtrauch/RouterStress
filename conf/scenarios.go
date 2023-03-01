package conf

import (
	"RouterStress/consts"
	"encoding/json"
	"os"
)

type Scenarios struct {
	Scenarios []Scenario
}

type Scenario struct {
	Name string `json:"name"`
	PipDependencies string `json:"pip_dependencies"`
	Script string `json:"script"`
	Params []string `json:"params"`
}

func GetScenarios() (Scenarios, error) {
	var scenarios Scenarios
	jsonData, err := os.ReadFile(consts.SCENARIOS_PATH)

	if err != nil {
		return scenarios, err
	}

	err = json.Unmarshal(jsonData, &scenarios)

	return scenarios, err
	
}
