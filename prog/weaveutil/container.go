package main

import (
	"fmt"
	"regexp"

	docker "github.com/fsouza/go-dockerclient"
)

func containerID(args []string) error {
	if len(args) < 1 {
		cmdUsage("container-id", "<container-name-or-short-id>")
	}
	containerID := args[0]

	c, err := docker.NewVersionedClientFromEnv("1.18")
	if err != nil {
		return fmt.Errorf("unable to connect to docker: %s", err)
	}

	container, err := c.InspectContainer(containerID)
	if err != nil {
		return fmt.Errorf("unable to inspect container %s: %s", containerID, err)
	}
	fmt.Print(container.ID)
	return nil
}

func containerState(args []string) error {
	if len(args) < 1 {
		cmdUsage("container-state", "<container-id> [<image-name-or-id>]")
	}
	containerID := args[0]

	c, err := docker.NewVersionedClientFromEnv("1.18")
	if err != nil {
		return fmt.Errorf("unable to connect to docker: %s", err)
	}

	container, err := c.InspectContainer(containerID)
	if err != nil {
		return fmt.Errorf("unable to inspect container %s: %s", containerID, err)
	}

	if len(args) == 1 {
		fmt.Print(container.State.StateString())
	} else {
		image := args[1]
		match1, _ := regexp.MatchString(image, container.Image)
		match2, _ := regexp.MatchString(image, container.Config.Image)
		if match1 || match2 {
			fmt.Print(container.State.StateString())
		} else {
			if container.State.Running {
				fmt.Print("running image mismatch: ", container.Config.Image)
			} else {
				fmt.Print("image mismatch: ", container.Config.Image)
			}
		}
	}
	return nil
}

func containerFQDN(args []string) error {
	if len(args) < 1 {
		cmdUsage("container-fqdn", "<container-id>")
	}
	containerID := args[0]

	c, err := docker.NewVersionedClientFromEnv("1.18")
	if err != nil {
		return fmt.Errorf("unable to connect to docker: %s", err)
	}

	container, err := c.InspectContainer(containerID)
	if err != nil {
		return fmt.Errorf("unable to inspect container %s: %s", containerID, err)
	}

	fmt.Print(container.Config.Hostname, ".", container.Config.Domainname)
	return nil
}

func listContainers(args []string) error {
	if len(args) < 1 {
		cmdUsage("list-containers", "[<label>]")
	}
	label := args[0]

	c, err := docker.NewVersionedClientFromEnv("1.18")
	if err != nil {
		return fmt.Errorf("unable to connect to docker: %s", err)
	}

	containers, err := c.ListContainers(docker.ListContainersOptions{All: true, Filters: map[string][]string{"label": []string{label}}})
	if err != nil {
		return fmt.Errorf("unable to list containers by label %s: %s", label, err)
	}

	for _, container := range containers {
		fmt.Println(container.ID)
	}
	return nil
}

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

func killContainer(args []string) error {
	if len(args) < 1 {
		cmdUsage("kill-container", "<container-id> [<container-id2> ...]")
	}

	c, err := docker.NewVersionedClientFromEnv("1.18")
	if err != nil {
		return fmt.Errorf("unable to connect to docker: %s", err)
	}

	for _, containerID := range args {
		err = c.KillContainer(docker.KillContainerOptions{ID: containerID, Signal: docker.SIGKILL})
		if err != nil {
			return fmt.Errorf("unable to stop container %s: %s", containerID, err)
		}
	}
	return nil
}

func removeContainer(args []string) error {
	if len(args) < 1 {
		cmdUsage("remove-container", "[-f | --force]  [-v | --volumes] <container-id> [<container-id2> ...]")
	}

	force := false
	volumes := false
	for i := 0; i < len(args); {
		switch args[i] {
		case "--force":
		case "-f":
			force = true
			args = append(args[:i], args[i+1:]...)
		case "--volumes":
		case "-v":
			volumes = true
			args = append(args[:i], args[i+1:]...)
		default:
			i++
		}
	}

	c, err := docker.NewVersionedClientFromEnv("1.18")
	if err != nil {
		return fmt.Errorf("unable to connect to docker: %s", err)
	}

	for _, containerID := range args {
		err = c.RemoveContainer(docker.RemoveContainerOptions{
			ID: containerID, Force: force, RemoveVolumes: volumes})
		if err != nil {
			return fmt.Errorf("unable to stop container %s: %s", containerID, err)
		}
	}
	return nil
}
