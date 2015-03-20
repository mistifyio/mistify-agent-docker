package mdocker_test

import (
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent/rpc"
)

func createMainContainer(t *testing.T) {
	if client.ContainerID != "" {
		return
	}

	req := &rpc.ContainerRequest{
		Opts: &docker.CreateContainerOptions{
			Config: &docker.Config{
				Image: client.ImageName,
				Cmd:   []string{"sleep", "5"},
			},
		},
	}
	resp := &rpc.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.CreateContainer", req, resp))
	client.ContainerID = resp.Containers[0].ID
}

func deleteMainContainer(t *testing.T) {
	if client.ContainerID == "" {
		return
	}

	req := &rpc.ContainerRequest{
		ID: client.ContainerID,
	}
	resp := &rpc.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.DeleteContainer", req, resp))
	client.ContainerID = ""
}

func TestCreateContainer(t *testing.T) {
	pullMainImage(t)
	deleteMainContainer(t)
	createMainContainer(t)

	req := &rpc.ContainerRequest{
		Opts: &docker.CreateContainerOptions{
			Config: &docker.Config{
				Image: "asdfasdfaf",
				Cmd:   []string{"sh"},
			},
		},
	}
	resp := &rpc.ContainerResponse{}
	h.Assert(t, client.rpc.Do("MDocker.CreateContainer", req, resp) != nil, "bad image should error")
}

func TestListContainers(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)

	req := &rpc.ContainerRequest{
		Opts: &docker.ListContainersOptions{
			All: true,
		},
	}
	resp := &rpc.ContainerResponse{}

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

	req := &rpc.ContainerRequest{
		ID: client.ContainerID,
	}
	resp := &rpc.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.GetContainer", req, resp))

	h.Equals(t, client.ContainerID, resp.Containers[0].ID)
}

func TestSaveContainer(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)

	req := &rpc.ContainerRequest{
		ID: client.ContainerID,
		Opts: &docker.CommitContainerOptions{
			Container:  client.ContainerID,
			Repository: "test-commit",
		},
	}
	resp := &rpc.ContainerImageResponse{}
	h.Ok(t, client.rpc.Do("MDocker.SaveContainer", req, resp))

	ireq := &rpc.ContainerImageRequest{
		Name: "test-commit",
	}
	iresp := &rpc.ContainerImageResponse{}
	h.Ok(t, client.rpc.Do("MDocker.DeleteImage", ireq, iresp))
}

func TestStartContainer(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)

	req := &rpc.ContainerRequest{
		ID: "asdfouasdfafd",
	}
	resp := &rpc.ContainerResponse{}
	h.Assert(t, client.rpc.Do("MDocker.StartContainer", req, resp) != nil, "bad container should error")

	req = &rpc.ContainerRequest{
		ID: client.ContainerID,
	}
	resp = &rpc.ContainerResponse{}
	h.Ok(t, client.rpc.Do("MDocker.StartContainer", req, resp))
	h.Assert(t, resp.Containers[0].State.Running, "container should be running")
}

func TestStopContainer(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)

	req := &rpc.ContainerRequest{
		ID: client.ContainerID,
	}
	resp := &rpc.ContainerResponse{}
	h.Ok(t, client.rpc.Do("MDocker.StopContainer", req, resp))
	h.Assert(t, !resp.Containers[0].State.Running, "container should not be running")

	req = &rpc.ContainerRequest{
		ID: "asdfasdfasdf",
	}
	resp = &rpc.ContainerResponse{}
	h.Assert(t, client.rpc.Do("MDocker.StopContainer", req, resp) != nil, "bad container should error")
}

func TestDeleteContainer(t *testing.T) {
	pullMainImage(t)
	createMainContainer(t)
	deleteMainContainer(t)
}
