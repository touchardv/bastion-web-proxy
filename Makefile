BUILD_DIR := $(shell pwd)/build
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
SOURCES := $(shell find . -name '*.go')

BINARY := bastion-web-proxy-$(GOOS)-$(GOARCH)
IMAGE := quay.io/touchardv/bastion-web-proxy
LD_ARGS :=
TAG := latest

ifeq ($(GOARCH), arm)
 DOCKER_BUILDX_PLATFORM := linux/arm/v7
else ifeq ($(GOARCH), arm64)
 DOCKER_BUILDX_PLATFORM := linux/arm64/v8
else ifeq ($(GOARCH), amd64)
 DOCKER_BUILDX_PLATFORM := linux/amd64
endif

.PHONY: build
build: $(BUILD_DIR)/$(BINARY)

build-image: $(BUILD_DIR)/$(BINARY)
	docker buildx build --progress plain \
	--platform $(DOCKER_BUILDX_PLATFORM) \
	--tag $(IMAGE):$(TAG) --load -f deployment/Dockerfile .

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(BUILD_DIR)/$(BINARY): $(BUILD_DIR) $(SOURCES)
	go build $(LD_ARGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/bastion-web-proxy

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	docker image rm -f $(IMAGE)/$(TAG)

release: $(BUILD_DIR)/$(BINARY)
	cd $(BUILD_DIR) && shasum -a 512 $(BINARY) >$(BINARY).shasum512
	cd $(BUILD_DIR) && tar cvzf $(BINARY)-$(GOOS)-$(GOARCH).tgz $(BINARY) $(BINARY).shasum512

run: $(BUILD_DIR)/$(BINARY)
	$(BUILD_DIR)/$(BINARY) --log-level debug

run-image:
	docker run -it --rm -v `pwd`:/etc/bastion-web-proxy --name my-bastion $(IMAGE):$(TAG)

test:
	go test -v -cover -timeout 10s ./...
