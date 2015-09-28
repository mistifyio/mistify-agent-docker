package mdocker_test

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/suite"
)

type MDockerTestSuite struct {
	APITestSuite
}

func TestMDockerTestSuite(t *testing.T) {
	suite.Run(t, new(MDockerTestSuite))
}

func (s *MDockerTestSuite) TestGetInfo() {
	request := &struct{}{}
	response := &docker.Env{}
	s.NoError(s.Client.Do("MDocker.GetInfo", request, response))
	s.True(len(*response) > 0)
}
