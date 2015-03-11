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
		Images []docker.APIImages `json:"images"` // Slice of one or more images
	}
)

// ListImages retrieves a list of Docker images
func (md *MDocker) ListImages(h *http.Request, request *ImageRequest, response *ImageResponse) error {
	return nil
}

// GetImage retrieves information about a specific Docker image
func (md *MDocker) GetImage(h *http.Request, request *ImageRequest, response *ImageResponse) error {
	return nil
}

// PullImage downloads a new Docker image
func (md *MDocker) PullImage(h *http.Request, request *ImageRequest, response *ImageResponse) error {
	return nil
}

// DeleteImage deletes a Docker image
func (md *MDocker) DeleteImage(h *http.Request, request *ImageRequest, response *ImageResponse) error {
	return nil
}
