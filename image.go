package mdocker

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent/rpc"
)

var repoPrefix = "mistify-imports"

func repoNameFromID(imageID string) string {
	return fmt.Sprintf("%s/%s", repoPrefix, imageID)
}

func idFromRepoTag(repoTag string) string {
	repo, _ := docker.ParseRepositoryTag(repoTag)
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return ""
	}
	if parts[0] == repoPrefix {
		return parts[1]
	}
	return ""
}

// ListImages retrieves a list of Docker images
func (md *MDocker) ListImages(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	opts := docker.ListImagesOptions{}

	apiImages, err := md.client.ListImages(opts)
	if err != nil {
		return err
	}
	images := make([]*rpc.Image, 0, len(apiImages))
	for _, ai := range apiImages {
		id := idFromRepoTag(ai.RepoTags[0])
		if id != "" {
			images = append(images, &rpc.Image{
				Id:   id,
				Type: "container",
				Size: uint64(ai.Size) / 1024 / 1024,
			})
		}
	}

	response.Images = images
	return nil
}

// GetImage retrieves information about a specific Docker image
func (md *MDocker) GetImage(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	image, err := md.client.InspectImage(repoNameFromID(request.Id))
	if err != nil {
		return err
	}

	response.Images = []*rpc.Image{
		&rpc.Image{
			Id:   request.Id,
			Type: "container",
			Size: uint64(image.Size) / 1024 / 1024,
		},
	}
	return nil
}

// ImportImage downloads a new container image from the image service and
// imports it into Docker
func (md *MDocker) ImportImage(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	repo := repoNameFromID(request.Id)

	opts := docker.ImportImageOptions{
		Repository: repo,
		Source:     fmt.Sprintf("http://%s/images/%s/download", md.imageService, request.Id),
		Tag:        "latest",
	}

	// Check if we already have the image to avoid unnecessary pulling
	image, err := md.client.InspectImage(repo)
	if err != nil && err != docker.ErrNoSuchImage {
		return err
	}

	if image == nil {
		if err := md.client.ImportImage(opts); err != nil {
			return err
		}

		image, err = md.client.InspectImage(repo)
		if err != nil {
			return err
		}
	}

	response.Images = []*rpc.Image{
		&rpc.Image{
			Id:   request.Id,
			Type: "container",
			Size: uint64(image.Size) / 1024 / 1024,
		},
	}
	return nil
}

// DeleteImage deletes a Docker image
func (md *MDocker) DeleteImage(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	image, err := md.client.InspectImage(repoNameFromID(request.Id))
	if err != nil {
		return err
	}

	opts := docker.RemoveImageOptions{}
	if err := md.client.RemoveImageExtended(image.ID, opts); err != nil {
		return err
	}

	response.Images = []*rpc.Image{
		&rpc.Image{
			Id:   request.Id,
			Type: "container",
			Size: uint64(image.Size) / 1024 / 1024,
		},
	}
	return nil
}
