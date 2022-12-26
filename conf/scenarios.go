package conf

import (
	"RouterStress/consts"
	"encoding/json"
	"os"
)

type Scenario struct {
	Name string
	PipDependencies string
	Script string
	Params []string
}

func GetScenarios() (Config, error) {
	var config Config
	jsonData, err := os.ReadFile(consts.SCENARIOS_PATH)

	if err != nil {
		return config, err
	}

	err = json.Unmarshal(jsonData, &config)

	return config, err
	
}
