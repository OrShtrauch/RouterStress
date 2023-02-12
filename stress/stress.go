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
	"math"
	"math/rand"
	"os"

	// "strconv"
	"time"

	"github.com/google/uuid"
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

	rand.Seed(time.Now().UnixNano())

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

	log.Logger.Info("Running inital traffic capture")
	trafficMessage := traffic.RunTrafficCapture(stress.Docker,
		consts.INITIAL_CAPTURE_DURATION, config.Settings.IpefHost, config.Settings.IperfPort,
		func() error {
			time.Sleep(time.Second * consts.INITIAL_CAPTURE_DURATION)
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

	totalTime := CalcTotalTime(s.Config)

	log.Logger.Info("starting")
	log.Logger.Sugar().Infof("Run Index: %v", consts.RUN_INDEX)

	for s.ShouldRunAgain(&data.Data, initial) {
		data = traffic.RunTrafficCapture(s.Docker, totalTime, s.Config.Settings.IpefHost, s.Config.Settings.IperfPort, func() error {
			var eg errgroup.Group
			var err error

			for iterationIndex, iteration := range s.Config.Iterations {
				duration := iteration.Duration

				index := 0
				for _, protocol := range iteration.Protocols {
					for j, container := range protocol.Containers {
						amount := s.GetAdjustedAmount(container.Amount)
						log.Logger.Sugar().Debugf("amount is %v", amount)
						for i := 0; i < amount; i++ {
							uid := uuid.New().String()[:5]
							name := fmt.Sprintf("stress_%v_%v_%v_%v_%v", protocol.Mode, uid, consts.RUN_INDEX, iterationIndex, index)
							env := protocol.Containers[j].Params
							mode := protocol.Mode

							eg.Go(func() error {
								// sleep for a random time between 1 and 5 seconds
								// to space out even more the containers
								time.Sleep(GenerateRandomSleep())

								return s.runStressContainer(name, mode, duration, iterationIndex, index, env)
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
	}

	return nil
}

func CalcTotalTime(config *conf.Config) int {
	var total int

	for _, iteration := range config.Iterations {
		total += (iteration.Duration * 60) + iteration.Cooldown
	}

	return total
}

func GenerateRandomSleep() time.Duration {
	min := 1
	max := 5

	return time.Duration(rand.Intn(max-min) + min)
}

func (s *Stress) ShouldRunAgain(data *traffic.TrafficData, initial bool) bool {
	if initial {
		return initial
	}

	if !s.Config.Settings.Recursive {
		return false
	}

	initialPercent := s.InitialCaptureData.Loss / s.InitialCaptureData.Total

	percent := data.Loss / data.Total

	log.Logger.Sugar().Debugf("inital percent: %v", initialPercent)
	log.Logger.Sugar().Debugf("current percent: %v", percent)

	percentDiff := math.Abs(percent - initialPercent)

	should_run_again := percentDiff < s.Config.Settings.PercentDiff

	log.Logger.Debug(fmt.Sprintf("percent diff: %v", percentDiff))

	if should_run_again {
		consts.RUN_INDEX += 1
		log.Logger.Info("running again with increased amount")
	}

	return should_run_again
}

func (s *Stress) GetAdjustedAmount(amount int) int {
	if consts.RUN_INDEX == 0 {
		return amount
	}

	maxAmount := len(s.DHCPServer.Pool)
	if amount > maxAmount {
		log.Logger.Sugar().Infof("%v is bigger than the server pool (%v), running with max amount", amount, maxAmount)
		return maxAmount
	}

	value := float64(amount) * ((0.5 * float64(consts.RUN_INDEX)) + 1)

	return int(math.Ceil(value))
}

func (s *Stress) runStressContainer(name string, mode string, duration int, iterationIndex int, index int, env map[string]string) error {
	ip, err := s.DHCPServer.Lease()

	if err != nil {
		return err
	}

	//log.Logger.Debug(fmt.Sprintf("starting container %v", name))

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

	//log.Logger.Debug(fmt.Sprintf("Killing Container %v", name))

	return s.Docker.KillContainer(container.ID)
}

func setupSlave(stress *Stress) error {
	slave, err := router.NewSlave(stress.Config.Network.Ssid)
	stress.Slave = slave

	if err != nil {
		return err
	}

	log.Logger.Debug("created slave object")

	err = slave.TransferSamplerToRouter()

	if err != nil {
		return err
	}

	log.Logger.Debug("transferred sampler, now starting sampler")

	return slave.StartSampler()
}

func setupDocker(stress *Stress) error {
	docker, err := docker.InitDocker(stress.Config)

	if err != nil {
		return err
	}
	log.Logger.Debug("docker init done, now building images")
	stress.Docker = docker

	return docker.BuildModeImages(stress.Config)
}

func setupDHCP(stress *Stress) error {
	dhcp := dhcp.NewDHCPServer(stress.Config.Router.GetSubnet())
	dhcp.PopulatePool()

	log.Logger.Debug("dhcp init done")
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
