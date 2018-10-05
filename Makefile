deps:
	@$(MAKE) -C ./contracts deps
	@$(MAKE) -C ./rpc-test-client deps
	@echo "--> Installing Go dependencies..."
	@dep ensure -v

migrate:
	$(MAKE) -C ./contracts migrate

build:
	go build -o ./build/plasma ./cmd/plasma/main.go
	
install:
	go install ./cmd/plasma

abigen:
	$(MAKE) -C ./contracts abigen

protogen:
	protoc -I rpc/proto rpc/proto/root.proto --go_out=plugins=grpc:rpc/pb

build-all: abigen build

start: deps build
	@./bin/start

setup: deps build
	@./bin/setup

clean:
	$(MAKE) -C ./contracts clean
	rm -rf ~/.plasma
	rm -r ./test/storage/ganache/*
	rm -rf ./test/storage/root

test:
	@go test ./...

fresh-start: clean start

.PHONY: build deps test
