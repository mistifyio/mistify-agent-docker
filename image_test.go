package mdocker_test

import (
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-agent-docker"
)

func pullMainImage(t *testing.T) {
	if client.ImageID != "" {
		return
	}
	req := &mdocker.ImageRequest{
		Name: client.ImageName,
	}
	resp := &mdocker.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.PullImage", req, resp))
	h.Equals(t, 1, len(resp.Images))

	client.ImageID = resp.Images[0].ID
}

func deleteMainImage(t *testing.T) {
	if client.ImageID == "" {
		return
	}
	req := &mdocker.ImageRequest{
		Name: client.ImageName,
	}
	resp := &mdocker.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.DeleteImage", req, resp))
	client.ImageID = ""
}

func TestPullImage(t *testing.T) {
	badreq := &mdocker.ImageRequest{
		Name: "asdfqewrty",
	}
	badresp := &mdocker.ImageResponse{}

	h.Assert(t, client.rpc.Do("MDocker.PullImage", badreq, badresp) != nil, "bad image name should error")

	// Make sure we're pulling a fresh image
	deleteMainImage(t)
	pullMainImage(t)
}

func TestListImages(t *testing.T) {
	pullMainImage(t)

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
	pullMainImage(t)

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

	req = &mdocker.ImageRequest{
		ID: "asdfasdfa",
	}
	resp = &mdocker.ImageResponse{}

	h.Assert(t, client.rpc.Do("MDocker.GetImage", req, resp) != nil, "bad id should error")
}

func TestDeleteImage(t *testing.T) {
	pullMainImage(t)
	deleteMainImage(t)

	req := &mdocker.ImageRequest{
		Name: client.ImageName,
	}
	resp := &mdocker.ImageResponse{}
	h.Assert(t, client.rpc.Do("MDocker.DeleteImage", req, resp) != nil, "deleting missing image should error")
}
