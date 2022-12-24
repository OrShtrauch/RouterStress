package conf

import (
	"RouterStress/consts"
	"encoding/json"
	"fmt"
	"os"
)

func GetConfig() (Config, error) {
	var config Config
	jsonData, err := os.ReadFile(consts.CONF_PATH)

	if err != nil {
		return config, err
	}

	err = json.Unmarshal(jsonData, &config)

	return config, err
}

func ParseIteration(iteration Iteration) []map[string]string {
	var iterationMap []map[string]string
    
	for _, protocol := range iteration.Protocols {
		for _, container := range protocol.Containers {
			m :=  make(map[string]string)

			precentToIncrease := 50
			new_amount := container.Amount
			
			if consts.Run_index > 0 {
				new_amount = int(new_amount * (precentToIncrease/100 * consts.Run_index) + 1) + 1
			}

			m["mode"] = protocol.Mode
			m["amount"] = fmt.Sprint(new_amount)

			for key, value := range container.ConfParams {
				m[key] = value
			}

			iterationMap = append(iterationMap, m)
		}
	}

	return iterationMap
}

type Config struct {
	Network Network
	Iterations []Iteration
}

type Network struct {
	Ssid string
	Interface string
}

type Iteration struct {
	Duration int
	Cooldown int
	Protocols []Protocol
}

type Protocol struct {
	Mode string
	Containers []Container
}

type Container struct {
	Amount int
	ConfParams map[string]string
}



