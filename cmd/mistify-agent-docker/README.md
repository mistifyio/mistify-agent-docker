# mistify-agent-docker

[![mistify-agent-docker](https://godoc.org/github.com/mistifyio/mistify-agent-docker/cmd/mistify-agent-docker?status.png)](https://godoc.org/github.com/mistifyio/mistify-agent-docker/cmd/mistify-agent-docker)

mistify-agent-docker runs the subagent and HTTP API.


### Usage

The following arguments are understood:

    $ mistify-agent-docker -h
    Usage of mistify-agent-docker:
    -d, --docker-cert-path="": docker tls cert path
    -e, --endpoint="unix:///var/run/docker.sock": docker endpoint
    -i, --image-service="images.service.lochness.local": image service. srv query used to find port if not specified
    -l, --log-level="warning": log level: debug/info/warning/error/critical/fatal
    -p, --port=30001: listen port


--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
