package traffic

type TrafficData struct {
	Total           string
	Retransmissions string
}

// func Get() (TrafficData, error) {
// 	var trafficData TrafficData

// 	data, err := os.ReadFile(consts.TRAFFIC_DATA_PATH)

// 	if err != nil {
// 		return trafficData, err
// 	}

// 	err = json.Unmarshal(data, &trafficData)

// 	return trafficData, err
// }
