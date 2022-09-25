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

gen:
	@echo "  >  Generating code for $(GOBASE)/..."
	@go generate $(GOBASE)/...

doc:
	@godoc -http=:8081 & open http://localhost:8081/pkg/?m=all

tests:
	@echo "  >  Running tests for $(GOBASE)/..."
	@go test $(GOBASE)/...

cover:
	@echo "  >  Running coverage for $(GOBASE)/..."
	@go test -coverprofile cover.out $(GOBASE)/...
	@go tool cover -html=cover.out

check:
	@echo "  >  Running staticcheck for $(GOBASE)/..."
	@staticcheck $(GOBASE)/...

lint: lint-build
	@echo "  >  Running $(LINT) for $(GOBASE)/..."
	@$(GOBIN)/$(LINT) -help $(GOBASE)/...

lint-build:
	@echo "  >  Building linters $(GOBASE)/$(LINT_PATH)"
	@GOBIN=$(GOBIN) go build -o $(GOBIN)/$(LINT) $(GOBASE)/$(LINT_PATH)

go-build:
	@echo "  >  Building binary for $(GOBASE)/$(MAIN_PATH)"
	@GOBIN=$(GOBIN) go build -o $(GOBIN)/$(MAIN) $(GOBASE)/$(MAIN_PATH)

go-run-mem:
	@GOBIN=$(GOBIN) \
	BASE_URL=$(BASE_URL) \
	SERVER_ADDRESS=$(SERVER_ADDRESS) \
	go run ./$(MAIN_PATH)

go-run-file:
	@GOBIN=$(GOBIN) \
	BASE_URL=$(BASE_URL) \
	SERVER_ADDRESS=$(SERVER_ADDRESS) \
	FILE_STORAGE_PATH=$(FILE_STORAGE_PATH) \
	go run ./$(MAIN_PATH)

go-run-db:
	@GOBIN=$(GOBIN) \
	BASE_URL=$(BASE_URL) \
	SERVER_ADDRESS=$(SERVER_ADDRESS) \
	DATABASE_DSN=$(DATABASE_DSN) \
	go run ./$(MAIN_PATH)

build:
	@docker build \
	--build-arg APP=$(PROJECT) \
	--build-arg MAIN_PATH=$(MAIN_PATH) \
	-t $(IMAGENAME) .

stop:
	@docker stop $(PROJECT)

start-mem:
	docker run --rm --name $(PROJECT) \
	-e BASE_URL=$(BASE_URL) \
	-e SERVER_ADDRESS=$(SERVER_ADDRESS) \
	$(IMAGENAME)

start-file:
	docker run --rm --name $(PROJECT) \
	-e BASE_URL=$(BASE_URL) \
	-e SERVER_ADDRESS=$(SERVER_ADDRESS) \
	-e FILE_STORAGE_PATH=$(FILE_STORAGE_PATH) \
	$(IMAGENAME)

start-db:
	docker run --rm --name $(PROJECT) \
	-e BASE_URL=$(BASE_URL) \
	-e SERVER_ADDRESS=$(SERVER_ADDRESS) \
	-e DATABASE_DSN=$(DATABASE_DSN_DOCKER) \
	-p $(HTTP_PORT):$(HTTP_PORT_EXPOSE)/tcp -p $(HTTP_PORT):$(HTTP_PORT_EXPOSE)/udp \
	$(IMAGENAME)
