package mdocker

import (
	"net/http"

	"github.com/fsouza/go-dockerclient"
)

// FakeContainer is a hardcoded container for stubbing purposes
// TODO: Remove once everything is built out
var FakeContainer = docker.APIContainers{
	ID:      "2f07b3c86f592d4c3ae6c45c058bff080490d8c62ce66530eb187b2b44f1c997",
	Image:   "centos:7",
	Command: "/bin/bash",
	Created: 1426187284,
	Status:  "Exited (0) 55 seconds ago",
	Names: []string{
		"/focused_brattain",
	},
}

type (
	// ContainerRequest is a container request to the Docker sub-agent
	ContainerRequest struct {
		ID   string      `json:"id"`   // Container ID
		Opts interface{} `json:"opts"` // Generic Options. Will need converting
	}

	// ContainerResponse is a container response from the Docker sub-agent
	ContainerResponse struct {
		Containers []docker.APIContainers `json:"containers"` // Slice of one or more containers
	}
)

// GetOpts returns the Opts property
func (cs *ContainerRequest) GetOpts() interface{} {
	return cs.Opts
}

// ListContainers retrieves a list of Docker containers
func (md *MDocker) ListContainers(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	opts := &docker.ListContainersOptions{}
	if err := md.RequestOpts(request, opts); err != nil {
		return err
	}

	containers, err := md.client.ListContainers(*opts)
	if err != nil {
		return err
	}
	response.Containers = containers
	return nil
}

// GetContainer retrieves information about a specific Docker container
func (md *MDocker) GetContainer(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	response.Containers = []docker.APIContainers{
		FakeContainer,
	}
	return nil
}

// DeleteContainer deletes a Docker container
func (md *MDocker) DeleteContainer(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	response.Containers = []docker.APIContainers{
		FakeContainer,
	}
	return nil
}

// SaveContainer saves a Docker container
func (md *MDocker) SaveContainer(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	response.Containers = []docker.APIContainers{
		FakeContainer,
	}
	return nil
}

// StartContainer starts a Docker container
func (md *MDocker) StartContainer(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	response.Containers = []docker.APIContainers{
		FakeContainer,
	}
	return nil
}

// StopContainer stop a Docker container or kills it after a timeout
func (md *MDocker) StopContainer(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	response.Containers = []docker.APIContainers{
		FakeContainer,
	}
	return nil
}
