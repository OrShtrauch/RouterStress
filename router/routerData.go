package router

import (
	"os"
	"encoding/json"
)

type RouterData struct {
	Name               string
	Ip                 string
	Ssid               string
	Password           string
	Busybox            string
	Sam_dir            string
	Login_info         []string
	Protocol           string
	Communication_port string
}

func LoadRouters(path string) ([]RouterData, error) {
	var routers []RouterData

	jsonData, err := os.ReadFile(path)

	if err != nil {
		return routers, err
	}
	err = json.Unmarshal(jsonData, &routers)

	if err != nil {
		return routers, err
	}

	return routers, err
}

func (r *RouterData) IsEmpty() bool {
	return r.Name == ""
}