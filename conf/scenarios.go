package conf

import (
	"RouterStress/consts"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Scenarios struct {
	Scenarios []Scenario
}

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

func (s *Scenarios) BuildDockerFiles() error {
	var err error

	for _, scenario := range s.Scenarios {
		err := BuildDockerFile(scenario)

		if err!= nil {
            return err
        }
	}

	return err
}

func BuildDockerFile(s Scenario) error {
	scriptName := strings.Split(s.Name, "/")[1]

	templateFile := fmt.Sprintf("%v/Dockerfile.template", consts.DOCKERFILES_PATH)
	dockerFile, err := os.ReadFile(templateFile)
	dockerFileText := string(dockerFile)

	if err != nil {
		return err 
	}

	dockerFileText = strings.Replace(dockerFileText, "{script_path}", consts.CONTAINER_SCRIPTS, -1)
	dockerFileText = strings.Replace(dockerFileText, "{script_name}", scriptName, -1)
	dockerFileText = strings.Replace(dockerFileText, "{pip}", s.PipDependencies, -1)

	newDockerFile := fmt.Sprintf("%v/Dockerfile.%v", consts.DOCKERFILES_PATH, s.Name)

	err = os.WriteFile(newDockerFile, []byte(dockerFileText), 0644)

	return err
}