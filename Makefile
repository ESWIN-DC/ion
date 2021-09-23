GO_LDFLAGS="-s -w -X 'main.goversion=`go version`' -X 'main.buildstamp=`date -u "+%Y-%m-%d_%I:%M:%S%p"`' -X 'main.githash=`git describe --tags 2>/dev/null`'"
GO_TESTPKGS:=$(shell go list ./... | grep -v cmd | grep -v conf | grep -v node)
GO_COVERPKGS:=$(shell echo $(GO_TESTPKGS) | paste -s -d ',')
TEST_UID:=$(shell id -u)
TEST_GID:=$(shell id -g)
GOCC ?= go
PROTOC ?= protoc

all: core app

go_deps:
	$(GOCC) mod download

core: go_deps proto
	$(GOCC) build -ldflags $(GO_LDFLAGS) -o bin/islb cmd/islb/main.go
	$(GOCC) build -ldflags $(GO_LDFLAGS) -o bin/sfu cmd/sfu/main.go
	$(GOCC) build -ldflags $(GO_LDFLAGS) -o bin/avp cmd/avp/main.go
	$(GOCC) build -ldflags $(GO_LDFLAGS) -o bin/signal cmd/signal/main.go
.PHONY: core

app:
	$(GOCC) build -ldflags $(GO_LDFLAGS) -o bin/app-biz apps/biz/main.go
.PHONY: app

clean:
	rm -rf bin
.PHONY: clean

start-bin:

start-services:
	docker network create ionnet || true
	docker-compose -f docker-compose.yml up -d redis nats
.PHONY: start-services

stop-services:
	docker-compose -f docker-compose.yml stop redis nats
.PHONY: stop-services

run:
	docker-compose up --build
.PHONY: run

test: go_deps start-services
	$(GOCC) test \
		-timeout 120s \
		-coverpkg=${GO_COVERPKGS} -coverprofile=cover.out -covermode=atomic \
		-v -race ${GO_TESTPKGS}
.PHONY: test

%.pb.go: %.proto
	@echo "protobuf generating: $<"
	$(PROTOC) --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.  --proto_path=.:$(GOPATH)/src:$(dir $@) $<

proto-gen-from-docker:
	docker build -t go-protoc ./proto
	docker run -v $(CURDIR):/workspace go-protoc proto
.PHONY: proto-gen-from-docker

proto: proto_core proto_app
.PHONY: proto

proto_core:
	$(PROTOC) --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.  proto/ion/ion.proto
	$(PROTOC) --experimental_allow_proto3_optional --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.  proto/debug/debug.proto
	$(PROTOC) --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.  proto/sfu/sfu.proto
	$(PROTOC) --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.  proto/islb/islb.proto
	$(PROTOC) --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.  proto/rtc/rtc.proto
.PHONY: proto_core

proto_app:
	$(PROTOC) apps/biz/proto/biz.proto --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.
.PHONY: proto_app

.PHONY: all
