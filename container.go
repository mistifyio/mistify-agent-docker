package mdocker

import (
	"net/http"

	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent/rpc"
)

func (md *MDocker) containersFromAPIContainers(acs []docker.APIContainers) ([]*docker.Container, error) {
	containers := make([]*docker.Container, 0, len(acs))
	for _, ac := range acs {
		container, err := md.client.InspectContainer(ac.ID)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}
	return containers, nil
}

// ListContainers retrieves a list of Docker containers
func (md *MDocker) ListContainers(h *http.Request, request *rpc.ContainerRequest, response *rpc.ContainerResponse) error {
	var opts docker.ListContainersOptions
	if err := md.RequestOpts(request, &opts); err != nil {
		return err
	}

	apiContainers, err := md.client.ListContainers(opts)
	if err != nil {
		return err
	}
	containers, err := md.containersFromAPIContainers(apiContainers)
	if err != nil {
		return err
	}

	response.Containers = containers
	return nil
}

// GetContainer retrieves information about a specific Docker container
func (md *MDocker) GetContainer(h *http.Request, request *rpc.ContainerRequest, response *rpc.ContainerResponse) error {
	container, err := md.client.InspectContainer(request.ID)
	if err != nil {
		return err
	}

	response.Containers = []*docker.Container{
		container,
	}
	return nil
}

// DeleteContainer deletes a Docker container
func (md *MDocker) DeleteContainer(h *http.Request, request *rpc.ContainerRequest, response *rpc.ContainerResponse) error {
	container, err := md.client.InspectContainer(request.ID)
	if err != nil {
		return err
	}

	var opts docker.RemoveContainerOptions
	if err := md.RequestOpts(request, &opts); err != nil {
		return err
	}
	if opts.ID == "" {
		opts.ID = container.ID
	}
	if err := md.client.RemoveContainer(opts); err != nil {
		return err
	}
	response.Containers = []*docker.Container{
		container,
	}
	return nil
}

// SaveContainer saves a Docker container
func (md *MDocker) SaveContainer(h *http.Request, request *rpc.ContainerRequest, response *rpc.ContainerImageResponse) error {
	var opts docker.CommitContainerOptions
	if err := md.RequestOpts(request, &opts); err != nil {
		return err
	}
	if request.ID != "" {
		opts.Container = request.ID
	}
	image, err := md.client.CommitContainer(opts)
	if err != nil {
		return err
	}
	response.Images = []*docker.Image{
		image,
	}
	return nil
}

// CreateContainer creates a new Docker container
func (md *MDocker) CreateContainer(h *http.Request, request *rpc.ContainerRequest, response *rpc.ContainerResponse) error {
	opts := docker.CreateContainerOptions{}
	if err := md.RequestOpts(request, &opts); err != nil {
		return nil
	}
	container, err := md.client.CreateContainer(opts)
	if err != nil {
		return err
	}
	response.Containers = []*docker.Container{
		container,
	}
	return nil
}

// StartContainer starts a Docker container
func (md *MDocker) StartContainer(h *http.Request, request *rpc.ContainerRequest, response *rpc.ContainerResponse) error {
	hostConfig := &docker.HostConfig{}
	if err := md.RequestOpts(request, hostConfig); err != nil {
		return err
	}
	if err := md.client.StartContainer(request.ID, hostConfig); err != nil {
		return err
	}

	container, err := md.client.InspectContainer(request.ID)
	if err != nil {
		return err
	}
	response.Containers = []*docker.Container{
		container,
	}
	return nil
}

// StopContainer stop a Docker container or kills it after a timeout
func (md *MDocker) StopContainer(h *http.Request, request *rpc.ContainerRequest, response *rpc.ContainerResponse) error {
	if err := md.client.StopContainer(request.ID, 60); err != nil {
		return err
	}

	container, err := md.client.InspectContainer(request.ID)
	if err != nil {
		return err
	}
	response.Containers = []*docker.Container{
		container,
	}
	return nil
}
