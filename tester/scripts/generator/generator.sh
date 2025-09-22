#!/usr/bin/env bash

# generate *.pb.go, *_grpc.pb.go
protoc --go_out=./ --go-grpc_out=./ ./api/*.proto
# inject go tags
protoc-go-inject-tag -input="./api/*.pb.go"

# format go codes
go fmt ./...