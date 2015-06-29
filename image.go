package mdocker

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
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
	name := request.Id

	// Check if we already have the image to avoid unnecessary pulling
	image, err := md.client.InspectImage(name)
	if err != nil && err != docker.ErrNoSuchImage {
		return err
	}

	if image == nil {
		// Docker Import lets the image get renamed, but it strips metadata
		// (which includes any CMD that had been set). Docker Load doesn't let
		// the image get renamed and doesn't return the image id or name after
		// loading, but does preserve metadata. Since a rename to use the
		// image-service assigned id and metadata with CMD are both necessary,
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

		// Use a response buffer so the first few bytes can be peeked at for
		// file type detection. Uncompress the image if it is gzipped
		responseBuffer := bufio.NewReader(resp.Body)
		var imageReader io.Reader = responseBuffer
		filetypeBytes, err := responseBuffer.Peek(512)
		if err != nil {
			return err
		}
		if http.DetectContentType(filetypeBytes) == "application/x-gzip" {
			gzipReader, err := gzip.NewReader(responseBuffer)
			if err != nil {
				return err
			}
			defer gzipReader.Close()
			imageReader = gzipReader
		}

		pipeReader, pipeWriter := io.Pipe()

		go fixRepositoriesFile(name, imageReader, pipeWriter)

		opts := docker.LoadImageOptions{
			InputStream: pipeReader,
		}
		if err := md.client.LoadImage(opts); err != nil {
			return err
		}

		image, err = md.client.InspectImage(name)
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

// fixRepositoriesFile changes the repo name to the mistify-image-service's
// assigned image id and tag to "latest" before it is loaded into docker
func fixRepositoriesFile(newName string, in io.Reader, out io.WriteCloser) {
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
			// Update the image name and tag in the repositories file
			if header.Name == "repositories" {
				// Read the file and parse the JSON
				// {"reponame":{"tag":"hash"}}
				repoMap := map[string]map[string]string{}
				jsonDecoder := json.NewDecoder(tarReader)
				if err := jsonDecoder.Decode(&repoMap); err != nil {
					log.WithField("error", err).Error("failed to parse repositories json")
					return
				}

				// Should only be one key. Replace it with the new repo name
				if len(repoMap) != 1 {
					log.WithFields(log.Fields{
						"error":   errors.New("incorrect number of repos"),
						"repoMap": repoMap,
					}).Error("must be only one repo specified")
					return
				}
				for oldName := range repoMap {
					tagMap := repoMap[oldName]
					delete(repoMap, oldName)

					// Should only be one tag. Replace it with the new repo name
					if len(tagMap) != 1 {
						log.WithFields(log.Fields{
							"error":   errors.New("incorrect number of tags"),
							"repoMap": repoMap,
						}).Error("must be only one tag specified")
						return
					}
					for oldTag := range tagMap {
						// Only rename if the tag is not already "latest"
						if oldTag == "latest" {
							break
						}
						tagMap["latest"] = tagMap[oldTag]
						delete(tagMap, oldTag)
					}
					repoMap[newName] = tagMap
				}

				// Update the header
				outBytes, err := json.Marshal(repoMap)
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
				continue
			}
			fallthrough
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
	if err := md.client.RemoveImageExtended(request.Id, opts); err != nil {
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
