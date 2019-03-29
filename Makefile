PKG_NAME:=github.com/sapcc/mosquitto-exporter
BUILD_DIR:=artifacts
MOSQUITTO_EXPORTER_BINARY_UNPACKED:=$(BUILD_DIR)/svc-unpacked
MOSQUITTO_EXPORTER_BINARY:=$(BUILD_DIR)/svc
IMAGE := sapcc/mosquitto-exporter
VERSION=0.5.0
LDFLAGS=-s -w -X main.Version=$(VERSION) -X main.GITCOMMIT=`git rev-parse --short HEAD`
CGO_ENABLED=0
GOARCH=amd64
.PHONY: help
help:
	@echo
	@echo "Available targets:"
	@echo "  * build             - build the binary, output to $(ARC_BINARY)"
	@echo "  * linux             - build the binary, output to $(ARC_BINARY)"
	@echo "  * docker            - build docker image"

.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(MOSQUITTO_EXPORTER_BINARY_UNPACKED) -ldflags="$(LDFLAGS)" $(PKG_NAME)
	rm -rf $(MOSQUITTO_EXPORTER_BINARY)
	upx -q -o $(MOSQUITTO_EXPORTER_BINARY) $(MOSQUITTO_EXPORTER_BINARY_UNPACKED)

linux: export GOOS=linux
linux: build

docker: linux
	docker build -t $(IMAGE):$(VERSION) .

push:
	docker push $(IMAGE):$(VERSION)
