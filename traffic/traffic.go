package traffic

import (
	"RouterStress/consts"
	"RouterStress/docker"
	"RouterStress/log"
	"encoding/json"
	"net"
	"os"
)

// type TrafficData struct {
// 	Total           string
// 	Retransmissions string
// }

type TrafficData struct {
	Loss  float64 `json:",string"`
	Total float64 `json:",string"`
}

type TrafficMessage struct {
	Data  TrafficData
	Error error
}

func RunTrafficCapture(d *docker.Docker, duration int, host string, port int, cb func() error) TrafficMessage {
	var traffic TrafficData
	var err error

	c, err := d.StartTrafficCaptureContainer(duration, host, port)

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

		if err != nil {
			channel <- TrafficMessage{
				Data:  TrafficData{},
				Error: err,
			}
			return
		}

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

	if err = d.WaitForContainerToDie(c); err != nil {
		return TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
	}

	msg := <-channel

	if msg.Error != nil {
		return TrafficMessage{
			Data:  TrafficData{},
			Error: err,
		}
	}

	log.Logger.Sugar().Debugf("losss is %v, total is %v", msg.Data.Loss, msg.Data.Total)
	close(channel)

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
