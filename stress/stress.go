package stress

import (
	"RouterStress/conf"
	"RouterStress/consts"
	"RouterStress/dhcp"
	"RouterStress/docker"
	"RouterStress/logger"
	"RouterStress/router"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"time"

	"golang.org/x/sync/errgroup"
)

type Stress struct {
	TestID     string
	Config     *conf.Config
	Docker     *docker.Docker
	DHCPServer *dhcp.DHCPServer
	Slave      *router.Slave
}

func NewStress(config *conf.Config) (Stress, error) {
	var eg errgroup.Group
	var stress Stress

	stress.Config = config
	consts.TestID = fmt.Sprintf("%v_%v", config.Router.Ssid, consts.TestUUID)
	logger.Logger.Info(fmt.Sprintf("TestID: %v", consts.TestID))

	createTestDir(consts.TestID)

	eg.Go(func() error {
		logger.Logger.Debug("Setting up Router Slave")
		return setupSlave(&stress)
	})

	eg.Go(func() error {
		logger.Logger.Debug("Setting up Docker Client and Network")
		return setupDocker(&stress)
	})

	eg.Go(func() error {
		logger.Logger.Debug("setting up DHCP Server")
		return setupDHCP(&stress)
	})

	return stress, eg.Wait()
}

func (s *Stress) Start() error {
	var eg errgroup.Group
	var err error

	for iterationIndex, iteration := range s.Config.Iterations {
		duration := iteration.Duration

		index := 0
		for _, protocol := range iteration.Protocols {
			for j, container := range protocol.Containers {
				for i := 0; i < container.Amount; i++ {
					name := fmt.Sprintf("stress_%v_%v_%v", protocol.Mode, iterationIndex, index)

					env := protocol.Containers[j].Params

					eg.Go(func() error {
						return s.RunStressContainer(name, protocol.Mode, duration, iterationIndex, index, env)
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
}

func (s *Stress) RunStressContainer(name string, mode string, duration int, iterationIndex int, index int, env map[string]string) error {
	env_vars := []string{
		fmt.Sprintf("threads=%v", consts.THREADS),
		fmt.Sprintf("concurrent=%v", consts.CONCURRENT),
		fmt.Sprintf("TZ=%v", consts.TZ),
		fmt.Sprintf("dt_format=%v", consts.DT_FORMAT),
		fmt.Sprintf("delay=%v", consts.DELAY),
		fmt.Sprintf("test_id=%v", consts.TestID),
		fmt.Sprintf("ssid=%v", s.Config.Router.Ssid),
		fmt.Sprintf("protocol=%v", s.Config.Router.Name),
		fmt.Sprintf("iteration_index=%v", iterationIndex),
		fmt.Sprintf("index=%v", index),
		fmt.Sprintf("name=%v", name),
	}

	for key, value := range env {
		env_vars = append(env_vars, fmt.Sprintf("%v=%v", key, value))
	}

	ip, err := s.DHCPServer.Lease()

	if err != nil {
		return err
	}

	logger.Logger.Debug(fmt.Sprintf("starting container %v", name))
	container, err := s.Docker.CreateAndStartContainer(name, mode, env_vars, ip)

	if err != nil {
		return err
	}

	time.Sleep(time.Duration(duration) * time.Minute)

	s.DHCPServer.Release(ip)

	return s.Docker.StopContainer(container, consts.SIGINT)
}

func setupSlave(stress *Stress) error {
	slave, err := router.NewSlave(stress.Config)
	stress.Slave = slave

	if err != nil {
		logger.Logger.Error(err.Error())
		return err
	}

	err = slave.TransferSamplerToRouter()

	if err != nil {
		logger.Logger.Error(err.Error())
		return err
	}

	err = slave.StartSampler()

	if err != nil {
		logger.Logger.Error(err.Error())
	}

	return err
}

func setupDocker(stress *Stress) error {
	docker, err := docker.NewDocker(stress.Config)

	if err != nil {
		return err
	}

	stress.Docker = docker

	return docker.BuildImages(stress.Config)
}

func setupDHCP(stress *Stress) error {
	dhcp := dhcp.NewDHCPServer(stress.Config)
	dhcp.PopulatePool()

	stress.DHCPServer = dhcp

	return nil
}

func (s *Stress) Cleanup() error {
	var eg errgroup.Group

	logger.Logger.Debug("Running cleanup")

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

	err := os.Mkdir(dirName, 0666)

	if err != nil {
		return err
	}

	return err

	// uid, gid, err := GetCurrentUser()

	// if err != nil {
	// 	return err
	// }

	// return os.Chown(dirName, uid, gid)
}

func GetCurrentUser() (int, int, error) {
	user, err := user.Current()

	if err != nil {
		return 0, 0, err
	}

	userID, err := strconv.Atoi(user.Uid)

	if err != nil {
		return 0, 0, err
	}

	userGID, err := strconv.Atoi(user.Gid)

	return userID, userGID, err
}
