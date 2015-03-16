package mdocker_test

import (
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent-docker"
)

func TestCreateContainer(t *testing.T) {
	ireq := &mdocker.ImageRequest{
		Name: client.ImageName,
	}
	iresp := &mdocker.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.PullImage", ireq, iresp))
	client.ImageID = iresp.Images[0].ID

	req := &mdocker.ContainerRequest{
		Opts: &docker.CreateContainerOptions{
			Config: &docker.Config{
				Image: client.ImageName,
			},
		},
	}
	resp := &mdocker.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.CreateContainer", req, resp))

	client.ContainerID = resp.Containers[0].ID
}

func TestListContainers(t *testing.T) {
	req := &mdocker.ContainerRequest{}
	resp := &mdocker.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.ListContainers", req, resp))
}

func TestGetContainer(t *testing.T) {
	req := &mdocker.ContainerRequest{
		ID: client.ContainerID,
	}
	resp := &mdocker.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.GetContainer", req, resp))

	h.Equals(t, client.ContainerID, resp.Containers[0].ID)
}

func TestSaveContainer(t *testing.T) {
	req := &mdocker.ContainerRequest{
		ID: client.ContainerID,
		Opts: &docker.CommitContainerOptions{
			Container:  client.ContainerID,
			Repository: "test-commit",
		},
	}
	resp := &mdocker.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.SaveContainer", req, resp))
}

func TestStartContainer(t *testing.T) {
	req := &mdocker.ContainerRequest{
		ID: client.ContainerID,
	}
	resp := &mdocker.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.StartContainer", req, resp))
}

func TestStopContainer(t *testing.T) {
	req := &mdocker.ContainerRequest{
		ID: client.ContainerID,
	}
	resp := &mdocker.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.StopContainer", req, resp))
}

func TestDeleteContainer(t *testing.T) {
	req := &mdocker.ContainerRequest{
		ID: client.ContainerID,
	}
	resp := &mdocker.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.DeleteContainer", req, resp))

	ireq := &mdocker.ImageRequest{
		Name: client.ImageName,
	}
	iresp := &mdocker.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.DeleteImage", ireq, iresp))
}
