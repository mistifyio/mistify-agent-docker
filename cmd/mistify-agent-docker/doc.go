/*
mistify-agent-libvirt runs the subagent and HTTP API.

Usage

The following arguments are understood:

	$ mistify-agent-docker -h
	Usage of mistify-agent-docker:
	-d, --docker-cert-path="": docker tls cert path
	-e, --endpoint="unix:///var/run/docker.sock": docker endpoint
	-l, --log-level="warning": log level: debug/info/warning/error/critical/fatal
	-p, --port=30001: listen port
*/
package main
