package mdocker

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

type (
	// MDocker is the Mistify Docker subagent service
	MDocker struct {
		endpoint string
		client   *docker.Client
	}
)

// NewMDocker creates a new MDocker with a docker client
func NewMDocker(endpoint, tlsCertPath string) (*MDocker, error) {
	// Create a new Docker client
	var client *docker.Client
	var err error
	if tlsCertPath == "" {
		client, err = docker.NewClient(endpoint)
		if err != nil {
			log.WithFields(log.Fields{
				"error":    err,
				"endpoint": endpoint,
				"func":     "docker.NewClient",
			}).Error("failed to create docker client")
			return nil, err
		}
	} else {
		ca := fmt.Sprintf("%s/ca.pem", tlsCertPath)
		cert := fmt.Sprintf("%s/cert.pem", tlsCertPath)
		key := fmt.Sprintf("%s/key.pem", tlsCertPath)
		client, err = docker.NewTLSClient(endpoint, cert, key, ca)
		if err != nil {
			log.WithFields(log.Fields{
				"error":    err,
				"endpoint": endpoint,
				"ca":       ca,
				"cert":     cert,
				"key":      key,
				"func":     "docker.NewTLSClient",
			}).Error("failed to create docker client")
			return nil, err
		}
	}

	// Make sure we can actually communicate with Docker
	if err := client.Ping(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "client.Ping",
		}).Error("failed to ping docker server")
		return nil, err
	}

	return &MDocker{
		endpoint: endpoint,
		client:   client,
	}, nil
}

// GetInfo provides general information about the system from Docker
func (md *MDocker) GetInfo(h *http.Request, request *struct{}, response *docker.Env) error {
	info, err := md.client.Info()
	if err != nil {
		return err
	}
	*response = *info
	return nil
}
