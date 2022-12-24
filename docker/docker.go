package docker

import (
	"fmt"

	dockerLib "github.com/fsouza/go-dockerclient"
)

var client *dockerLib.Client

func Init() error {
	var err error

	client, err = dockerLib.NewClientFromEnv()

	return err
}

func Test() {
	imgs, _ := client.ListImages(dockerLib.ListImagesOptions{All: false})

	for _, img := range imgs {
		fmt.Println("RepoTags: ", img.RepoTags)
	}
}
