package mdocker_test

import (
	"testing"

	h "github.com/bakins/test-helpers"
	"github.com/mistifyio/mistify-agent/rpc"
)

func importMainImage(t *testing.T) {
	if client.ImageImported {
		return
	}
	req := &rpc.ImageRequest{
		ID: client.ImageID,
	}
	resp := &rpc.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.LoadImage", req, resp))
	h.Equals(t, 1, len(resp.Images))

	client.ImageImported = true
}

func deleteMainImage(t *testing.T) {
	if !client.ImageImported {
		return
	}
	req := &rpc.ImageRequest{
		ID: client.ImageID,
	}
	resp := &rpc.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.DeleteImage", req, resp))
	client.ImageImported = false
}

func TestLoadImage(t *testing.T) {
	badreq := &rpc.ImageRequest{
		ID: "asdfqewrty",
	}
	badresp := &rpc.ImageResponse{}

	h.Assert(t, client.rpc.Do("MDocker.LoadImage", badreq, badresp) != nil, "bad image id should error")

	// Make sure we're importing a fresh image
	deleteMainImage(t)
	importMainImage(t)
}

func TestListImages(t *testing.T) {
	importMainImage(t)

	req := &rpc.ImageRequest{}
	resp := &rpc.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.ListImages", req, resp))

	found := false
	for _, i := range resp.Images {
		if i.ID == client.ImageID {
			found = true
			break
		}
	}
	h.Assert(t, found, "did not find imported image in list")
}

func TestGetImage(t *testing.T) {
	importMainImage(t)

	req := &rpc.ImageRequest{
		ID: client.ImageID,
	}
	resp := &rpc.ImageResponse{}

	h.Ok(t, client.rpc.Do("MDocker.GetImage", req, resp))
	h.Equals(t, client.ImageID, resp.Images[0].ID)

	req = &rpc.ImageRequest{
		ID: "asdfasdfa",
	}
	resp = &rpc.ImageResponse{}

	h.Assert(t, client.rpc.Do("MDocker.GetImage", req, resp) != nil, "bad id should error")
}

func TestDeleteImage(t *testing.T) {
	importMainImage(t)
	deleteMainImage(t)

	req := &rpc.ImageRequest{
		ID: client.ImageID,
	}
	resp := &rpc.ImageResponse{}
	h.Assert(t, client.rpc.Do("MDocker.DeleteImage", req, resp) != nil, "deleting missing image should error")
}
