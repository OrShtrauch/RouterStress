// Package containing all the data models for the configuration object, and helper functions
package conf

import (
	"RouterStress/consts"
	"RouterStress/errors"
	"RouterStress/router"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Config struct {
	Settings   Settings
	Network    Network
	Iterations []Iteration
	Router     router.RouterData
	Scenarios  Scenarios
}

type Settings struct {
	S3          bool
	Debug       bool
	IpefHost    string `json:"iperf_host"`
	IperfPort   int    `json:"iperf_port"`
	Recursive   bool
	PercentDiff float64 `json:"percent_diff"`
}

type Network struct {
	Ssid   string
	Parent string
}

type Iteration struct {
	Duration  int
	Cooldown  int
	Protocols []Protocol
}

type Protocol struct {
	Mode       string
	Containers []Container
}

type Container struct {
	Amount int
	Params map[string]string
}

func GetConfig() (Config, error) {
	var config Config
	jsonData, err := os.ReadFile(consts.CONF_PATH)

	if err != nil {
		return config, err
	}

	err = json.Unmarshal(jsonData, &config)

	if err != nil {
		return config, err
	}

	scenarios, err := GetScenarios()

	if err != nil {
		return config, err
	}

	config.Scenarios = scenarios

	config.BuildDockerFiles()

	routers, err := router.LoadRouters(consts.ROUTERS_PATH)

	if err != nil {
		return config, err
	}

	for _, router := range routers {
		if router.Ssid == config.Network.Ssid {
			config.Router = router

			return config, nil
		}
	}

	return config, errors.NoSSIDFound{}
}

func ParseIteration(iteration Iteration) []map[string]string {
	var iterationMap []map[string]string

	for _, protocol := range iteration.Protocols {
		for _, container := range protocol.Containers {
			m := make(map[string]string)

			precentToIncrease := 50
			new_amount := container.Amount

			if consts.RUN_INDEX > 0 {
				new_amount = int(new_amount*(precentToIncrease/100*consts.RUN_INDEX)+1) + 1
			}

			m["mode"] = protocol.Mode
			m["amount"] = fmt.Sprint(new_amount)

			for key, value := range container.Params {
				m[key] = value
			}

			iterationMap = append(iterationMap, m)
		}
	}

	return iterationMap
}

func (c *Config) BuildDockerFiles() error {
	var eg errgroup.Group

	for _, s := range c.Scenarios.Scenarios {
		scenario := s
		eg.Go(func() error {
			return writeDockerFile(scenario)
		})
	}

	return eg.Wait()
}

func writeDockerFile(s Scenario) error {
	wd, err := os.Getwd()

	if err != nil {
		return err
	}

	newDockerFile := fmt.Sprintf("%v/%vDockerfile.%v", wd, consts.DOCKERFILES_PATH, s.Name)
	err = os.Remove(newDockerFile)

	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	templateFile := fmt.Sprintf("%v/Dockerfile_template", consts.DOCKERFILES_PATH)
	dockerFile, err := os.ReadFile(templateFile)
	dockerFileText := string(dockerFile)

	if err != nil {
		return err
	}

	pipCmd := ""
	if s.PipDependencies != "" {
		pipCmd = fmt.Sprintf("RUN pip3 install %v", s.PipDependencies)
	}

	dockerFileText = strings.Replace(dockerFileText, "{script_name}", s.Script, -1)
	dockerFileText = strings.Replace(dockerFileText, "{pip}", pipCmd, -1)

	err = os.WriteFile(newDockerFile, []byte(dockerFileText), 0644)

	return err
}
