@echo off

set PATH=E:\Programs\protobuf\bin;%PATH%

protoc --go_out=. --go_opt=paths=import --go-grpc_out=. --go-grpc_opt=paths=import middleware/class.proto
protoc --go_out=. --go_opt=paths=import --go-grpc_out=. --go-grpc_opt=paths=import middleware/hasq.proto
go build -o .bin/class.exe pet/services/cmd/class
go build -o .bin/hasq.exe pet/services/cmd/hasq