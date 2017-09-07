package main

import (
	"bytes"
	"fmt"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

func pullImage(args []string) error {
	if len(args) < 1 {
		cmdUsage("pullImage", "<image-name>[:tag>]")
	}
	image := ""
	tag := ""
	imageParts := strings.Split(args[0], ":")
	if len(imageParts) == 2 {
		image, tag = imageParts[0], imageParts[1]
	} else {
		image = imageParts[0]
		tag = "latest"
		fmt.Println("Using default tag: " + tag)
	}
	fmt.Println(tag + ": Pulling from " + image)

	c, err := docker.NewVersionedClientFromEnv("1.18")
	if err != nil {
		return fmt.Errorf("unable to connect to docker: %s", err)
	}

	var buf bytes.Buffer
	err = c.PullImage(docker.PullImageOptions{Repository: image, Tag: tag, OutputStream: &buf}, docker.AuthConfiguration{})
	if err != nil {
		return fmt.Errorf("unable to pull image %s: %s", image, err)
	}
	return nil
}
