#PKG_NAME:=github.com/sapcc/mosquitto-exporter
PKG_NAME:=github.com/daviddetorres/mosquitto-exporter
BUILD_DIR:=bin
MOSQUITTO_EXPORTER_BINARY:=$(BUILD_DIR)/mosquitto_exporter
IMAGE := sapcc/mosquitto-exporter
VERSION=0.6.0
LDFLAGS=-s -w -X main.Version=$(VERSION) -X main.GITCOMMIT=`git rev-parse --short HEAD`
CGO_ENABLED=0
GOARCH=amd64
.PHONY: help
help:
	@echo
	@echo "Available targets:"
	@echo "  * build             - build the binary, output to $(BUILD_DIR)"
	@echo "  * linux             - build the binary, output to $(BUILD_DIR)"
	@echo "  * docker            - build docker image"

.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	# Install dependencies
	go get github.com/codegangsta/cli
	go get github.com/eclipse/paho.mqtt.golang
	go get github.com/prometheus/client_golang/prometheus
	# Build sources
	go build -o $(MOSQUITTO_EXPORTER_BINARY) -ldflags="$(LDFLAGS)" $(PKG_NAME)

linux: export GOOS=linux
linux: build

docker: linux
	docker build -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)
