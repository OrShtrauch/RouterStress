package router

import (
	"bytes"
	"fmt"
	"io"	
	"net"
	"os"
	"strings"
	"time"
	"RouterStress/consts"
)

type Slave struct {
	Data       RouterData
	Client     Client
	SamplerPID string
}

var cmds_to_replace = []string{"awk", "nc"}

func NewSlave(ssid string) (*Slave, error) {
	var slave *Slave
	var selectedRouter RouterData
	var c Client

	routers, err := LoadRouters(consts.ROUTERS_PATH)

	if err != nil {
		return slave, err
	}

	for _, r := range routers {
		if r.Ssid == ssid {
			selectedRouter = r

			break
		}
	}

	if !selectedRouter.IsEmpty() {
		if selectedRouter.Protocol == consts.SSH {
			var username string
			var password string

			if len(selectedRouter.Login_info) == 2 {
				username = selectedRouter.Login_info[0]
				password = selectedRouter.Login_info[1]
			}

			c, err = NewSSHClient(selectedRouter.Ip, selectedRouter.Communication_port, username, password)

			if err != nil {
				return slave, err
			}

			slave = &Slave{
				Data:   selectedRouter,
				Client: c,
			}

			return slave, err
		} else {
			c, err = NewTelnetClient(selectedRouter.Ip, selectedRouter.Communication_port, selectedRouter.Login_info...)

			if err != nil {
				return slave, err
			}

			slave = &Slave{
				Data:   selectedRouter,
				Client: c,
			}

			return slave, err
		}
	}

	return slave, err
}

func (s *Slave) StartSampler() error {
	cmd := fmt.Sprintf("$SHELL %v %v &", consts.SAMPLER_PATH, consts.SAMPLER_DELAY)

	_, err := s.Run(cmd)
	
	if err != nil {
		return err
	}

	output, err := s.Run("echo $!")

	if err != nil {
		return err
	}

	s.SamplerPID = strings.TrimSpace(output)
	return err
}

func (s *Slave) StopSampler() error {
	cmd := fmt.Sprintf("kill -9 %v", s.SamplerPID)

	_, err := s.Run(cmd)

	return err
}

func (s *Slave) Run(cmd string) (string, error) {
	data, err := s.Client.Run(cmd)

	return data, err
}

func (s *Slave) TransferSamplerToRouter() error {
	toRouter := true
	err := s.WithListener(consts.SOCKET_PORT, toRouter, func() error {
		host := fmt.Sprintf("%v:%v", s.Data.Ip, consts.SOCKET_PORT)
		connection, err := net.Dial(consts.TCP, host)

		if err != nil {
			return err
		}
		samplerData, err := s.GetRouterSampler()

		if err != nil {
			return err
		}

		time.Sleep(5 * time.Second)
		_, err = connection.Write([]byte(samplerData))

		if err != nil {
			panic(err)
		}

		connection.Close()

		return err

	})

	s.Run(fmt.Sprintf("chmod +x %s", consts.SAMPLER_PATH))

	return err
}

func (s *Slave) GetSamplerData() (string, error) {
	var data string

	toRouter := false
	err := s.WithListener(consts.SOCKET_PORT, toRouter, func() error {
		host := fmt.Sprintf("%v:%v", s.Data.Ip, consts.SOCKET_PORT)
		connection, err := net.Dial(consts.TCP, host)

		if err != nil {
			return err
		}

		var buf bytes.Buffer
		io.Copy(&buf, connection)

		data = buf.String()
		connection.Close()

		return err
	})

	return data, err
}

func (s *Slave) WithListener(port string, toRouter bool, cb func() error) error {
	var cmd_direction string

	if toRouter {
		cmd_direction = fmt.Sprintf(" >  %v", consts.SAMPLER_PATH)
	} else {
		cmd_direction = fmt.Sprintf(" < %v", consts.SAMPLER_DATA_PATH)
	}

	cmd := fmt.Sprintf("%s -lp %s %s", s.GetCommand(consts.NETCAT), port, cmd_direction)

	s.Run(fmt.Sprintf("%v &", cmd))

	time.Sleep(time.Second)

	err := cb()

	s.Client.CloseListenerSession(fmt.Sprintf("ps | grep %s | grep -v grep | %s '{print $1}'",
		consts.NETCAT, s.GetCommand(consts.AWK)))

	return err
}

func (s *Slave) GetRouterSampler() (string, error) {
	data, err := os.ReadFile(consts.LOCAL_SAMPLER_PATH)

	text_string := string(data)

	if err != nil {
		return "", err
	}

	for _, cmd := range cmds_to_replace {
		text_string = strings.Replace(text_string, cmd, s.GetCommand(cmd), -1)
	}

	return text_string, err
}

func (s *Slave) GetCommand(cmd string) string {

	if s.contains(cmd, cmds_to_replace) {
		return fmt.Sprintf("%v %v", s.Data.Busybox, cmd)
	} else {
		return cmd
	}
}

func (s *Slave) contains(cmd string, slice []string) bool {
	for _, v := range slice {
		if v == cmd {
			return true
		}
	}
	return false
}

func (s *Slave) Cleanup() error {
	err := s.StopSampler()

	if err != nil {
		return err
	}

	data, err := s.GetSamplerData()

	if err != nil {
		return err
	}

	return writeSamplerData(data)

}

func writeSamplerData(data string) error {
	path := fmt.Sprintf("results/%v/%v", consts.TEST_ID, consts.SAMPLER_LOCAL_NAME)

	return os.WriteFile(path, []byte(data), 0644)
}
