package mdocker

import (
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent/rpc"
	"github.com/tylerb/graceful"
)

// RunHTTP creates and runs the RPC HTTP server
func (md *MDocker) RunHTTP(port uint) *graceful.Server {
	s, err := rpc.NewServer(port)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to create rpc server")
		return nil
	}
	if err := s.RegisterService(md); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to register mdocker service")
		return nil
	}

	server := &graceful.Server{
		Timeout: 5 * time.Second,
		Server:  s.HTTPServer,
	}
	go listenAndServe(server)
	return server
}

func listenAndServe(server *graceful.Server) {
	if err := server.ListenAndServe(); err != nil {
		// Ignore the error from closing the listener, which is involved in the
		// graceful shutdown
		if !strings.Contains(err.Error(), "use of closed network connection") {
			log.WithField("error", err).Fatal("server error")
		}
	}
}
