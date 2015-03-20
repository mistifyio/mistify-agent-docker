package mdocker

import (
	"net/http"

	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent/rpc"
)

// imagesFromAPIImages converts an array of APIImages into an array of Image
func (md *MDocker) imagesFromAPIImages(ais []docker.APIImages) ([]*docker.Image, error) {
	images := make([]*docker.Image, 0, len(ais))
	for _, ai := range ais {
		image, err := md.client.InspectImage(ai.ID)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, nil
}

// ListImages retrieves a list of Docker images
func (md *MDocker) ListImages(h *http.Request, request *rpc.ContainerImageRequest, response *rpc.ContainerImageResponse) error {
	opts := docker.ListImagesOptions{}
	if err := md.RequestOpts(request, &opts); err != nil {
		return err
	}

	apiImages, err := md.client.ListImages(opts)
	if err != nil {
		return err
	}
	images, err := md.imagesFromAPIImages(apiImages)
	if err != nil {
		return err
	}

	response.Images = images
	return nil
}

// GetImage retrieves information about a specific Docker image
func (md *MDocker) GetImage(h *http.Request, request *rpc.ContainerImageRequest, response *rpc.ContainerImageResponse) error {
	image, err := md.client.InspectImage(request.GetLookup(""))
	if err != nil {
		return err
	}

	response.Images = []*docker.Image{
		image,
	}
	return nil
}

// PullImage downloads a new Docker image
func (md *MDocker) PullImage(h *http.Request, request *rpc.ContainerImageRequest, response *rpc.ContainerImageResponse) error {
	var opts docker.PullImageOptions
	if err := md.RequestOpts(request, opts); err != nil {
		return err
	}

	opts.Repository = request.GetLookup(opts.Repository)
	if opts.Tag == "" {
		opts.Tag = "latest"
	}

	// Check if we already have the image to avoid unnecessary pulling
	image, err := md.client.InspectImage(opts.Repository)
	if err != nil && err != docker.ErrNoSuchImage {
		return err
	}

	if image == nil {
		if err := md.client.PullImage(opts, docker.AuthConfiguration{}); err != nil {
			return err
		}

		image, err = md.client.InspectImage(opts.Repository)
		if err != nil {
			return err
		}
	}

	response.Images = []*docker.Image{
		image,
	}
	return nil
}

// DeleteImage deletes a Docker image
func (md *MDocker) DeleteImage(h *http.Request, request *rpc.ContainerImageRequest, response *rpc.ContainerImageResponse) error {
	image, err := md.client.InspectImage(request.GetLookup(""))
	if err != nil {
		return err
	}

	var opts docker.RemoveImageOptions
	if err := md.RequestOpts(request, opts); err != nil {
		return err
	}
	if err := md.client.RemoveImageExtended(image.ID, opts); err != nil {
		return err
	}

	response.Images = []*docker.Image{
		image,
	}
	return nil
}
