package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify-agent-docker"
	flag "github.com/ogier/pflag"
)

func main() {
	// Handle cli flags
	var port uint
	var endpoint, logLevel, tlsCertPath string
	var tls bool
	flag.UintVarP(&port, "port", "p", 30001, "listen port")
	flag.StringVarP(&endpoint, "endpoint", "e", "unix:///var/run/docker.sock", "docker endpoint")
	flag.BoolVarP(&tls, "tls", "t", false, "use TLS")
	flag.StringVarP(&tlsCertPath, "docker-cert-path", "d", os.Getenv("DOCKER_CERT_PATH"), "docker tls cert path")
	flag.StringVarP(&logLevel, "log-level", "l", "warning", "log level: debug/info/warning/error/critical/fatal")
	flag.Parse()

	// Set up logging
	log.SetFormatter(&log.JSONFormatter{})
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"func":  "log.ParseLevel",
		}).Fatal("invalid log level")
	}
	log.SetLevel(level)

	// Prepare docker connection configuration
	if !tls {
		tlsCertPath = ""
	}
	log.WithFields(log.Fields{
		"endpoint":   endpoint,
		"tlsEnabled": tls,
		"certPath":   tlsCertPath,
	}).Info("docker configuration")

	// Create the MDocker instance
	md, err := mdocker.NewMDocker(endpoint, tlsCertPath)
	if err != nil {
		os.Exit(1)
	}

	// Create and run the HTTP server
	_ = md.RunHTTP(port)
}
