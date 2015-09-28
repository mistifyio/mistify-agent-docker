package mdocker_test

import (
	"testing"

	"github.com/mistifyio/mistify-agent/rpc"
	"github.com/stretchr/testify/suite"
)

type ImageTestSuite struct {
	APITestSuite
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
}

func (s *ImageTestSuite) TestListImages() {
	_ = s.loadImage()

	request := &rpc.ImageRequest{}
	response := &rpc.ImageResponse{}
	s.NoError(s.Client.Do("MDocker.ListImages", request, response))
	images := response.Images
	s.Len(images, 1)
	s.Equal(s.ImageID, images[0].ID)
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
