package mdocker_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	h "github.com/bakins/test-helpers"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	"github.com/mistifyio/mistify-agent-docker"
	"github.com/mistifyio/mistify-agent/rpc"
	logx "github.com/mistifyio/mistify-logrus-ext"
)

type TestClient struct {
	rpc           *rpc.Client
	ImageImported bool
	ImageID       string
	ContainerName string
}

var client TestClient

// TestMain sets up the server and RPC client before running tests
func TestMain(m *testing.M) {
	code := 0
	defer func() {
		os.Exit(code)
	}()

	var port uint = 30001
	dockerSocket := "unix:///var/run/docker.sock"
	imageServicePort := 12345

	logLevel := "warning"
	if err := logx.DefaultSetup(logLevel); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "logx.DefaultSetup",
			"level": logLevel,
		}).Fatal("failed to set up logging")
	}
	md, err := mdocker.New(dockerSocket, fmt.Sprintf(":%d", imageServicePort), "")
	if err != nil {
		log.Fatal("Can't create mdocker:", err)
	}

	go md.RunHTTP(port)
	time.Sleep(1 * time.Second)

	client = TestClient{
		ImageID: "mistify-docker-agent-test",
	}

	client.rpc, err = rpc.NewClient(port, "")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"port":  port,
		}).Error("can't create rpc client")
		code = 1
		return
	}

	imagefile, err := prepareImageTar(dockerSocket, "tauzero/test-loop")
	if imagefile != "" {
		defer os.Remove(imagefile)
	}
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"imagefile": imagefile,
		}).Error("failed to create test image tar")
		code = 1
		return
	}

	go mockImageService(imagefile, imageServicePort)
	time.Sleep(1 * time.Second)

	code = m.Run()
}

func prepareImageTar(dockerSocket, repo string) (string, error) {
	file, err := ioutil.TempFile("", "dockerTestImageTar")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to create tempfile")
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	dockerClient, err := docker.NewClient(dockerSocket)
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"dockerSocket": dockerSocket,
		}).Error("failed to create docker dockerClient")
		return file.Name(), err
	}

	repoName := "tauzero/test-loop"
	pullOpts := docker.PullImageOptions{
		Repository: repoName,
		Tag:        "latest",
	}
	if err := dockerClient.PullImage(pullOpts, docker.AuthConfiguration{}); err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"exportOpts": pullOpts,
		}).Error("failed to pull image")
		return file.Name(), err
	}
	exportOpts := docker.ExportImageOptions{
		Name:         repoName,
		OutputStream: file,
	}
	if err := dockerClient.ExportImage(exportOpts); err != nil {
		log.WithFields(log.Fields{
			"error":      err,
			"exportOpts": exportOpts,
		}).Error("failed to export image")
		return file.Name(), err
	}

	if err := dockerClient.RemoveImage(repoName); err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"repoName": repoName,
		}).Error("failed to remove image")
	}

	return file.Name(), nil
}

func mockImageService(imageFile string, port int) {
	router := mux.NewRouter()
	router.StrictSlash(true)

	router.HandleFunc("/images/{imageID}/download", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if vars["imageID"] != "mistify-docker-agent-test" {
			http.Error(w, "", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		file, _ := os.Open(imageFile)
		defer file.Close()

		io.Copy(w, file)
		return
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	if err := server.ListenAndServe(); err != nil {
		log.WithField("error", err).Error("failed to run mock image service")
	}
}

func TestGetInfo(t *testing.T) {
	req := &struct{}{}
	resp := &docker.Env{}
	h.Ok(t, client.rpc.Do("MDocker.GetInfo", req, resp))
	h.Assert(t, len(*resp) > 0, "response should not be empty")
}
