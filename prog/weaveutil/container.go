package main

import (
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
)

func stopContainer(args []string) error {
	if len(args) < 1 {
		cmdUsage("stop-container", "<container-id> [<container-id2> ...]")
	}

	c, err := docker.NewVersionedClientFromEnv("1.18")
	if err != nil {
		return fmt.Errorf("unable to connect to docker: %s", err)
	}

	for _, containerID := range args {
		err = c.StopContainer(containerID, 10)
		if err != nil {
			return fmt.Errorf("unable to stop container %s: %s", containerID, err)
		}
	}
	return nil
}
