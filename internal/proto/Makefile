MAKEFILE_PATH := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

BLACK = "\033[30m"
RED  =  "\033[31m"
GREEN = "\033[32m"
YELLOW = "\033[33m"
BLUE  = "\033[34m"
PURPLE = "\033[35m"
CYAN  = "\033[36m"
WHITE = "\033[37m"
CLOSE_COLOR = "\033[0m"

git_tag=$(shell git describe --long --tags --dirty --tags --always)
git_commit=$(shell git rev-parse --short HEAD)
built=$(shell date '+%Y%m%d')

GO := go

.PHONY: all
all: proto

.PHONY: protoset
protoset:
	@echo -e $(PURPLE)">>>  init protoset ..."$(CLOSE_COLOR)
	protoc -I . -I${MAKEFILE_PATH}/googleapis \
                --include_imports --descriptor_set_out=./bundle.pb.desc *.proto

.PHONY: proto
proto: protoset
	@echo -e $(PURPLE)">>>  init proto ..."$(CLOSE_COLOR)
	protoc -I . -I${MAKEFILE_PATH}/googleapis \
                --go_out=paths=source_relative:. \
                --go-grpc_out paths=source_relative:. \
                --grpc-gateway_out paths=source_relative:. \
                --grpc-gateway_opt logtostderr=true \
                --grpc-gateway_opt generate_unbound_methods=true echo_service.proto
