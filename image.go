package mdocker

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
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
	images := make([]*rpc.Image, 0, len(apiImages))
	for _, ai := range apiImages {
		id, _ := docker.ParseRepositoryTag(ai.RepoTags[0])
		images = append(images, &rpc.Image{
			Id:   id,
			Type: "container",
			Size: uint64(ai.Size) / 1024 / 1024,
		})
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
			Id:   request.Id,
			Type: "container",
			Size: uint64(image.Size) / 1024 / 1024,
		},
	}
	return nil
}

// LoadImage downloads a new container image from the image service and
// imports it into Docker
func (md *MDocker) LoadImage(h *http.Request, request *rpc.ImageRequest, response *rpc.ImageResponse) error {
	repo := request.Id

	// Check if we already have the image to avoid unnecessary pulling
	image, err := md.client.InspectImage(repo)
	if err != nil && err != docker.ErrNoSuchImage {
		return err
	}

	if image == nil {
		// Docker Import lets the image repo get renamed, but it strips
		// metadata (which includes the set cmd). Docker Load doesn't let the
		// repo get renamed and doesn't return the image id or repo after
		// loading, but does preserve metadata. Since a rename to use the
		// image-service assigned id and metadata with cmd are both necessary,
		// the only way forward is to update the repositories file inside and
		// then load it.
		source := fmt.Sprintf("http://%s/images/%s/download", md.imageService, request.Id)
		resp, err := http.Get(source)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return ErrorHTTPCode{
				Expected: http.StatusOK,
				Code:     resp.StatusCode,
				Source:   source,
			}
		}

		pipeReader, pipeWriter := io.Pipe()

		go fixRepositoriesFile(repo, resp.Body, pipeWriter)

		opts := docker.LoadImageOptions{
			InputStream: pipeReader,
		}
		if err := md.client.LoadImage(opts); err != nil {
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

func fixRepositoriesFile(newRepoName string, in io.Reader, out io.WriteCloser) {
	defer out.Close()
	tarReader := tar.NewReader(in)
	tarWriter := tar.NewWriter(out)
	defer tarWriter.Close()

	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.WithField("error", err).Error("failed to get next tar header")
			return
		}

		switch header.Typeflag {
		case tar.TypeReg:
			// Update the repo name tarReader the repositories file
			if header.Name == "repositories" {
				// Read the file and parse the JSON
				inBytes := make([]byte, header.Size)
				if _, err := tarReader.Read(inBytes); err != nil {
					log.WithField("error", err).Error("failed to read repositories file")
					return
				}
				origRepo := map[string]interface{}{}
				if err := json.Unmarshal(inBytes, &origRepo); err != nil {
					log.WithField("error", err).Error("failed to parse repositories json")
					return
				}
				// Should only be one key. Replace it with the new repo name
				newRepo := map[string]interface{}{}
				for oldName := range origRepo {
					newRepo[newRepoName] = origRepo[oldName]
					break
				}

				// Update the header
				outBytes, err := json.Marshal(newRepo)
				if err != nil {
					log.WithField("error", err).Error("failed to marshal repositories json")
					return
				}
				header.Size = int64(len(outBytes))
				header.ModTime = time.Now()

				// Write the new header and data
				if err := tarWriter.WriteHeader(header); err != nil {
					log.WithField("error", err).Error("failed to write repositories header")
					return
				}
				if _, err := tarWriter.Write(outBytes); err != nil {
					log.WithField("error", err).Error("failed to write repositories json")
					return
				}
			} else {
				// Direct copy
				if err := tarWriter.WriteHeader(header); err != nil {
					log.WithField("error", err).Error("failed to write tar header")
					return
				}
				if _, err := io.Copy(tarWriter, tarReader); err != nil {
					log.WithField("error", err).Error("failed to copy tar file")
					return
				}
			}
		default:
			// Direct copy
			if err := tarWriter.WriteHeader(header); err != nil {
				log.WithField("error", err).Error("failed to write tar header")
				return
			}
			if _, err := io.Copy(tarWriter, tarReader); err != nil {
				log.WithField("error", err).Error("failed to copy tar body")
				return
			}
		}
	}
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
			Id:   request.Id,
			Type: "container",
			Size: uint64(image.Size) / 1024 / 1024,
		},
	}
	return nil
}
