compile:
	go install ./...

start: compile
	${GOPATH}/bin/plasma --node-url ${GETH_WS_URL} start

clean:
	rm -rf ~/.plasma

fresh-start: clean start