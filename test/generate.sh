#!/usr/bin/env bash

protoc -I . --go_out=paths=source_relative:. test/test.proto

protoc -I . -I ../googleapis --go_out . --go_opt paths=source_relative \
--go-grpc_out . --go-grpc_opt paths=source_relative \
--grpc-gateway_out . --grpc-gateway_opt paths=source_relative test/hello_world.proto
