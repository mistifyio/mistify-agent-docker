package mdocker_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent-docker"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
	logx "github.com/mistifyio/mistify-logrus-ext"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/tylerb/graceful"
)

type APITestSuite struct {
	suite.Suite
	Port         int
	Client       *rpc.Client
	MDocker      *mdocker.MDocker
	ImageServer  *httptest.Server
	ImageService string
	ImageID      string
	ImageData    []byte
	Docker       *docker.Client
	ContainerIDs []string
	Server       *graceful.Server
	Bridge       string
}

func (s *APITestSuite) SetupSuite() {
	log.SetLevel(log.FatalLevel)

	// Define ovs bridge
	// This will get created when ports are added for a container.
	s.Bridge = "mistify-agent-docker-test"

	// Set up client to interact with API
	s.Port = 54321
	s.Client, _ = rpc.NewClient(uint(s.Port), "")

	// Get the docker image to serve
	imgName := "tauzero/test-loop"
	s.Docker, _ = docker.NewClient("unix:///var/run/docker.sock")
	// Pull
	pullOpts := docker.PullImageOptions{
		Repository: imgName,
	}
	s.Require().NoError(s.Docker.PullImage(pullOpts, docker.AuthConfiguration{}))
	// Export
	output := new(bytes.Buffer)
	exportOpts := docker.ExportImageOptions{
		Name:         imgName,
		OutputStream: output,
	}
	s.Require().NoError(s.Docker.ExportImage(exportOpts))
	s.ImageData = output.Bytes()
	s.ImageID = uuid.New()

	// Remove
	s.Require().NoError(s.Docker.RemoveImage(imgName))

	// Set up a fake ImageService to fetch images from
	s.ImageServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/images/%s/download", s.ImageID) {
			if _, err := w.Write(s.ImageData); err != nil {
				log.WithField("error", err).Error("Failed to write mock image data to response")
			}
			return
		}

		if r.URL.Path == "/images/gzipID/download" {
			gzipWriter := gzip.NewWriter(w)
			defer logx.LogReturnedErr(gzipWriter.Close, nil, "failed to close gzip writer")
			if _, err := gzipWriter.Write(s.ImageData); err != nil {
				log.WithField("error", err).Error("Failed to write mock image data to response")
			}
			return
		}

		http.NotFound(w, r)
		return
	}))
	imageURL, _ := url.Parse(s.ImageServer.URL)
	s.ImageService = imageURL.Host
}

func (s *APITestSuite) SetupTest() {
	// Run the MDocker
	s.MDocker, _ = mdocker.New("unix:///var/run/docker.sock", s.ImageService, "")
	s.Server = s.MDocker.RunHTTP(uint(s.Port))
	// Sleep to give the server time to start listening
	time.Sleep(200 * time.Millisecond)
}

func (s *APITestSuite) TearDownTest() {
	// Stop the image store
	stopChan := s.Server.StopChan()
	s.Server.Stop(5 * time.Second)
	<-stopChan

	// Clean up docker
	for _, containerID := range s.ContainerIDs {
		opts := docker.RemoveContainerOptions{
			ID:    containerID,
			Force: true,
		}
		if err := s.Docker.RemoveContainer(opts); err != nil {
			log.WithField("error", err).Error("failed to remove container")
		}
	}
	if err := s.Docker.RemoveImage(s.ImageID); err != nil {
		log.WithField("error", err).Error("failed to remove image")
	}

	// Clean up ovs
	if output, err := exec.Command("ovs-vsctl", "del-br", s.Bridge).CombinedOutput(); err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"output": string(output),
		}).Error("failed to remove ovs bridge")
	}
}

func (s *APITestSuite) loadImage() *rpc.Image {
	response := &rpc.ImageResponse{}
	request := &rpc.ImageRequest{
		ID: s.ImageID,
	}
	s.NoError(s.Client.Do("MDocker.LoadImage", request, response))
	return response.Images[0]
}

func (s *APITestSuite) createContainer() *client.Guest {
	response := &rpc.GuestResponse{}
	request := &rpc.GuestRequest{
		Guest: &client.Guest{
			ID:     uuid.New(),
			Image:  s.ImageID,
			Memory: 10,
			Nics: []client.Nic{
				client.Nic{
					Name:    "test",
					Network: s.Bridge,
					Mac:     "13:7D:DA:F2:ED:63",
					VLANs:   []int{1, 2, 3},
				},
			},
		},
	}
	_ = s.Client.Do("MDocker.CreateContainer", request, response)
	// Keep track of the ID for easier cleanup later
	s.ContainerIDs = append(s.ContainerIDs, response.Guest.ID)
	return response.Guest
}

func testMsgFunc(prefix string) func(...interface{}) string {
	return func(val ...interface{}) string {
		if len(val) == 0 {
			return prefix
		}
		msgPrefix := prefix + " : "
		if len(val) == 1 {
			return msgPrefix + val[0].(string)
		} else {
			return msgPrefix + fmt.Sprintf(val[0].(string), val[1:]...)
		}
	}
}

func init() {
	d, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.WithField("error", err).Fatal("could not create docker client")
	}
	if err := d.Ping(); err != nil {
		log.WithField("error", err).Fatal("could not ping docker server")
	}
}
