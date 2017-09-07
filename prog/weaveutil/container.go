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

func runContainer(args []string) error {
	env := []string{}
	name := ""
	net := ""
	pid := ""
	privileged := false
	restart := docker.NeverRestart()
	volumes := []string{}
	volumesFrom := []string{}

	done := false
	for i := 0; i < len(args) && !done; {
		switch args[i] {
		case "-e", "--env":
			env = append(env, args[i+1])
			args = append(args[:i], args[i+2:]...)
		case "--name":
			name = args[i+1]
			args = append(args[:i], args[i+2:]...)
		case "--net":
			net = args[i+1]
			args = append(args[:i], args[i+2:]...)
		case "--pid":
			pid = args[i+1]
			args = append(args[:i], args[i+2:]...)
		case "--privileged":
			privileged = true
			args = append(args[:i], args[i+1:]...)
		case "--restart":
			restart = docker.RestartPolicy{Name: args[i+1]}
			args = append(args[:i], args[i+2:]...)
		case "-v", "--volume":
			// Must dedup binds, otherwise, container fails creation
			skip := false
			for _, v := range volumes {
				if v == args[i+1] {
					skip = true
				}
			}
			if !skip {
				volumes = append(volumes, args[i+1])
			}
			args = append(args[:i], args[i+2:]...)
		case "--volumes-from":
			volumesFrom = append(volumesFrom, args[i+1])
			args = append(args[:i], args[i+2:]...)
		default:
			done = true
		}
	}

	if len(args) < 2 {
		cmdUsage("run-container", `[options] <image> <cmd> [[<cmd-options>] <cmd-arg1> [<cmd-arg2> ...]]

  -e, --env list                       Set environment variables
      --name string                    Assign a name to the container
      --net string                     Network Mode
      --pid string                     PID namespace to use
      --privileged                     Give extended privileges to this container
      --restart string                 Restart policy to apply when a container exits (default "no")
  -v, --volume list                    Bind mount a volume
      --volumes-from list              Mount volumes from the specified container(s)
`)
	}

	image := args[0]
	cmds := args[1:]

	c, err := docker.NewVersionedClientFromEnv("1.18")
	if err != nil {
		return fmt.Errorf("unable to connect to docker: %s", err)
	}

	config := docker.Config{Image: image, Env: env, Cmd: cmds}
	hostConfig := docker.HostConfig{NetworkMode: net, PidMode: pid, Privileged: privileged, RestartPolicy: restart, Binds: volumes, VolumesFrom: volumesFrom}
	container, err := c.CreateContainer(docker.CreateContainerOptions{Name: name, Config: &config, HostConfig: &hostConfig})
	if err != nil {
		return fmt.Errorf("unable to create container: %s", err)
	}

	err = c.StartContainer(container.ID, &hostConfig)
	if err != nil {
		return fmt.Errorf("unable to start container: %s", err)
	}

	fmt.Print(container.ID)
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

	containers, err := c.ListContainers(docker.ListContainersOptions{All: true, Filters: map[string][]string{"label": {label}}})
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
