package router

import (
	"RouterStress/consts"
	"RouterStress/log"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

type Slave struct {
	Data       RouterData
	Client     Client
	SamplerPID string
	SamplerDir string
}

var CmdsToReplace = []string{"awk", "nc"}

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

		} else {
			c, err = NewTelnetClient(selectedRouter.Ip, selectedRouter.Communication_port, selectedRouter.Login_info...)

			if err != nil {
				return slave, err
			}

			slave = &Slave{
				Data:   selectedRouter,
				Client: c,
			}
		}
	}

	path := strings.Split(slave.Data.Sam_dir, "/")
	path[len(path)-1] = ""

	slave.SamplerDir = strings.Join(path, "/")

	if err := slave.CreateSamplerDir(); err != nil {
		return slave, err
	}

	return slave, err
}

func (s *Slave) StartSampler() error {
	cmd := fmt.Sprintf("$SHELL %v%v %v &", s.SamplerDir, consts.SAMPLER_NAME, consts.SAMPLER_DELAY)

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
			return err
		}

		connection.Close()

		return err

	})

	if err != nil {
		return err
	}

	_, err = s.Run(fmt.Sprintf("chmod +x %s", fmt.Sprintf("%v%v", s.SamplerDir, consts.SAMPLER_NAME)))

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
		cmd_direction = fmt.Sprintf(" >  %v%v", s.SamplerDir, consts.SAMPLER_NAME)
	} else {
		cmd_direction = fmt.Sprintf(" < %v%v", s.SamplerDir, consts.SAMPLER_DATA_NAME)
	}

	cmd := fmt.Sprintf("%s -lp %s %s", s.GetCommand(consts.NETCAT), port, cmd_direction)

	if _, err := s.Run(fmt.Sprintf("%v &", cmd)); err != nil {
		return err
	}

	log.Logger.Debug("running netcat listener")

	time.Sleep(time.Second)

	err := cb()

	s.Client.CloseListenerSession(fmt.Sprintf("ps | grep %s | grep %s | grep -v grep | %s '{print $1}'",
		consts.NETCAT, consts.SOCKET_PORT, s.GetCommand(consts.AWK)))

	return err
}

func (s *Slave) GetRouterSampler() (string, error) {
	data, err := os.ReadFile(consts.LOCAL_SAMPLER_PATH)

	textString := string(data)

	if err != nil {
		return "", err
	}

	for _, cmd := range CmdsToReplace {
		textString = strings.Replace(textString, cmd, s.GetCommand(cmd), -1)
	}

	textString = strings.ReplaceAll(textString, "SAMPLER_DATA", fmt.Sprintf("%v%v", s.SamplerDir, consts.SAMPLER_DATA_NAME))

	return textString, err
}

func (s *Slave) GetCommand(cmd string) string {

	if s.contains(cmd, CmdsToReplace) {
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

func (s *Slave) CreateSamplerDir() error {
	var err error

	cmd := fmt.Sprintf("%v %v", "mkdir", s.SamplerDir)

	_, err = s.Run(cmd)

	return err
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
