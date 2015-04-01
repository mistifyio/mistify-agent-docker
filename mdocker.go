package mdocker

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

type (
	// RPCRequest is an interface for incoming RPC requests
	RPCRequest interface {
		GetOpts() interface{}
	}

	// MDocker is the Mistify Docker subagent service
	MDocker struct {
		endpoint string
		client   *docker.Client
	}
)

// New creates a new MDocker with a docker client
func New(endpoint, tlsCertPath string) (*MDocker, error) {
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

// RequestOpts extracts the request opts into an appropriate struct
// Nested structs stored in interface{} don't convert directly, so use JSON as
// an intermediate
func (md *MDocker) RequestOpts(req RPCRequest, opts interface{}) error {
	o := req.GetOpts()
	if o == nil {
		return nil
	}

	oJSON, err := json.Marshal(o)
	if err != nil {
		return err
	}

	return json.Unmarshal(oJSON, opts)
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
