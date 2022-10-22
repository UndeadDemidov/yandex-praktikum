PROJECT=$(shell basename "$(PWD)")

# Go variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

# Docker variables
IMAGENAME=ya_pract
HTTP_PORT_EXPOSE=8080
DOCKER_HOST=docker.for.mac.host.internal

# Project variables
LINT=staticlint
MAIN=shortener
LINT_PATH=cmd/$(LINT)
MAIN_PATH=cmd/$(MAIN)
HTTP_PORT=8080
BASE_URL="http://localhost:$(HTTP_PORT)/"
SERVER_ADDRESS=":$(HTTP_PORT)"
FILE_STORAGE_PATH=./test/data/test.tst
DATABASE_DSN="postgres://postgres:postgres@localhost:5432/ya_pract?sslmode=disable"
DATABASE_DSN_DOCKER="postgres://postgres:postgres@$(DOCKER_HOST):5432/ya_pract?sslmode=disable"

# Build variables
BUILD_VERSION=`git describe --tags`
BUILD_DATE=`date +%FT%T%z`
LDFLAGS=-ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE)"

gen:
	@echo "  >  Generating code for $(GOBASE)/..."
	@go generate $(GOBASE)/...

doc:
	@godoc -http=:8081 & open http://localhost:8081/pkg/?m=all

tests: gen
	@echo "  >  Running tests for $(GOBASE)/..."
	@go test $(GOBASE)/...

cover: gen
	@echo "  >  Running coverage for $(GOBASE)/..."
	@go test -coverprofile cover.out $(GOBASE)/...
	@go tool cover -html=cover.out

check:
	@echo "  >  Running staticcheck for $(GOBASE)/..."
	@staticcheck $(GOBASE)/...

lint: lint-build
	@echo "  >  Running $(LINT) for $(GOBASE)/..."
	@$(GOBIN)/$(LINT) $(GOBASE)/...

lint-build:
	@echo "  >  Building linters $(GOBASE)/$(LINT_PATH)"
	@GOBIN=$(GOBIN) go build -o $(GOBIN)/$(LINT) $(GOBASE)/$(LINT_PATH)

go-build:
	@echo "  >  Building binary for $(GOBASE)/$(MAIN_PATH)"
	@GOBIN=$(GOBIN) go build $(LDFLAGS) -o $(GOBIN)/$(MAIN) $(GOBASE)/$(MAIN_PATH)

# -ldflags "-X main.Version=v1.0.1 -X 'main.BuildTime=$(date +'%Y/%m/%d %H:%M:%S')'"
go-run-mem:
	@GOBIN=$(GOBIN) \
	BASE_URL=$(BASE_URL) \
	SERVER_ADDRESS=$(SERVER_ADDRESS) \
	go run $(LDFLAGS) ./$(MAIN_PATH)

go-run-file:
	@GOBIN=$(GOBIN) \
	BASE_URL=$(BASE_URL) \
	SERVER_ADDRESS=$(SERVER_ADDRESS) \
	FILE_STORAGE_PATH=$(FILE_STORAGE_PATH) \
	go run $(LDFLAGS) ./$(MAIN_PATH)

go-run-db:
	@GOBIN=$(GOBIN) \
	BASE_URL=$(BASE_URL) \
	SERVER_ADDRESS=$(SERVER_ADDRESS) \
	DATABASE_DSN=$(DATABASE_DSN) \
	go run $(LDFLAGS) ./$(MAIN_PATH) -s

go-run-cfg:
	@GOBIN=$(GOBIN) \
	go run $(LDFLAGS) ./$(MAIN_PATH) -c shortener.json

build:
	@docker build \
	--build-arg APP=$(PROJECT) \
	--build-arg MAIN_PATH=$(MAIN_PATH) \
	-t $(IMAGENAME) .

stop:
	@docker stop $(PROJECT)

start-mem:build
	docker run --rm --name $(PROJECT) \
	-e BASE_URL=$(BASE_URL) \
	-e SERVER_ADDRESS=$(SERVER_ADDRESS) \
	-p $(HTTP_PORT):$(HTTP_PORT_EXPOSE)/tcp -p $(HTTP_PORT):$(HTTP_PORT_EXPOSE)/udp \
	$(IMAGENAME)

start-file:build
	docker run --rm --name $(PROJECT) \
	-e BASE_URL=$(BASE_URL) \
	-e SERVER_ADDRESS=$(SERVER_ADDRESS) \
	-e FILE_STORAGE_PATH=$(FILE_STORAGE_PATH) \
	-p $(HTTP_PORT):$(HTTP_PORT_EXPOSE)/tcp -p $(HTTP_PORT):$(HTTP_PORT_EXPOSE)/udp \
	$(IMAGENAME)

start-db:build
	docker run --rm --name $(PROJECT) \
	-e BASE_URL=$(BASE_URL) \
	-e SERVER_ADDRESS=$(SERVER_ADDRESS) \
	-e DATABASE_DSN=$(DATABASE_DSN_DOCKER) \
	-p $(HTTP_PORT):$(HTTP_PORT_EXPOSE)/tcp -p $(HTTP_PORT):$(HTTP_PORT_EXPOSE)/udp \
	$(IMAGENAME)
