package mdocker

import (
	"errors"
	"net/http"

	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent/rpc"
)

const (
	cStateRunning = "running"
	cStatePaused  = "paused"
	cStateStopped = "stopped"
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

func (md *MDocker) fetchContainerState(containerID string) (string, error) {
	container, err := md.client.InspectContainer(containerID)
	if err != nil {
		return "", err
	}
	if container.State.Paused {
		return cStatePaused, nil
	}
	if container.State.Running {
		return cStateRunning, nil
	}
	return cStateStopped, nil
}

func assertContainerState(expected, actual string) error {
	if actual != expected {
		return errors.New("unexpected container state")
	}
	return nil
}

func requestContainerName(request *rpc.GuestRequest) (string, error) {
	if request.Guest == nil || request.Guest.Id == "" {
		return "", errors.New("missing guest with id")
	}
	return request.Guest.Id, nil
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
func (md *MDocker) DeleteContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	containerName, err := requestContainerName(request)
	if err != nil {
		return err
	}

	opts := docker.RemoveContainerOptions{
		ID: containerName,
	}
	if err := md.client.RemoveContainer(opts); err != nil {
		return err
	}

	response.Guest = request.Guest
	response.Guest.State = "deleted"
	return nil
}

// SaveContainer saves a Docker container
func (md *MDocker) SaveContainer(h *http.Request, request *rpc.ContainerRequest, response *rpc.ImageResponse) error {
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
	response.Images = []*rpc.Image{
		&rpc.Image{
			Id:   image.ID,
			Type: "container",
			Size: uint64(image.Size),
		},
	}
	return nil
}

// CreateContainer creates a new Docker container
func (md *MDocker) CreateContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	containerName, err := requestContainerName(request)
	if err != nil {
		return err
	}
	guest := request.Guest
	opts := docker.CreateContainerOptions{
		Name: containerName,
		Config: &docker.Config{
			Hostname: guest.Id,
			Image:    guest.Image,
			Memory:   int64(guest.Memory) * 1024 * 1024, // Convert MB to bytes
		},
		HostConfig: &docker.HostConfig{
			PublishAllPorts: true,
		},
	}
	container, err := md.client.CreateContainer(opts)
	if err != nil {
		return err
	}

	state, err := md.fetchContainerState(container.ID)
	if err != nil {
		return err
	}

	guest.State = state
	response.Guest = guest
	return nil
}

// StartContainer starts a Docker container
func (md *MDocker) StartContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	containerName, err := requestContainerName(request)
	if err != nil {
		return err
	}
	hostConfig := &docker.HostConfig{
		PublishAllPorts: true,
	}
	err = md.client.StartContainer(containerName, hostConfig)
	if err, ok := err.(*docker.ContainerAlreadyRunning); err != nil && !ok {
		return err
	}
	state, err := md.fetchContainerState(containerName)
	if err != nil {
		return err
	}
	if err := assertContainerState(cStateRunning, state); err != nil {
		return err
	}

	response.Guest = request.Guest
	response.Guest.State = state
	return nil
}

// StopContainer stops a Docker container or kills it after a timeout
func (md *MDocker) StopContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestRequest) error {
	containerName, err := requestContainerName(request)
	if err != nil {
		return err
	}
	err = md.client.StopContainer(containerName, 10)
	if err, ok := err.(*docker.ContainerNotRunning); err != nil && !ok {
		return err
	}
	state, err := md.fetchContainerState(containerName)
	if err != nil {
		return err
	}
	if err := assertContainerState(cStateStopped, state); err != nil {
		return err
	}

	response.Guest = request.Guest
	response.Guest.State = state
	return nil
}

// RestartContainer restarts a Docker container
func (md *MDocker) RestartContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestRequest) error {
	containerName, err := requestContainerName(request)
	if err != nil {
		return err
	}
	if err := md.client.RestartContainer(containerName, 60); err != nil {
		return err
	}
	state, err := md.fetchContainerState(containerName)
	if err != nil {
		return err
	}
	if err := assertContainerState(cStateRunning, state); err != nil {
		return err
	}

	response.Guest = request.Guest
	response.Guest.State = state
	return nil
}

// RebootContainer restarts a Docker container
func (md *MDocker) RebootContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestRequest) error {
	return md.RestartContainer(h, request, response)
}

// PauseContainer pauses a Docker container
func (md *MDocker) PauseContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestRequest) error {
	containerName, err := requestContainerName(request)
	if err != nil {
		return err
	}
	if err := md.client.PauseContainer(containerName); err != nil {
		return err
	}
	state, err := md.fetchContainerState(containerName)
	if err != nil {
		return err
	}
	if err := assertContainerState(cStatePaused, state); err != nil {
		return err
	}

	response.Guest = request.Guest
	response.Guest.State = state
	return nil
}

// UnpauseContainer restarts a Docker container
func (md *MDocker) UnpauseContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestRequest) error {
	containerName, err := requestContainerName(request)
	if err != nil {
		return err
	}
	if err := md.client.UnpauseContainer(containerName); err != nil {
		return err
	}
	state, err := md.fetchContainerState(containerName)
	if err != nil {
		return err
	}
	if err := assertContainerState(cStateRunning, state); err != nil {
		return err
	}

	response.Guest = request.Guest
	response.Guest.State = state
	return nil
}
