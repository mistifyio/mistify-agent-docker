package mdocker_test

import (
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-agent-docker"
)

// NOTE: Must Run First
func TestPullImage(t *testing.T) {
	req := &mdocker.ImageRequest{
		Name: client.ImageName,
	}
	resp := &mdocker.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.PullImage", req, resp))

	client.ImageID = resp.Images[0].ID
}

func TestListImages(t *testing.T) {
	req := &mdocker.ImageRequest{}
	resp := &mdocker.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.ListImages", req, resp))

	found := false
	for _, i := range resp.Images {
		if i.ID == client.ImageID {
			found = true
			break
		}
	}
	h.Assert(t, found, "did not find pulled image in list")
}

func TestGetImage(t *testing.T) {
	req := &mdocker.ImageRequest{
		Name: client.ImageName,
	}
	resp := &mdocker.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.GetImage", req, resp))
	h.Equals(t, client.ImageID, resp.Images[0].ID)

	req = &mdocker.ImageRequest{
		ID: client.ImageID,
	}
	resp = &mdocker.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.GetImage", req, resp))
	h.Equals(t, client.ImageID, resp.Images[0].ID)
}

// NOTE: Must Run Last
func TestDeleteImage(t *testing.T) {
	req := &mdocker.ImageRequest{
		Name: client.ImageName,
	}
	resp := &mdocker.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.DeleteImage", req, resp))
}
