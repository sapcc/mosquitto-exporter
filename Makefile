PKG_NAME:=github.com/sapcc/mosquitto-exporter
BUILD_DIR:=bin
MOSQUITTO_EXPORTER_BINARY:=$(BUILD_DIR)/mosquitto_exporter
IMAGE := sapcc/mosquitto-exporter
VERSION=0.6.0
LDFLAGS=-s -w -X main.Version=$(VERSION) -X main.GITCOMMIT=`git rev-parse --short HEAD`
.PHONY: help
help:
	@echo
	@echo "Available targets:"
	@echo "  * build             - build the binary, output to $(BUILD_DIR)"
	@echo "  * linux             - build the binary, output to $(BUILD_DIR)"
	@echo "  * docker            - build docker image"

.PHONY: build
build: export CGO_ENABLED=0
build:
	@mkdir -p $(BUILD_DIR)
	# Build sources
	go build -o $(MOSQUITTO_EXPORTER_BINARY) -ldflags="$(LDFLAGS)" $(PKG_NAME)

linux: export GOOS=linux
linux: build

docker:
	docker build -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)
