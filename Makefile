GO_LDFLAGS="-s -w -X 'main.goversion=`go version`' -X 'main.buildstamp=`date -u "+%Y-%m-%d_%I:%M:%S%p"`' -X 'main.githash=`git describe --tags 2>/dev/null`'"
GO_TESTPKGS:=$(shell go list ./... | grep -v cmd | grep -v conf | grep -v node)
GO_COVERPKGS:=$(shell echo $(GO_TESTPKGS) | paste -s -d ',')
TEST_UID:=$(shell id -u)
TEST_GID:=$(shell id -g)

all: core app

go_deps:
	go mod download

core: go_deps
	go build -ldflags $(GO_LDFLAGS) -o bin/islb cmd/islb/main.go
	go build -ldflags $(GO_LDFLAGS) -o bin/sfu cmd/sfu/main.go
	go build -ldflags $(GO_LDFLAGS) -o bin/avp cmd/avp/main.go
	go build -ldflags $(GO_LDFLAGS) -o bin/signal cmd/signal/main.go

app:
	go build -ldflags $(GO_LDFLAGS) -o bin/app-biz apps/biz/main.go

clean:
	rm -rf bin

start-bin:

start-services:
	docker network create ionnet || true
	docker-compose -f docker-compose.yml up -d redis nats

stop-services:
	docker-compose -f docker-compose.yml stop redis nats

run:
	docker-compose up --build

test: go_deps start-services
	go test \
		-timeout 120s \
		-coverpkg=${GO_COVERPKGS} -coverprofile=cover.out -covermode=atomic \
		-v -race ${GO_TESTPKGS}

proto-gen-from-docker:
	docker build -t go-protoc ./proto
	docker run -v $(CURDIR):/workspace go-protoc proto

proto: proto_core proto_app

proto_core: 
	protoc proto/ion/ion.proto --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.
	protoc proto/debug/debug.proto --experimental_allow_proto3_optional --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.
	protoc proto/sfu/sfu.proto --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.
	protoc proto/islb/islb.proto --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.
	protoc proto/rtc/rtc.proto --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.

proto_app:
	protoc apps/biz/proto/biz.proto --go_opt=module=github.com/pion/ion --go_out=. --go-grpc_opt=module=github.com/pion/ion --go-grpc_out=.
