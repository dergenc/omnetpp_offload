#!/bin/bash

# run a broker
nohup opp_offload_broker > opp_offload_broker.log 2>&1 &

# run a provider
docker run --rm dalpergenc/omnetpp_provider opp_offload_worker -broker 127.0.0.1 -name `hostname -s` 2>&1 &

docker run --rm pzierahn/omnetpp_offload opp_offload_worker -broker 127.0.0.1 -name `hostname -s` 2>&1 &

# run a consumer 
go run cmd/run/opp_offload_run.go -path evaluation/tictoc