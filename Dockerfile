# Start from a Apline image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:alpine as builder
ARG APP
ARG MAIN_PATH
RUN echo "app name is $APP"
RUN echo "main.go path is $MAIN_PATH"

# Alpine has not git installed - install it b4 run "go get" command
RUN apk update && apk add git

# Copy the local package files to the container's workspace.
RUN mkdir -p /opt/src/github.com/UndeadDemidov/$APP
WORKDIR /opt/src/github.com/UndeadDemidov/$APP

# Some speedup from  https://medium.com/@petomalina/using-go-mod-download-to-speed-up-golang-docker-builds-707591336888
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/$APP ./$MAIN_PATH

# Start from scratch.
FROM scratch
ARG APP
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/$APP /go/bin/entry

ENTRYPOINT ["/go/bin/entry"]