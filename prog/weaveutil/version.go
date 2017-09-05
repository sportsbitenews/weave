package main

import (
	"bytes"
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
)

func askVersion(args []string) error {
	if len(args) < 1 {
		cmdUsage("ask-version", "<container-name-or-id> <image-name-or-id>")
	}
	containerID := args[0]
	imageID := args[1]

	c, err := docker.NewVersionedClientFromEnv("1.18")
	if err != nil {
		return fmt.Errorf("unable to connect to docker: %s", err)
	}

	var image string
	container, err := c.InspectContainer(containerID)
	if err != nil {
		img, err := c.InspectImage(imageID)
		if err != nil {
			return fmt.Errorf("unable to find image %s: %s", imageID, err)
		} else {
			image = img.ID
		}
	} else {
		image = container.Image
	}

	//
	// TODO: Move this code to a helper in container.go for reuse with other `docker run` commands
	//
	config := docker.Config{Image: image, Env: []string{"WEAVE_CIDR=none"}, Cmd: []string{"--version"}}
	hostConfig := docker.HostConfig{AutoRemove: true, NetworkMode: "none"}
	versionContainer, err := c.CreateContainer(docker.CreateContainerOptions{Config: &config, HostConfig: &hostConfig})
	if err != nil {
		return fmt.Errorf("unable to create container: %s", err)
	}

	err = c.StartContainer(versionContainer.ID, &hostConfig)
	if err != nil {
		return fmt.Errorf("unable to start container: %s", err)
	}

	var buf bytes.Buffer
	err = c.AttachToContainer(docker.AttachToContainerOptions{Container: versionContainer.ID, OutputStream: &buf, Stdout: true, Stream: true})
	if err != nil {
		return fmt.Errorf("unable to attach to container: %s", err)
	}

	// AutoRemove doesn't work with old Docker daemons (< 1.25), manually remove
	err = c.RemoveContainer(docker.RemoveContainerOptions{ID: versionContainer.ID, Force: true})
	if err != nil {
		return fmt.Errorf("unable to remove container: %s", err)
	}

	fmt.Print(buf.String())
	return nil
}
