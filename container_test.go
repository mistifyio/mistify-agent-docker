package mdocker_test

import (
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent-docker"
)

func createMainContainer(t *testing.T) {
	if client.ContainerID != "" {
		return
	}

	req := &mdocker.ContainerRequest{
		Opts: &docker.CreateContainerOptions{
			Config: &docker.Config{
				Image: client.ImageName,
				Cmd:   []string{"sh"},
			},
		},
	}
	resp := &mdocker.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.CreateContainer", req, resp))
	client.ContainerID = resp.Containers[0].ID
}

func deleteMainContainer(t *testing.T) {
	if client.ContainerID == "" {
		return
	}

	req := &mdocker.ContainerRequest{
		ID: client.ContainerID,
	}
	resp := &mdocker.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.DeleteContainer", req, resp))
	client.ContainerID = ""
}

func TestCreateContainer(t *testing.T) {
	pullMainImage(t)
	deleteMainContainer(t)
	createMainContainer(t)

	req := &mdocker.ContainerRequest{
		Opts: &docker.CreateContainerOptions{
			Config: &docker.Config{
				Image: "asdfasdfaf",
				Cmd:   []string{"sh"},
			},
		},
	}
	resp := &mdocker.ContainerResponse{}
	h.Assert(t, client.rpc.Do("MDocker.CreateContainer", req, resp) != nil, "bad image should error")
}

func TestListContainers(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)

	req := &mdocker.ContainerRequest{
		Opts: &docker.ListContainersOptions{
			All: true,
		},
	}
	resp := &mdocker.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.ListContainers", req, resp))

	found := false
	for _, c := range resp.Containers {
		if c.ID == client.ContainerID {
			found = true
			break
		}
	}
	h.Assert(t, found, "did not find created container in list")

}

func TestGetContainer(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)

	req := &mdocker.ContainerRequest{
		ID: client.ContainerID,
	}
	resp := &mdocker.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.GetContainer", req, resp))

	h.Equals(t, client.ContainerID, resp.Containers[0].ID)
}

func TestSaveContainer(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)

	req := &mdocker.ContainerRequest{
		ID: client.ContainerID,
		Opts: &docker.CommitContainerOptions{
			Container:  client.ContainerID,
			Repository: "test-commit",
		},
	}
	resp := &mdocker.ImageResponse{}
	h.Ok(t, client.rpc.Do("MDocker.SaveContainer", req, resp))

	ireq := &mdocker.ImageRequest{
		Name: "test-commit",
	}
	iresp := &mdocker.ImageResponse{}
	h.Ok(t, client.rpc.Do("MDocker.DeleteImage", ireq, iresp))
}

func TestStartContainer(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)

	req := &mdocker.ContainerRequest{
		ID: "asdfouasdfafd",
	}
	resp := &mdocker.ContainerResponse{}
	h.Assert(t, client.rpc.Do("MDocker.StartContainer", req, resp) != nil, "bad container should error")

	req = &mdocker.ContainerRequest{
		ID: client.ContainerID,
	}
	resp = &mdocker.ContainerResponse{}
	h.Ok(t, client.rpc.Do("MDocker.StartContainer", req, resp))
}

func TestStopContainer(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)

	req := &mdocker.ContainerRequest{
		ID: client.ContainerID,
	}
	resp := &mdocker.ContainerResponse{}
	h.Ok(t, client.rpc.Do("MDocker.StopContainer", req, resp))

	req = &mdocker.ContainerRequest{
		ID: "asdfasdfasdf",
	}
	resp = &mdocker.ContainerResponse{}
	h.Assert(t, client.rpc.Do("MDocker.StopContainer", req, resp) != nil, "bad container should error")
}

func TestDeleteContainer(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)
	deleteMainContainer(t)
}
