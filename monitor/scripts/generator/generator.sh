#!/usr/bin/env bash

# generate *.pb.go, *_grpc.pb.go
# protoc --go_out=./ --go-grpc_out=./ ./api/*.proto
# inject go tags
# protoc-go-inject-tag -input="./api/*.pb.go"

# generate models from local db
gentool -dsn 'root:waf12@tcp(127.0.0.1:3306)/monitor?charset=utf8&parseTime=true' -onlyModel -outPath './internal/model'

# format go codes
go fmt ./...