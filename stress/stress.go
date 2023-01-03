package stress

import (
	"RouterStress/conf"
	"RouterStress/consts"
	"RouterStress/dhcp"
	"RouterStress/docker"
	"RouterStress/log"
	"RouterStress/router"
	"RouterStress/traffic"
	"fmt"
	"os"
	"math"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
)

type Stress struct {
	TestID             string
	Config             *conf.Config
	Docker             *docker.Docker
	DHCPServer         *dhcp.DHCPServer
	Slave              *router.Slave
	InitialCaptureData *traffic.TrafficData
}

func NewStress(config *conf.Config) (Stress, error) {
	var eg errgroup.Group
	var stress Stress

	stress.Config = config

	stress.TestID = fmt.Sprintf("%v_%v", config.Router.Ssid, consts.TEST_UUID)
	consts.TEST_ID = stress.TestID

	log.Logger.Info(fmt.Sprintf("TestID: %v", stress.TestID))

	createTestDir(stress.TestID)

	eg.Go(func() error {
		log.Logger.Debug("Setting up Router Slave")
		return setupSlave(&stress)
	})

	eg.Go(func() error {
		log.Logger.Debug("Setting up Docker Client and Network")
		return setupDocker(&stress)
	})

	eg.Go(func() error {
		log.Logger.Debug("setting up DHCP Server")
		return setupDHCP(&stress)
	})

	if err := eg.Wait(); err != nil {
		return stress, err
	}

	trafficMessage := traffic.RunTrafficCapture(stress.Docker, func() error {
		time.Sleep(time.Second * 60)
		return nil
	})		

	if trafficMessage.Error != nil {
		return stress, trafficMessage.Error
	}

	stress.InitialCaptureData = &trafficMessage.Data
	
	return stress, nil
}

func (s *Stress) Start() error {
	var data traffic.TrafficMessage
	initial := true

	for s.ShouldRunAgain(&data.Data, initial) {				
		data = traffic.RunTrafficCapture(s.Docker, func() error {
			var eg errgroup.Group
			var err error

			for iterationIndex, iteration := range s.Config.Iterations {
				duration := iteration.Duration

				index := 0
				for _, protocol := range iteration.Protocols {
					for j, container := range protocol.Containers {
						for i := 0; i < s.GetAdjustedAmount(container.Amount); i++ {
							name := fmt.Sprintf("stress_%v_%v_%v_%v", protocol.Mode, consts.RUN_INDEX, iterationIndex, index)

							env := protocol.Containers[j].Params

							eg.Go(func() error {
								return s.runStressContainer(name, protocol.Mode, duration, iterationIndex, index, env)
							})

							index++
						}
					}
				}

				if err = eg.Wait(); err != nil {
					return err
				}

				time.Sleep(time.Duration(iteration.Cooldown) * time.Second)
			}

			return err
		})

		if data.Error != nil {
			return data.Error
		}

		initial = false
		consts.RUN_INDEX += 1
	}

	
	return nil 
}


func (s *Stress) ShouldRunAgain(data *traffic.TrafficData, initial bool) bool {
	if initial {
		return initial
	}

	initialTotal, _ := strconv.ParseFloat(s.InitialCaptureData.Total, 32)
	initialRetransmissions, _ := strconv.ParseFloat(s.InitialCaptureData.Retransmissions, 32)


	total, _ := strconv.ParseFloat(data.Total, 32)
	retransmissions, _ := strconv.ParseFloat(data.Retransmissions, 32)

	initialRtRate := initialRetransmissions / initialTotal
	rtRate := retransmissions / total
	
	should_run_again := rtRate / initialRtRate < consts.RT_MAX_RATE

	log.Logger.Debug(fmt.Sprintf("packet loss ratio: %v < %v is %v", rtRate / initialRtRate, consts.RT_MAX_RATE, should_run_again))

	if should_run_again {
		log.Logger.Info("running again with increased amount")
	}	
	
	return should_run_again
}

func (s *Stress) GetAdjustedAmount(amount int) int {	
	value := float64(amount) * ((0.5 * float64(consts.RUN_INDEX)) + 1)

	return int(math.Ceil(value))
}

func (s *Stress) runStressContainer(name string, mode string, duration int, iterationIndex int, index int, env map[string]string) error {
	ip, err := s.DHCPServer.Lease()

	if err != nil {
		return err
	}

	log.Logger.Debug(fmt.Sprintf("starting container %v", name))
	container, err := s.Docker.RunContainer(docker.ContainerData{
		Ssid:           s.Config.Router.Ssid,
		Platform:       s.Config.Router.Name,
		RunIndex:       consts.RUN_INDEX,
		TestID:         s.TestID,
		Mode:           mode,
		Name:           name,
		Ip:             ip,
		Index:          index,
		IterationIndex: iterationIndex,
		Params:         env,
	})

	if err != nil {
		return err
	}

	time.Sleep(time.Duration(duration) * time.Minute)

	s.DHCPServer.Release(ip)

	return s.Docker.KillContainer(container.ID)
}

func setupSlave(stress *Stress) error {
	slave, err := router.NewSlave(stress.Config.Network.Ssid)
	stress.Slave = slave

	if err != nil {
		return err
	}

	err = slave.TransferSamplerToRouter()

	if err != nil {
		return err
	}

	return slave.StartSampler()
}

func setupDocker(stress *Stress) error {
	docker, err := docker.InitDocker(stress.Config)

	if err != nil {
		return err
	}

	stress.Docker = docker

	return docker.BuildImages(stress.Config)
}

func setupDHCP(stress *Stress) error {
	dhcp := dhcp.NewDHCPServer(stress.Config.Router.GetSubnet())
	dhcp.PopulatePool()

	stress.DHCPServer = dhcp

	return nil
}

func (s *Stress) Cleanup() error {
	var eg errgroup.Group

	log.Logger.Debug("Running cleanup")

	eg.Go(func() error {
		return s.Slave.Cleanup()
	})

	eg.Go(func() error {
		return s.Docker.Cleanup()
	})

	return eg.Wait()
}

func createTestDir(testID string) error {
	dirName := fmt.Sprintf("results/%v", testID)

	err := os.Mkdir(dirName, 0766)

	if err != nil {
		return err
	}

	return err
}
