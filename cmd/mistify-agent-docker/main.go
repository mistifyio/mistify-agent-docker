package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent-docker"
	logx "github.com/mistifyio/mistify-logrus-ext"
	flag "github.com/ogier/pflag"
)

func main() {
	// Handle cli flags
	var port uint
	var endpoint, logLevel, tlsCertPath string
	flag.UintVarP(&port, "port", "p", 30001, "listen port")
	flag.StringVarP(&endpoint, "endpoint", "e", "unix:///var/run/docker.sock", "docker endpoint")
	flag.StringVarP(&tlsCertPath, "docker-cert-path", "d", os.Getenv("DOCKER_CERT_PATH"), "docker tls cert path")
	flag.StringVarP(&logLevel, "log-level", "l", "warning", "log level: debug/info/warning/error/critical/fatal")
	flag.Parse()

	// Set up logging
	if err := logx.DefaultSetup(logLevel); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "logx.DefaultSetup",
		}).Fatal("Could not set up logging")
	}

	// Prepare docker connection configuration
	log.WithFields(log.Fields{
		"port":     port,
		"logLevel": logLevel,
		"docker": map[string]interface{}{
			"endpoint": endpoint,
			"certPath": tlsCertPath,
		},
	}).Info("configuration")

	// Create the MDocker instance
	md, err := mdocker.New(endpoint, tlsCertPath)
	if err != nil {
		os.Exit(1)
	}

	// Create and run the HTTP server
	_ = md.RunHTTP(port)
}
