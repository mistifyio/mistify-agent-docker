package mdocker

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/rpc"
)

type (
	// ErrorHTTPCode should be used for errors resulting from an http response
	// code not matching the expected code
	ErrorHTTPCode struct {
		Expected int
		Code     int
		Source   string
	}
)

// Error returns a string error message
func (e ErrorHTTPCode) Error() string {
	return fmt.Sprintf("unexpected http response code: expected %d, received %d, url: %s", e.Expected, e.Code, e.Source)
}

// RunHTTP creates and runs the RPC HTTP server
func (md *MDocker) RunHTTP(port uint) error {
	server, err := rpc.NewServer(port)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "mdocker.RunHTTP",
		}).Error("failed to create rpc server")
		return err
	}
	if err := server.RegisterService(md); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "server.RegisterService",
		}).Error("failed to register mdocker service")
		return err
	}

	if err := server.ListenAndServe(); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "server.ListenAndServe",
		}).Error("failed to run server")
		return err
	}
	return nil
}
