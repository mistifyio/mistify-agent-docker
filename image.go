package mdocker

import (
	"net/http"

	"github.com/fsouza/go-dockerclient"
)

type (
	// ImageRequest is an image request to the Docker sub-agent
	ImageRequest struct {
		ID   string      `json:"id"`   // Image ID
		Name string      `json:"name"` // Image name
		Opts interface{} `json:"opts"` // Generic Options. Will need converting
	}

	// ImageResponse is an image response from the Docker sub-agent
	ImageResponse struct {
		Images []*docker.Image `json:"images"` // Slice of one or more images
	}
)

// GetOpts returns the Opts property
func (ireq *ImageRequest) GetOpts() interface{} {
	return ireq.Opts
}

// GetLookup returns the string to look an image up by based on field priority
func (ireq *ImageRequest) GetLookup(d string) string {
	if ireq.ID != "" {
		return ireq.ID
	}
	if ireq.Name != "" {
		return ireq.Name
	}
	return d
}

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
func (md *MDocker) ListImages(h *http.Request, request *ImageRequest, response *ImageResponse) error {
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
func (md *MDocker) GetImage(h *http.Request, request *ImageRequest, response *ImageResponse) error {
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
func (md *MDocker) PullImage(h *http.Request, request *ImageRequest, response *ImageResponse) error {
	var opts docker.PullImageOptions
	if err := md.RequestOpts(request, opts); err != nil {
		return err
	}

	opts.Repository = request.GetLookup(opts.Repository)
	if opts.Tag == "" {
		opts.Tag = "latest"
	}

	if err := md.client.PullImage(opts, docker.AuthConfiguration{}); err != nil {
		return err
	}

	image, err := md.client.InspectImage(opts.Repository)
	if err != nil {
		return err
	}

	response.Images = []*docker.Image{
		image,
	}
	return nil
}

// DeleteImage deletes a Docker image
func (md *MDocker) DeleteImage(h *http.Request, request *ImageRequest, response *ImageResponse) error {
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
