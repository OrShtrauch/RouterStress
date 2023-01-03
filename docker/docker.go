package docker

import (
	"RouterStress/conf"
	"RouterStress/consts"
	"fmt"
	"io"
	"os"

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

	err = docker.Cleanup()

	if err != nil {
		return &docker, err
	}

	err = docker.BuildImages(config)

	if err != nil {
		return &docker, err
	}

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
			"parent": config.Network.Parent,
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
		fmt.Sprintf("test_id=%v", data.TestID),
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
	imageName := fmt.Sprintf("%v-%v:%v", consts.STRESS_CONTAINER_PREFIX, mode, consts.CONTAINER_VERSION)

	workingDir, err := os.Getwd()

	if err != nil {
		return nil, err
	}

	localPath := fmt.Sprintf("%v/%v/%v", workingDir, consts.RESULTS_DIR, consts.TEST_ID)

	binds := []string{fmt.Sprintf("%v:%v", localPath, consts.REMOTE_VOLUME_PATH)}

	return d.Client.CreateContainer(dockerlib.CreateContainerOptions{
		Name: name,
		Config: &dockerlib.Config{
			Image: imageName,
			Env:   env,
			//Cmd:   []string{"sleep", "500"},
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
	fmt.Print("killing container\n")
	return d.Client.KillContainer(dockerlib.KillContainerOptions{
		ID:     c.ID,
		Signal: consts.SIGTERM,
	})
}

func (d *Docker) BuildImages(config *conf.Config) error {
	var eg errgroup.Group

	eg.Go(func() error {
		return d.buildTrafficCaptureImage()
	})

	for _, s := range config.Scenarios.Scenarios {
		scenario := s

		eg.Go(func() error {
			return d.buildImage(scenario.Name)
		})
	}

	return eg.Wait()
}

func (d *Docker) buildImage(mode string) error {
	imageName := fmt.Sprintf("stress-%v:%v", mode, consts.CONTAINER_VERSION)
	dockerFile := fmt.Sprintf("Dockerfile.%v", mode)

	return d.Client.BuildImage(dockerlib.BuildImageOptions{
		Name:         imageName,
		Dockerfile:   dockerFile,
		ContextDir:   "containers/",
		OutputStream: io.Discard,
	})
}

func (d *Docker) buildTrafficCaptureImage() error {
	imageName := fmt.Sprintf("%v:%v", consts.TRAFFIC_CAPTURE_IMAGE_NAME, consts.CONTAINER_VERSION)

	return d.Client.BuildImage(dockerlib.BuildImageOptions{
		Name:         imageName,
		ContextDir:   consts.TRAFFIC_CAPTURE_PATH,
		OutputStream: io.Discard,
	})
}

func (d *Docker) StartTrafficCaptureContainer() (*dockerlib.Container, error) {
	c, err := d.createTrafficCaptureContainer()

	if err != nil {
		return c, err
	}

	return c, d.startContainer(c)
}

func (d *Docker) createTrafficCaptureContainer() (*dockerlib.Container, error) {
	name := fmt.Sprintf("traffic-capture-%v", consts.TEST_UUID[:5])
	imageName := fmt.Sprintf("%v:%v", consts.TRAFFIC_CONTAINER_PREFIX, consts.CONTAINER_VERSION)

	env := []string{
		fmt.Sprintf("URL=%v", consts.TRAFFIC_CAPTURE_URL),
		fmt.Sprintf("SOCKET=%v", consts.TRAFFIC_UNIX_SOCKET),
		fmt.Sprintf("SLEEP=%v", consts.DELAY),
	}

	//workingDir, err := os.Getwd()

	// if err != nil {
	// 	return nil, err
	// }

	//localPath := fmt.Sprintf("%v/%v", workingDir, consts.RESULTS_DIR)
	binds := []string{
		//fmt.Sprintf("%v:%v", localPath, consts.REMOTE_VOLUME_PATH),
		"/tmp:/tmp",
	}

	return d.Client.CreateContainer(dockerlib.CreateContainerOptions{
		Name: name,
		Config: &dockerlib.Config{
			Image: imageName,
			Env:   env,
		},
		HostConfig: &dockerlib.HostConfig{
			Binds: binds,
		},
	})
}

func (d *Docker) Cleanup() error {
	var err error

	_, err = d.Client.PruneContainers(dockerlib.PruneContainersOptions{})

	if err != nil {
		return err
	}

	_, err = d.Client.PruneNetworks(dockerlib.PruneNetworksOptions{})

	if err != nil {
		return err
	}

	_, err = d.Client.PruneImages(dockerlib.PruneImagesOptions{})

	if err != nil {
		return err
	}

	_, err = d.Client.PruneVolumes(dockerlib.PruneVolumesOptions{})

	return err
}
