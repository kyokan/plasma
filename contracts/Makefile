NETWORK=development

.PHONY: test

deps:
	@echo "--> Installing Truffle dependencies..."
	@npm install

test:
	truffle test --network $(NETWORK)

migrate:
	@echo "--> Migrating smart contracts to network $(NETWORK)"
	@truffle migrate --network $(NETWORK) --reset

abigen:
	./generate.js --network $(NETWORK)

testnet:
	@mkdir -p ~/.plasma/ganache
	@ganache-cli -m "candy maple cake sugar pudding cream honey rich smooth crumble sweet treat" -d -b 1 -v -p 9545 --db ~/.plasma/ganache

clean:
	rm -rf ./build
