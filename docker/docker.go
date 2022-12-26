package docker

import (
	"RouterStress/conf"
	"RouterStress/consts"
	"fmt"

	dockerlib "github.com/fsouza/go-dockerclient"
	"golang.org/x/sync/errgroup"
)

type ContainerData struct {
	Ssid           string
	Platform       string
	Mode           string
	Name           string
	Ip             string
	TestID         string
	Index          int
	IterationIndex int
	RunIndex       int
	Params         map[string]string
}

type Docker struct {
	Network *dockerlib.Network
	Client  *dockerlib.Client
}

func InitDocker(config *conf.Config) (*Docker, error) {
	var docker Docker
	var err error

	client, err := dockerlib.NewClientFromEnv()

	if err != nil {
		return &docker, err
	}

	docker.Client = client

	network, err := docker.createMacvlan(config)

	if err != nil {
		return &docker, err
	}

	docker.Network = network

	return &docker, err
}

func (d *Docker) createMacvlan(config *conf.Config) (*dockerlib.Network, error) {
	return d.Client.CreateNetwork(dockerlib.CreateNetworkOptions{
		Name:   consts.MACVLAN,
		Driver: consts.MACVLAN,
		Options: map[string]interface{}{
			"parent": config.Network.Interface,
		},
		IPAM: &dockerlib.IPAMOptions{
			Config: []dockerlib.IPAMConfig{
				{
					Subnet:  config.Router.GetSubnet(),
					Gateway: config.Router.Ip,
				},
			},
		},
	})
}

func (d *Docker) DeleteMacvlan() error {
	return d.Client.RemoveNetwork(d.Network.ID)
}

func (d *Docker) RunContainer(data ContainerData) (*dockerlib.Container, error) {
	var c *dockerlib.Container
	var err error

	env := []string{
		fmt.Sprintf("ssid=%v", data.Ssid),
		fmt.Sprintf("platform=%v", data.Platform),
		fmt.Sprintf("name=%v", data.Name),
		fmt.Sprintf("mode=%v", data.Mode),
		fmt.Sprintf("index=%v", data.Index),
		fmt.Sprintf("iteration_index=%v", data.IterationIndex),
		fmt.Sprintf("run_index=%v", data.RunIndex),
		fmt.Sprintf("threads=%v", consts.THREADS),
		fmt.Sprintf("concurrent=%v", consts.CONCURRENT),
		fmt.Sprintf("TZ=%v", consts.TZ),
		fmt.Sprintf("dt_format=%v", consts.DT_FORMAT),
		fmt.Sprintf("delay=%v", consts.DELAY),
	}

	for key, value := range data.Params {
		env = append(env, fmt.Sprintf("%v=%v", key, value))
	}

	c, err = d.createContainer(data.Name, data.Mode, env)

	if err != nil {
		return c, err
	}

	err = d.connectContainerToNetwork(c, data.Ip)

	if err != nil {
		return c, err
	}

	return c, d.startContainer(c)
}

func (d *Docker) createContainer(name string, mode string, env []string) (*dockerlib.Container, error) {
	imageName := fmt.Sprintf("stress-%v:%v", mode, consts.CONTAINER_VERSION)
	binds := []string{fmt.Sprintf("%v:%v", consts.LOCAL_VOLUME_PATH, consts.REMOTE_VOLUME_PATH)}

	return d.Client.CreateContainer(dockerlib.CreateContainerOptions{
		Name: name,
		Config: &dockerlib.Config{
			Image: imageName,
			Env:   env,
			// Cmd: []string{"sleep", "500"},
		},
		HostConfig: &dockerlib.HostConfig{
			Binds: binds,
		},
	})
}

func (d *Docker) connectContainerToNetwork(c *dockerlib.Container, ip string) error {
	return d.Client.ConnectNetwork(d.Network.ID, dockerlib.NetworkConnectionOptions{
		Container: c.ID,
	})
}

func (d *Docker) startContainer(c *dockerlib.Container) error {
	return d.Client.StartContainer(c.ID, nil)
}

func (d *Docker) KillContainer(c *dockerlib.Container) error {
	return d.Client.KillContainer(dockerlib.KillContainerOptions{
		ID:     c.ID,
		Signal: consts.SIGTERM,
	})
}

func (d *Docker) BuildImages(config *conf.Config) error {
	var eg errgroup.Group

	for _, s := range config.Scenarios {
		scenario := s

		eg.Go(func() error  {
			return d.buildImage(scenario.Name)
		})
	}

	return eg.Wait()
}

func (d *Docker) buildImage(mode string) error {
	imageName := fmt.Sprintf("stress-%v:%v", mode , consts.CONTAINER_VERSION)
	dockerFile := fmt.Sprintf("stress.%v", mode)

	return d.Client.BuildImage(dockerlib.BuildImageOptions{
		Name: imageName,
		Dockerfile: dockerFile,
	})
}