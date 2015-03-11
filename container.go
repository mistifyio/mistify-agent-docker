package mdocker

import (
	"net/http"

	"github.com/fsouza/go-dockerclient"
)

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

// ListContainers retrieves a list of Docker containers
func (md *MDocker) ListContainers(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	return nil
}

// GetContainer retrieves information about a specific Docker container
func (md *MDocker) GetContainer(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	return nil
}

// DeleteContainer deletes a Docker container
func (md *MDocker) DeleteContainer(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	return nil
}

// SaveContainer saves a Docker container
func (md *MDocker) SaveContainer(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	return nil
}

// StartContainer starts a Docker container
func (md *MDocker) StartContainer(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	return nil
}

// StopContainer stop a Docker container or kills it after a timeout
func (md *MDocker) StopContainer(h *http.Request, request *ContainerRequest, response *ContainerResponse) error {
	return nil
}
