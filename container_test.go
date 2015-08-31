package mdocker_test

import (
	"testing"
	"time"

	h "github.com/bakins/test-helpers"
	"github.com/fsouza/go-dockerclient"
	rpcClient "github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
)

func newGuest() *rpcClient.Guest {
	return &rpcClient.Guest{
		ID:    "foobar",
		Type:  "container",
		Image: client.ImageID,
		Nics: []rpcClient.Nic{
			rpcClient.Nic{
				Name:    "test",
				Network: "mistify0",
				Mac:     "C0:B6:C5:EA:93:AC",
				VLANs:   []int{},
			},
		},
	}
}

func createMainContainer(t *testing.T) {
	if client.ContainerName != "" {
		return
	}

	req := &rpc.GuestRequest{
		Guest:  newGuest(),
		Action: "containerCreate",
	}
	resp := &rpc.GuestResponse{}

	h.Ok(t, client.rpc.Do("MDocker.CreateContainer", req, resp))
	client.ContainerName = req.Guest.ID
}

func deleteMainContainer(t *testing.T) {
	if client.ContainerName == "" {
		return
	}

	req := &rpc.GuestRequest{
		Guest:  newGuest(),
		Action: "containerDelete",
	}
	resp := &rpc.GuestResponse{}

	h.Ok(t, client.rpc.Do("MDocker.DeleteContainer", req, resp))
	client.ContainerName = ""
}

func TestCreateContainer(t *testing.T) {
	importMainImage(t)
	deleteMainContainer(t)
	createMainContainer(t)

	req := &rpc.GuestRequest{
		Guest: &rpcClient.Guest{
			ID:    "foobar",
			Type:  "container",
			Image: "asdfasdfpoih",
		},
		Action: "containerDelete",
	}
	resp := &rpc.GuestResponse{}

	h.Assert(t, client.rpc.Do("MDocker.CreateContainer", req, resp) != nil, "bad image should error")
}

func TestListContainers(t *testing.T) {
	importMainImage(t)
	createMainContainer(t)

	req := &rpc.ContainerRequest{
		Opts: &docker.ListContainersOptions{
			All: true,
		},
	}
	resp := &rpc.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.ListContainers", req, resp))
}

func TestGetContainer(t *testing.T) {
	importMainImage(t)
	createMainContainer(t)

	req := &rpc.ContainerRequest{
		ID: client.ContainerName,
	}
	resp := &rpc.ContainerResponse{}

	h.Ok(t, client.rpc.Do("MDocker.GetContainer", req, resp))
}

func TestSaveContainer(t *testing.T) {
	importMainImage(t)
	createMainContainer(t)

	req := &rpc.ContainerRequest{
		ID: client.ContainerName,
		Opts: &docker.CommitContainerOptions{
			Container:  client.ContainerName,
			Repository: "test-commit",
		},
	}
	resp := &rpc.ImageResponse{}
	h.Ok(t, client.rpc.Do("MDocker.SaveContainer", req, resp))

	ireq := &rpc.ImageRequest{
		ID: "test-commit",
	}
	iresp := &rpc.ImageResponse{}
	h.Ok(t, client.rpc.Do("MDocker.DeleteImage", ireq, iresp))
}

func TestStartContainer(t *testing.T) {
	importMainImage(t)
	createMainContainer(t)

	badreq := &rpc.GuestRequest{
		Guest: &rpcClient.Guest{
			ID:    "foobar2",
			Type:  "container",
			Image: "asdfasdfpoih",
		},
		Action: "containerDelete",
	}
	badresp := &rpc.GuestResponse{}

	h.Assert(t, client.rpc.Do("MDocker.StartContainer", badreq, badresp) != nil, "bad container should error")

	req := &rpc.GuestRequest{
		Guest:  newGuest(),
		Action: "containerStart",
	}
	resp := &rpc.GuestResponse{}
	h.Ok(t, client.rpc.Do("MDocker.StartContainer", req, resp))
	h.Equals(t, resp.Guest.State, "running")
	// Sleep a second here to avoid a race in subsequent container stopping
	time.Sleep(time.Second)
}

func TestStopContainer(t *testing.T) {
	importMainImage(t)
	createMainContainer(t)

	badreq := &rpc.GuestRequest{
		Guest: &rpcClient.Guest{
			ID:    "foobar2",
			Type:  "container",
			Image: "asdfasdfpoih",
		},
		Action: "containerStop",
	}
	badresp := &rpc.GuestResponse{}
	h.Assert(t, client.rpc.Do("MDocker.StopContainer", badreq, badresp) != nil, "bad container should error")
	req := &rpc.GuestRequest{
		Guest:  newGuest(),
		Action: "containerStop",
	}
	resp := &rpc.GuestResponse{}
	h.Ok(t, client.rpc.Do("MDocker.StopContainer", req, resp))
	h.Equals(t, resp.Guest.State, "stopped")
}

func TestDeleteContainer(t *testing.T) {
	importMainImage(t)
	createMainContainer(t)
	deleteMainContainer(t)
}
