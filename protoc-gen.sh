#!/bin/sh
protoc --proto_path=protos --go_out=pb --go-grpc_out=pb protos/*.proto 