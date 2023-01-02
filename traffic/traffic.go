package traffic

import (
	"RouterStress/consts"
	"RouterStress/docker"
	"encoding/json"
	"net"
	"os"
)

type TrafficData struct {
	Total           string
	Retransmissions string
}

type TrafficMessage struct {
	Data  TrafficData
	Error error
}

func RunTrafficCapture(d *docker.Docker, cb func() error) {
	var traffic TrafficData
	var err error

	c, err := d.StartTrafficCaptureContainer(duration, true)

	if err != nil {
		channel <- TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
		return
	}
}

func RunInitialTrafficCapture(d *docker.Docker, duration int, channel chan TrafficMessage) {
	var traffic TrafficData
	var err error

	c, err := d.StartTrafficCaptureContainer(duration, true)

	if err != nil {
		channel <- TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
		return
	}

	jsonData, err := ListenForTrafficData()

	if err != nil {
		channel <- TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
		return
	}

	traffic, err = parseJsonData(jsonData)

	if err != nil {
		channel <- TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
		return
	}

	if err = d.KillContainer(c); err != nil {
		channel <- TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
		return
	} else {
		channel <- TrafficMessage{
			Data:  traffic,
			Error: err,
		}
		return
	}
}

func ListenForTrafficData() (string, error) {
	var data string
	var err error

	if err = os.RemoveAll(consts.TRAFFIC_UNIX_SOCKET); err != nil {
		return data, err
	}

	listener, err := net.Listen("unix", consts.TRAFFIC_UNIX_SOCKET)
	if err != nil {
		return data, err
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			return data, err
		}

		buffer := make([]byte, 512)
		if mLen, err := conn.Read(buffer); err != nil {
			return data, err
		} else {
			return string(buffer[:mLen]), err
		}
	}
}

func parseJsonData(jsonData string) (TrafficData, error) {
	var data TrafficData

	err := json.Unmarshal([]byte(string(jsonData)), &data)

	return data, err
}

/*
file := fmt.Sprintf("%v%v", consts.RESULTS_DIR, consts.TRAFFIC_DATA_NAME)
	fmt.Println(file)
	jsonData, err := os.ReadFile(file)

	if err != nil {
		return data, err
	}

	fmt.Printf("len: %v\n", len(jsonData))
	fmt.Printf("file %v\n",string(jsonData))

	err = os.Remove(file)

	if err != nil {
		return data, err
	}

	err = json.Unmarshal([]byte(string(jsonData)), &data)
	fmt.Printf("data: %v\n", data)

	if err != nil {
		fmt.Println("error")
		fmt.Println(err.Error())
		return data, err
	}

	return data, err
*/
