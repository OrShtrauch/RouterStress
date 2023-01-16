package traffic

import (
	"RouterStress/consts"
	"RouterStress/docker"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// type TrafficData struct {
// 	Total           string
// 	Retransmissions string
// }

type TrafficData struct {
	Percent string
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

	channel := make(chan TrafficMessage)

	go func(channel chan TrafficMessage) {
		jsonData, err := ListenForTrafficData()
		fmt.Printf("got data: %v\n", jsonData)

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
	
	err = cb()

	if err != nil {
		return TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
	}

	fmt.Println("calling killContainer")
	if err = d.KillContainer(c.ID); err != nil {		
		return TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
	}

	msg := <-channel
	close(channel)

	fmt.Printf("download: %v", msg.Data.Percent)
	return msg
}

func ListenForTrafficData() (string, error) {
	var data string
	var err error

	if err = os.RemoveAll(consts.TRAFFIC_UNIX_SOCKET); err != nil {
		return data, err
	}

	listener, err := net.Listen(consts.UNIX, consts.TRAFFIC_UNIX_SOCKET)
	if err != nil {
		return data, err
	}

	for {
		conn, err := listener.Accept()
		fmt.Println("got connection: ")
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
