#!/bin/sh

export PATH=~/go/bin:"$PATH"
#go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
#go install google.golang.org/grpc/cmd/protoc-gen-go@latest
protoc --go_out=. --go_opt=paths=import --go-grpc_out=. --go-grpc_opt=paths=import middleware/class.proto
protoc --go_out=. --go_opt=paths=import --go-grpc_out=. --go-grpc_opt=paths=import middleware/hasq.proto
go build -o .bin/class pet/services/cmd/class
go build -o .bin/hasq pet/services/cmd/hasq