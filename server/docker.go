package main

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

var cli *client.Client
var err error

func init_docker() error {
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return err
}

func getDockerHosts() (string, error) {
	header := "## Platform START\n"
	end := "## Platform END\n"
	k8sHosts := header

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return k8sHosts, err
	}

	for _, container := range containers {
		containerInfo, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			log.Fatalf("Error inspecting container: %v", err)
		}
		for _, network := range containerInfo.NetworkSettings.Networks {
			for _, alias := range network.Aliases {
				k8sHosts += fmt.Sprintf("%s\t%s\n", network.IPAddress, alias)
			}
		}
	}

	k8sHosts += end
	return k8sHosts, err
}

func getDockerNet() (string, error) {
	networks, err := cli.NetworkList(context.Background(), network.ListOptions{})
	if err != nil {
		return "", err
	}
	nets := ""
	for _, network := range networks {
		for _, config := range network.IPAM.Config {
			if config.Subnet != "" {
				nets += config.Subnet + "\n"
			}
		}
	}

	return nets, nil
}
