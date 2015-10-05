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
	if request.Guest == nil || request.Guest.ID == "" {
		return "", errors.New("missing guest with id")
	}
	return request.Guest.ID, nil
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
			ID:   image.ID,
			Type: "container",
			Size: uint64(image.Size) / 1024 / 1024,
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

	if len(guest.Nics) == 0 {
		return errors.New("must specify at least one nic")
	}

	// TODO: Some of these options might be better handled as guest metadata
	// instead of hardcoding, such as the openstdin option, port forwarding,
	// and zfs devices
	opts := docker.CreateContainerOptions{
		Name: containerName,
		Config: &docker.Config{
			Hostname:   guest.ID,
			Image:      guest.Image,
			OpenStdin:  true,
			MacAddress: guest.Nics[0].Mac,
			Memory:     int64(guest.Memory) * 1024 * 1024, // Convert MB to bytes
		},
		HostConfig: &docker.HostConfig{
			// A network interface will be added separately. The "none" option
			// may not be listed in the docker remote api docs, but it works
			NetworkMode: "none",
			// Expose /dev/zfs inside all containers (okay because they are unprivileged)
			Devices: []docker.Device{
				docker.Device{
					PathOnHost:        "/dev/zfs",
					PathInContainer:   "/dev/zfs",
					CgroupPermissions: "rwm",
				},
			},
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
	// Make sure there are no lingering interfaces for the guest from a previous
	// run
	if err := removeInterfaces(request.Guest); err != nil {
		return err
	}

	containerName, err := requestContainerName(request)
	if err != nil {
		return err
	}
	err = md.client.StartContainer(containerName, nil)
	if _, ok := err.(*docker.ContainerAlreadyRunning); err != nil && !ok {
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

	if err := addInterfaces(request.Guest); err != nil {
		return err
	}

	return nil
}

// StopContainer stops a Docker container or kills it after a timeout
func (md *MDocker) StopContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	containerName, err := requestContainerName(request)
	if err != nil {
		return err
	}
	err = md.client.StopContainer(containerName, 10)
	if _, ok := err.(*docker.ContainerNotRunning); err != nil && !ok {
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

	// The virtual interfaces are destroyed when the container stops, but are
	// still being tracked in OVS. Clean things up.
	if err := removeInterfaces(request.Guest); err != nil {
		return err
	}

	return nil
}

// RestartContainer restarts a Docker container
func (md *MDocker) RestartContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	// Since action needs to be taken relating to network interfaces when the guest
	// is stopped and again when the guest is started, the individual methods are
	// used here. The `docker restart` command is an alias for a stop then start
	// anyway.
	if err := md.StopContainer(h, request, response); err != nil {
		return err
	}
	if err := md.StartContainer(h, request, response); err != nil {
		return err
	}

	return nil
}

// RebootContainer restarts a Docker container
func (md *MDocker) RebootContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
	return md.RestartContainer(h, request, response)
}

// PauseContainer pauses a Docker container
func (md *MDocker) PauseContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
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
func (md *MDocker) UnpauseContainer(h *http.Request, request *rpc.GuestRequest, response *rpc.GuestResponse) error {
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
