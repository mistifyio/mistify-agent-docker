package mdocker_test

import (
	"log"
	"os"
	"testing"
	"time"

	h "github.com/bakins/test-helpers"
	"github.com/fsouza/go-dockerclient"
	"github.com/mistifyio/mistify-agent-docker"
	"github.com/mistifyio/mistify-agent/rpc"
)

type TestClient struct {
	rpc       *rpc.Client
	ImageName string
	ImageID   string
}

var client TestClient

// TestMain sets up the server and RPC client before running tests
func TestMain(m *testing.M) {
	var port uint = 30001
	md, err := mdocker.NewMDocker("unix:///var/run/docker.sock", "")
	if err != nil {
		log.Fatal("Can't create mdocker:", err)
	}

	go md.RunHTTP(port)
	time.Sleep(1 * time.Second)

	client = TestClient{
		ImageName: "busybox",
	}

	client.rpc, err = rpc.NewClient(port, "")
	if err != nil {
		log.Fatal("Can't create rpc client:", err)
	}

	os.Exit(m.Run())
}

func TestGetInfo(t *testing.T) {
	req := &struct{}{}
	resp := &docker.Env{}
	h.Ok(t, client.rpc.Do("MDocker.GetInfo", req, resp))
}
