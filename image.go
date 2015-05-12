package mdocker

import (
	"net/http"

	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent/rpc"
)

// ListImages retrieves a list of Docker images
func (md *MDocker) ListImages(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	opts := docker.ListImagesOptions{}

	apiImages, err := md.client.ListImages(opts)
	if err != nil {
		return err
	}
	images := make([]*rpc.Image, len(apiImages))
	for i, ai := range apiImages {
		images[i] = &rpc.Image{
			Id:   ai.ID,
			Type: "container",
			Size: uint64(ai.Size) / 1024 / 1024,
		}
	}

	response.Images = images
	return nil
}

// GetImage retrieves information about a specific Docker image
func (md *MDocker) GetImage(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	image, err := md.client.InspectImage(request.Id)
	if err != nil {
		return err
	}

	response.Images = []*rpc.Image{
		&rpc.Image{
			Id:   image.ID,
			Type: "container",
			Size: uint64(image.Size) / 1024 / 1024,
		},
	}
	return nil
}

// PullImage downloads a new Docker image
func (md *MDocker) PullImage(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	opts := docker.PullImageOptions{
		Repository: request.Source,
		Tag:        "latest",
	}

	// Check if we already have the image to avoid unnecessary pulling
	image, err := md.client.InspectImage(request.Source)
	if err != nil && err != docker.ErrNoSuchImage {
		return err
	}

	if image == nil {
		if err := md.client.PullImage(opts, docker.AuthConfiguration{}); err != nil {
			return err
		}

		image, err = md.client.InspectImage(request.Source)
		if err != nil {
			return err
		}
	}

	response.Images = []*rpc.Image{
		&rpc.Image{
			Id:   image.ID,
			Type: "container",
			Size: uint64(image.Size) / 1024 / 1024,
		},
	}
	return nil
}

// DeleteImage deletes a Docker image
func (md *MDocker) DeleteImage(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	image, err := md.client.InspectImage(request.Id)
	if err != nil {
		return err
	}

	opts := docker.RemoveImageOptions{}
	if err := md.client.RemoveImageExtended(image.ID, opts); err != nil {
		return err
	}

	response.Images = []*rpc.Image{
		&rpc.Image{
			Id:   image.ID,
			Type: "container",
			Size: uint64(image.Size) / 1024 / 1024,
		},
	}
	return nil
}
