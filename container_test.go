package mdocker_test

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
	"github.com/stretchr/testify/suite"
)

type ContainerTestSuite struct {
	APITestSuite
}

func (s *ContainerTestSuite) SetupTest() {
	s.APITestSuite.SetupTest()
	_ = s.loadImage()
}

func TestContainerTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerTestSuite))
}

/*
func (s *ContainerTestSuite) TestCreateContainer() {

}
*/

func (s *ContainerTestSuite) TestListContainers() {
	_ = s.createContainer()

	request := &rpc.ContainerRequest{
		Opts: &docker.ListContainersOptions{
			All: true,
		},
	}
	response := &rpc.ContainerResponse{}

	s.NoError(s.Client.Do("MDocker.ListContainers", request, response))
	s.Len(response.Containers, 1)
}

func (s *ContainerTestSuite) TestGetContainer() {
	guest := s.createContainer()

	request := &rpc.ContainerRequest{
		ID: guest.ID,
	}
	response := &rpc.ContainerResponse{}

	s.NoError(s.Client.Do("MDocker.GetContainer", request, response))
	s.Len(response.Containers, 1)
	s.Equal("/"+guest.ID, response.Containers[0].Name)
}

func (s *ContainerTestSuite) TestSaveContainer() {
	guest := s.createContainer()

	request := &rpc.ContainerRequest{
		ID: guest.ID,
		Opts: &docker.CommitContainerOptions{
			Container:  guest.ID,
			Repository: "test-commit",
		},
	}
	response := &rpc.ImageResponse{}
	s.NoError(s.Client.Do("MDocker.SaveContainer", request, response))

	// Cleanup
	delRequest := &rpc.ImageRequest{
		ID: "test-commit",
	}
	delResponse := &rpc.ImageResponse{}
	s.NoError(s.Client.Do("MDocker.DeleteImage", delRequest, delResponse))
}

func (s *ContainerTestSuite) TestStartContainer() {
	guest := s.createContainer()

	badrequest := &rpc.GuestRequest{
		Guest: &client.Guest{
			ID:    "foobar2",
			Type:  "container",
			Image: "asdfasdfpoih",
		},
	}
	badresponse := &rpc.GuestResponse{}

	s.Error(s.Client.Do("MDocker.StartContainer", badrequest, badresponse))

	request := &rpc.GuestRequest{
		Guest: &client.Guest{
			ID: guest.ID,
		},
	}
	response := &rpc.GuestResponse{}
	s.NoError(s.Client.Do("MDocker.StartContainer", request, response))
	s.Equal("running", response.Guest.State)
}

func (s *ContainerTestSuite) TestStopContainer() {
	guest := s.createContainer()

	badrequest := &rpc.GuestRequest{
		Guest: &client.Guest{
			ID:    "foobar2",
			Type:  "container",
			Image: "asdfasdfpoih",
		},
	}
	badresponse := &rpc.GuestResponse{}
	s.Error(s.Client.Do("MDocker.StopContainer", badrequest, badresponse))

	request := &rpc.GuestRequest{
		Guest: &client.Guest{
			ID: guest.ID,
		},
	}
	response := &rpc.GuestResponse{}
	s.NoError(s.Client.Do("MDocker.StopContainer", request, response))
	s.Equal("stopped", response.Guest.State)
}

func (s *ContainerTestSuite) TestDeleteContainer() {
	guest := s.createContainer()

	request := &rpc.GuestRequest{
		Guest: &client.Guest{
			ID: guest.ID,
		},
	}
	response := &rpc.GuestResponse{}
	s.NoError(s.Client.Do("MDocker.DeleteContainer", request, response))
}
