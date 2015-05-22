package main

import (
	"net"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
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
	flag.StringVarP(&imageService, "image-service", "i", "images.service.lochness.local", "image service")
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
	// Using strings.Split instead of net.SplitHostPort since the latter errors
	// if no port is present and it doesn't provide any error type checking
	// convenience methods
	// TODO: not ipv6 compatible
	imageServiceParts := strings.Split(imageService, ":")
	partsLength := len(imageServiceParts)
	// Empty or too many colons
	if partsLength == 0 || partsLength > 2 {
		log.WithField("imageService", imageService).Fatal("invalid image-service value")
	}

	// Try to lookup port if only host/service is provided
	if partsLength == 1 || imageServiceParts[1] == "" {
		_, addrs, err := net.LookupSRV("", "", imageService)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"func":  "net.LookupSRV",
			}).Fatal("srv lookup failed")
		}
		if len(addrs) == 0 {
			log.WithField("imageService", imageService).Fatal("invalid image-service value")
		}
		imageServiceParts[1] = strconv.FormatUint(uint64(addrs[0].Port), 10)
	}
	imageService = net.JoinHostPort(imageServiceParts[0], imageServiceParts[1])

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
