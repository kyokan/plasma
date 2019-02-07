deps:
	@echo "--> Installing dependencies for Plasma MVP Rootchain..."
	git submodule init
	git submodule update --remote --recursive
	npm install --prefix plasma-mvp-rootchain plasma-mvp-rootchain
	@echo "--> Installing Go dependencies..."
	@dep ensure -v

build-plasmad:
	go build -o ./target/plasmad ./cmd/plasmad/main.go

build-harness:
	go build -o ./target/plasma-harness ./cmd/harness/main.go

build-plasmacli:
	go build -o ./target/plasmacli ./cmd/plasmacli/main.go

build-plasmad-debug:
	go build -gcflags "all=-N -l" -o ./target/plasmad ./cmd/plasmad/main.go

build-cross:
	docker build ./build --no-cache -t plasma-cross-compilation:latest
	mkdir -p ./target
	docker run --name cp-tmp plasma-cross-compilation:latest /bin/true
	docker cp cp-tmp:/go/src/github.com/kyokan/plasma/target/plasma ./target/plasma-linux-amd64
	docker rm cp-tmp

install:
	go install ./cmd/plasmad

abigen: deps
	cd plasma-mvp-rootchain && \
	truffle compile && \
	cat ./build/contracts/PlasmaMVP.json | jq ".abi" > ../eth/contracts/PlasmaMVP.abi && \
	abigen --abi ../eth/contracts/PlasmaMVP.abi --pkg contracts --type Plasma --out ../eth/contracts/plasma_mvp.go && \
	cp ../eth/contracts/PlasmaMVP.abi ../integration_tests/test/abi/PlasmaMVP.abi.json && \
	rm -rf abi && \
	rm -rf gen

protogen:
	protoc -I rpc/proto rpc/proto/root.proto --go_out=plugins=grpc:rpc/pb

clean:
	rm -rf ./plasma-mvp-rootchain/node_modules
	rm -rf ./plasma-mvp-rootchain/build
	rm -rf ./integration_tests/node_modules
	rm -rf ./eth/contracts/plasma_mvp.go
	rm -rf ./eth/contracts/PlasmaMVP.abi
	rm -rf ./target
	rm -rf ~/.plasma
	rm -rf .vendor-new

test-integration:
	npm --prefix ./integration_tests test

package:
	fpm -f -p target -s dir -t deb -n plasma -a amd64 -m "Kyokan, LLC <mslipper@kyokan.io>" \
    	--url "https://plasma.kyokan.io" \
    	--description "A Golang implementation of the Minimum Viable Plasma spec." \
    	--license "AGPL-V3" \
        --vendor "Kyokan, LLC" \
		--config-files /etc/plasma/config.yaml -v $(VERSION) \
		target/plasma-linux-amd64=/usr/bin/plasma \
		example/config.yaml=/etc/plasma/

deploy:
	@echo "--> Uploading version $(VERSION) to Bintray..."
	@curl -s -S -T ./target/plasma_$(VERSION)_amd64.deb -u$(USERNAME):$(API_KEY) \
		-H "X-GPG-PASSPHRASE: $(GPG_PASSPHRASE)" \
		-H "X-Bintray-Debian-Distribution: any" \
        -H "X-Bintray-Debian-Component: main" \
        -H "X-Bintray-Debian-Architecture: amd64" \
		https://api.bintray.com/content/kyokan/oss-deb/plasma/$(VERSION)/plasma_$(VERSION)_amd64.deb
	@sleep 3
	@echo "--> Publishing package..."
	@curl -s -S -X POST -u$(USERNAME):$(API_KEY) \
			https://api.bintray.com/content/kyokan/oss-deb/plasma/$(VERSION)/publish
	@sleep 10
	@echo "--> Forcing metadata calculation..."
	@curl -s -S -X POST -H "X-GPG-PASSPHRASE: $(GPG_PASSPHRASE)" -u$(USERNAME):$(API_KEY) https://api.bintray.com/calc_metadata/kyokan/oss-deb/

fresh-start: clean start

.PHONY: build deps test package
