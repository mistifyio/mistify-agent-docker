package mdocker_test

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent-docker"
	"github.com/mistifyio/mistify-agent/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MDockerTestSuite struct {
	APITestSuite
}

func TestMDockerTestSuite(t *testing.T) {
	suite.Run(t, new(MDockerTestSuite))
}

func (s *MDockerTestSuite) TestNew() {
	tests := []struct {
		description string
		endpoint    string
		tlsCertPath string
		expectedErr bool
	}{
		{"missing endpoint",
			"", "", true},
		{"bad endpoint",
			"unix:///dev/null", "", true},
		{"valid endpoint",
			"unix:///var/run/docker.sock", "", false},
		{"bad tls path",
			"unix:///var/run/docker.sock", "/dev/null", true},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)

		md, err := mdocker.New(test.endpoint, s.ImageService, test.tlsCertPath)
		if test.expectedErr {
			s.Error(err, msg("should fail"))
			s.Nil(md, msg("failure shouldn't return a client"))
		} else {
			s.NoError(err, msg("should succeed"))
			s.NotNil(md, msg("success should return a client"))
		}
	}
}

func (s *MDockerTestSuite) TestRequestOpts() {
	var listOpts docker.ListContainersOptions
	request := &rpc.ContainerRequest{
		Opts: docker.ListContainersOptions{
			Limit: 5,
		},
	}
	s.NoError(s.MDocker.RequestOpts(request, &listOpts))
	s.True(assert.ObjectsAreEqualValues(request.Opts, listOpts))

	var saveOpts docker.CommitContainerOptions
	request = &rpc.ContainerRequest{
		Opts: docker.CommitContainerOptions{
			Author: "asdf",
		},
	}
	s.NoError(s.MDocker.RequestOpts(request, &saveOpts))
	s.True(assert.ObjectsAreEqualValues(request.Opts, saveOpts))
}

func (s *MDockerTestSuite) TestGetInfo() {
	request := &struct{}{}
	response := &docker.Env{}
	s.NoError(s.Client.Do("MDocker.GetInfo", request, response))
	s.True(len(*response) > 0)
}
