#!/usr/bin/env bash

SLARK_MYSQL_DNS='root:waf12KFkwo22@tcp(172.31.30.33:3306)/slark?charset=utf8&parseTime=true' PSWDS_MYSQL_DNS='root:waf12KFkwo22@tcp(172.31.30.33:3306)/pswds?charset=utf8&parseTime=true' go run ./cmd/gofakeit 

# docker run --rm -v ./:/test -w /test golang:1.23 sh -c "git config --global url.https://github@github.com.insteadof https://github.com && GOSUMDB=off go mod tidy && ./scripts/generator/testdata.sh"