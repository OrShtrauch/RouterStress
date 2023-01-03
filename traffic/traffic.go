package traffic

import (
	"RouterStress/consts"
	"RouterStress/docker"
	"encoding/json"
	"fmt"
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

func RunTrafficCapture(d *docker.Docker, cb func() error) TrafficMessage {
	var traffic TrafficData
	var err error

	c, err := d.StartTrafficCaptureContainer()

	if err != nil {
		return TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
	}

	err = cb()

	if err != nil {
		return TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
	}

	channel := make(chan TrafficMessage)

	go func(channel chan TrafficMessage) {		
		jsonData, err := ListenForTrafficData()

		if err != nil {
			channel <- TrafficMessage{
				Data:  TrafficData{},
				Error: err,
			}
			return
		}

		traffic, err = parseJsonData(jsonData)

		channel <- TrafficMessage{
			Data:  traffic,
			Error: err,
		}
	}(channel)

	if err = d.KillContainer(c); err != nil {
		return TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
	}

	msg := <-channel
	close(channel)

	fmt.Printf("msg: %v\n", msg.Data)

	return msg
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
			fmt.Printf("data: %v\n", string(buffer[:mLen]))
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
