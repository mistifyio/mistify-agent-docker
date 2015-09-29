package mdocker_test

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent/client"
	"github.com/mistifyio/mistify-agent/rpc"
	"github.com/pborman/uuid"
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

func (s *ContainerTestSuite) TestCreateContainer() {
	nics := []client.Nic{
		client.Nic{
			Name:    "test",
			Network: s.Bridge,
			Mac:     "13:7D:DA:F2:ED:63",
			VLANs:   []int{1, 2, 3},
		},
	}

	tests := []struct {
		description string
		guest       *client.Guest
		expectedErr bool
	}{
		{"missing id",
			&client.Guest{}, true},
		{"missing nics",
			&client.Guest{ID: uuid.New()}, true},
		{"missing image",
			&client.Guest{ID: uuid.New(), Nics: nics}, true},
		{"valid request",
			&client.Guest{ID: uuid.New(), Nics: nics, Image: s.ImageID, Memory: 10}, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)

		response := &rpc.GuestResponse{}
		request := &rpc.GuestRequest{Guest: test.guest}
		err := s.Client.Do("MDocker.CreateContainer", request, response)
		if test.expectedErr {
			s.Error(err, msg("should fail"))
		} else {
			s.NoError(err, msg("should succeed"))
		}
		// Track any IDs for cleanup
		if response.Guest != nil {
			s.ContainerIDs = append(s.ContainerIDs, response.Guest.ID)
		}
	}

}

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

	tests := []struct {
		description string
		ID          string
		expectedErr bool
	}{
		{"missing id", "", true},
		{"bad id", "asdf", true},
		{"valid id", guest.ID, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		request := &rpc.ContainerRequest{
			ID: test.ID,
		}
		response := &rpc.ContainerResponse{}

		err := s.Client.Do("MDocker.GetContainer", request, response)
		if test.expectedErr {
			s.Error(err, msg("should fail"))
		} else {
			s.NoError(s.Client.Do("MDocker.GetContainer", request, response), msg("should succeed"))
			s.Len(response.Containers, 1, msg("should only return one container"))
			s.Equal("/"+test.ID, response.Containers[0].Name, msg("should return the correct container"))
		}
	}
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
	s.testContainerAction("StartContainer", guest, "running")
}

func (s *ContainerTestSuite) TestStopContainer() {
	guest := s.createContainer()
	_, _ = s.containerAction("StartContainer", guest)
	s.testContainerAction("StopContainer", guest, "stopped")
}

func (s *ContainerTestSuite) TestRestartContainer() {
	guest := s.createContainer()
	_, _ = s.containerAction("StartContainer", guest)
	s.testContainerAction("RestartContainer", guest, "running")
}

func (s *ContainerTestSuite) TestRebootContainer() {
	guest := s.createContainer()
	_, _ = s.containerAction("StartContainer", guest)
	s.testContainerAction("RebootContainer", guest, "running")
}

func (s *ContainerTestSuite) TestPauseContainer() {
	guest := s.createContainer()
	_, _ = s.containerAction("StartContainer", guest)
	s.testContainerAction("PauseContainer", guest, "paused")
	_, _ = s.containerAction("UnpauseContainer", guest)
}

func (s *ContainerTestSuite) TestUnpauseContainer() {
	guest := s.createContainer()
	_, _ = s.containerAction("StartContainer", guest)
	_, _ = s.containerAction("PauseContainer", guest)
	s.testContainerAction("UnpauseContainer", guest, "running")
}

func (s *ContainerTestSuite) testContainerAction(action string, guest *client.Guest, finalState string) {
	tests := []struct {
		description string
		guest       *client.Guest
		expectedErr bool
	}{
		{"missing id",
			&client.Guest{}, true},
		{"invalid id",
			&client.Guest{ID: "asdf"}, true},
		{"valid id",
			guest, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		response, err := s.containerAction(action, test.guest)
		if test.expectedErr {
			s.Error(err, msg("should fail"))
		} else {
			s.NoError(err, msg("should succeed"))
			s.Equal(finalState, response.Guest.State, msg("container should be in expected state"))
		}
	}
}

func (s *ContainerTestSuite) containerAction(action string, guest *client.Guest) (*rpc.GuestResponse, error) {
	response := &rpc.GuestResponse{}
	request := &rpc.GuestRequest{
		Guest: guest,
	}
	return response, s.Client.Do("MDocker."+action, request, response)
}

func (s *ContainerTestSuite) TestDeleteContainer() {
	guest := s.createContainer()

	request := &rpc.GuestRequest{
		Guest: guest,
	}
	response := &rpc.GuestResponse{}
	s.NoError(s.Client.Do("MDocker.DeleteContainer", request, response))
}
