PREFIX := /opt/mistify
SBIN_DIR=$(PREFIX)/sbin
ETC_DIR=$(PREFIX)/etc

cmd/mistify-agent-docker/mistify-agent-docker: cmd/mistify-agent-docker/main.go
	cd cmd/mistify-agent-docker && \
	go get -v && \
	go build -v

clean:
	cd cmd/mistify-agent-docker && \
	go clean -v

install: cmd/mistify-agent-docker/mistify-agent-docker
	install -D cmd/mistify-agent-docker/mistify-agent-docker $(DESTDIR)$(SBIN_DIR)/mistify-agent-docker
