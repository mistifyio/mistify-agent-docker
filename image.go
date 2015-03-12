package mdocker

import (
	"net/http"

	"github.com/fsouza/go-dockerclient"
)

// FakeImage is a hardcoded images for stubbing
// TODO: Remove once everything is built
var FakeImage = docker.APIImages{
	ID: "88f9454e60ddf4ae5f23fad8247a2c53e8d3ff63b0bdac59fc17ceceab058ce6",
	RepoTags: []string{
		"centos:7",
		"centos:centos7",
		"centos:latest",
	},
	Created:     1425505110,
	Size:        223930859,
	VirtualSize: 223930859,
	ParentID:    "5b12ef8fd57065237a6833039acc0e7f68e363c15d8abb5cacce7143a1f7de8a",
}

type (
	// ImageRequest is an image request to the Docker sub-agent
	ImageRequest struct {
		ID   string      `json:"id"`   // Image ID
		Name string      `json:"name"` // Image name
		Opts interface{} `json:"opts"` // Generic Options. Will need converting
	}

	// ImageResponse is an image response from the Docker sub-agent
	ImageResponse struct {
		Images []docker.APIImages `json:"images"` // Slice of one or more images
	}
)

// GetOpts returns the Opts property
func (ireq *ImageRequest) GetOpts() interface{} {
	return ireq.Opts
}

// ListImages retrieves a list of Docker images
func (md *MDocker) ListImages(h *http.Request, request *ImageRequest, response *ImageResponse) error {
	opts := &docker.ListImagesOptions{}
	if err := md.RequestOpts(request, opts); err != nil {
		return err
	}

	images, err := md.client.ListImages(*opts)
	if err != nil {
		return err
	}
	response.Images = images
	return nil
}

// GetImage retrieves information about a specific Docker image
func (md *MDocker) GetImage(h *http.Request, request *ImageRequest, response *ImageResponse) error {
	response.Images = []docker.APIImages{
		FakeImage,
	}
	return nil
}

// PullImage downloads a new Docker image
func (md *MDocker) PullImage(h *http.Request, request *ImageRequest, response *ImageResponse) error {
	response.Images = []docker.APIImages{
		FakeImage,
	}
	return nil
}

// DeleteImage deletes a Docker image
func (md *MDocker) DeleteImage(h *http.Request, request *ImageRequest, response *ImageResponse) error {
	response.Images = []docker.APIImages{
		FakeImage,
	}
	return nil
}
