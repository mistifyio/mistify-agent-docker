package mdocker_test

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/rpc"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

type ImageTestSuite struct {
	APITestSuite
}

func (s ImageTestSuite) TearDownTest() {
	s.APITestSuite.TearDownTest()

	// Clean up docker image
	if err := s.Docker.RemoveImage(s.ImageID); err != nil {
		log.WithField("error", err).Error("failed to remove image")
	}
}

func TestImageTestSuite(t *testing.T) {
	suite.Run(t, new(ImageTestSuite))
}

func (s *ImageTestSuite) TestLoadImage() {
	tests := []struct {
		description string
		requestID   string
		expectedErr bool
	}{
		{"missing id", "", true},
		{"bad id", "asdf", true},
		{"valid id", s.ImageID, false},
		{"valid gzip id", "gzipID", false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		response := &rpc.ImageResponse{}
		request := &rpc.ImageRequest{
			ID: test.requestID,
		}
		err := s.Client.Do("MDocker.LoadImage", request, response)
		if test.expectedErr {
			s.Error(err, msg("should error"))
		} else {
			s.NoError(err, msg("should not error"))
			s.Len(response.Images, 1)
			s.Equal(test.requestID, response.Images[0].ID, msg("should be correct image"))
		}
	}

	if err := s.Docker.RemoveImage("gzipID"); err != nil {
		log.WithField("error", err).Error("failed to remove image")
	}
}

func (s *ImageTestSuite) TestListImages() {
	_ = s.loadImage()

	request := &rpc.ImageRequest{}
	response := &rpc.ImageResponse{}
	s.NoError(s.Client.Do("MDocker.ListImages", request, response))
	found := false
	for _, image := range response.Images {
		if s.ImageID == image.ID {
			found = true
		}
		s.NotEmpty(uuid.Parse(image.ID))
	}
	s.True(found)
}

func (s *ImageTestSuite) TestGetImage() {
	_ = s.loadImage()

	tests := []struct {
		description string
		requestID   string
		expectedErr bool
	}{
		{"missing id", "", true},
		{"bad id", "asdf", true},
		{"valid id", s.ImageID, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		response := &rpc.ImageResponse{}
		request := &rpc.ImageRequest{
			ID: test.requestID,
		}
		err := s.Client.Do("MDocker.GetImage", request, response)
		if test.expectedErr {
			s.Error(err, msg("should error"))
		} else {
			s.NoError(err, msg("should not error"))
			s.Len(response.Images, 1)
			s.Equal(test.requestID, response.Images[0].ID, msg("should be correct image"))
		}
	}
}

func (s *ImageTestSuite) TestDeleteImage() {
	_ = s.loadImage()

	tests := []struct {
		description string
		requestID   string
		expectedErr bool
	}{
		{"missing id", "", true},
		{"bad id", "asdf", true},
		{"valid id", s.ImageID, false},
	}

	for _, test := range tests {
		msg := testMsgFunc(test.description)
		response := &rpc.ImageResponse{}
		request := &rpc.ImageRequest{
			ID: test.requestID,
		}
		err := s.Client.Do("MDocker.DeleteImage", request, response)
		if test.expectedErr {
			s.Error(err, msg("should error"))
		} else {
			s.NoError(err, msg("should not error"))
			s.Len(response.Images, 1)
			s.Equal(test.requestID, response.Images[0].ID, msg("should be correct image"))
		}
	}
}
