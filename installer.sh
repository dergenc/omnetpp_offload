#!/bin/bash

#export PATH=$PATH:/usr/local/go/bin:~/go/bin

# install all dependencies locally
#go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
#go get -u google.golang.org/grpc

# load all local code to create executables
go install cmd/worker/opp_offload_worker.go
go install cmd/run/opp_offload_run.go
go install cmd/config/opp_offload_config.go
go install cmd/broker/opp_offload_broker.go

# create docker container locally fetching the OMNeT++ from docker registries
# change architecture if necessary
#mkdir docker-build
#sudo docker buildx build --push --platform linux/amd64 --progress=plain --tag dalpergenc/omnetpp_provider:latest .