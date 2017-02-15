
PKG_NAME:=github.com/sapcc/mosquitto-exporter
BUILD_DIR:=bin
MOSQUITTO_EXPORTER_BINARY:=$(BUILD_DIR)/exporter

.PHONY: help
help:
	@echo
	@echo "Available targets:"
	@echo "  * build             - build the binary, output to $(ARC_BINARY)"

.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(MOSQUITTO_EXPORTER_BINARY) -ldflags="$(LDFLAGS)" $(PKG_NAME)
