# OMNeT++ simulation distributor

## Install command line tools

```shell
go install cmd/worker/opp_edge_worker.go
go install cmd/consumer/opp_edge_run.go
go install cmd/config/opp_edge_config.go
go install cmd/broker/opp_edge_broker.go
```

## Install and run with Docker

```shell
docker pull pzierahn/omnetpp_edge
docker run --rm pzierahn/omnetpp_edge opp_edge_worker -broker 85.214.35.83 -name `hostname -s`
```

## Build and upload docker images

Build cross-platform images for amd64 and arm64.

```shell
docker buildx build \
    --push \
    --platform linux/arm64,linux/amd64 \
    --tag pzierahn/omnetpp_edge:latest .
```

> Build alternative: ```docker build -t pzierahn/omnetpp_edge .```

## Run example simulations

```shell
go run cmd/consumer/opp_edge_run.go -path ~/github/TaskletSimulator -config ~/github/TaskletSimulator/opp-edge-config.json
go run cmd/consumer/opp_edge_run.go -path ~/Desktop/tictoc -config ~/Desktop/tictoc/opp-edge-config.json
```

## Install and run broker

```shell
go install cmd/broker/opp_edge_broker.go
nohup opp_edge_broker > opp_edge_broker.log 2>&1 &
```

## Developer Notes (ignore)

Install protobuf dependencies.

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

go get -u google.golang.org/grpc
GOOS=linux GOARCH=amd64 go build cmd/consumer/opp_edge_run.go
```
