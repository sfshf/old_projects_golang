#!/usr/bin/env bash

# generate *.pb.go, *_grpc.pb.go
protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf --go_out=./ --go-grpc_out=./ ./api/alchemist.proto
protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf --go_out=./ --go-grpc_out=./ ./api/alchemist_notification.proto
# inject go tags
protoc-go-inject-tag -input="./api/*.pb.go"
protoc-go-inject-tag -input="./api/notification/*.pb.go"

# generate models from local db
gentool -dsn 'root:waf12@tcp(127.0.0.1:3306)/alchemist?charset=utf8&parseTime=true' -onlyModel -outPath './internal/pkg/model'

# format go codes
go fmt ./...