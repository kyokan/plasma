NETWORK=development

.PHONY: test

test:
	truffle test --network $(NETWORK)

migrate:
	truffle migrate --network $(NETWORK)

abigen:
	./generate.js --network $(NETWORK)

clean:
	rm -rf ./build