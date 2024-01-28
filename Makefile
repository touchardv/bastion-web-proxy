BINARY := bastion-web-proxy
BUILD_DIR := $(shell pwd)/build
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
SOURCES = $(shell find . -name '*.go')

.PHONY: build
build: $(BUILD_DIR)/$(BINARY)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(BUILD_DIR)/$(BINARY): $(BUILD_DIR) $(SOURCES)
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/bastion-web-proxy

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

release: $(BUILD_DIR)/$(BINARY)
	cd $(BUILD_DIR) && shasum -a 512 $(BINARY) >$(BINARY).shasum512
	cd $(BUILD_DIR) && tar cvzf $(BINARY)-$(GOOS)-$(GOARCH).tgz $(BINARY) $(BINARY).shasum512

run: $(BUILD_DIR)/$(BINARY)
	$(BUILD_DIR)/$(BINARY) --log-level debug

test:
	go test -v -cover -timeout 10s ./...
