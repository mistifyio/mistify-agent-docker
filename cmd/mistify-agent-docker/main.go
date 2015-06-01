package main

import (
	"fmt"
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/mistifyio/lochness/pkg/hostport"
	"github.com/mistifyio/mistify-agent-docker"
	logx "github.com/mistifyio/mistify-logrus-ext"
	flag "github.com/ogier/pflag"
)

func main() {
	// Handle cli flags
	var port uint
	var endpoint, logLevel, tlsCertPath, imageService string
	flag.UintVarP(&port, "port", "p", 30001, "listen port")
	flag.StringVarP(&endpoint, "endpoint", "e", "unix:///var/run/docker.sock", "docker endpoint")
	flag.StringVarP(&tlsCertPath, "docker-cert-path", "d", os.Getenv("DOCKER_CERT_PATH"), "docker tls cert path")
	flag.StringVarP(&imageService, "image-service", "i", "images.service.lochness.local", "image service. srv query used to find port if not specified")
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

	// Parse image service and do any necessary lookups
	iHost, iPort, err := hostport.Split(imageService)
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err,
			"imageService": imageService,
			"func":         "hostport.Split",
		}).Fatal("host port split failed")
	}

	// Try to lookup port if only host/service is provided
	if iPort == "" {
		_, addrs, err := net.LookupSRV("", "", iHost)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"func":  "net.LookupSRV",
			}).Fatal("srv lookup failed")
		}
		if len(addrs) == 0 {
			log.WithField("imageService", iHost).Fatal("invalid host value")
		}
		iPort = fmt.Sprintf("%d", addrs[0].Port)
	}
	imageService = net.JoinHostPort(iHost, iPort)

	// Create the MDocker instance
	md, err := mdocker.New(endpoint, imageService, tlsCertPath)
	if err != nil {
		os.Exit(1)
	}

	// Create and run the HTTP server
	if err := md.RunHTTP(port); err != nil {
		os.Exit(1)
	}
}
