deps:
	@$(MAKE) -C ./contracts deps
	@echo "--> Installing Go dependencies..."
	@dep ensure -v

migrate:
	$(MAKE) -C ./contracts migrate

build:
	$(MAKE) -C ./contracts abigen
	go install ./...

start: compile
	@./bin/start

clean:
	$(MAKE) -C ./contracts clean
	rm -rf ~/.plasma

fresh-start: clean start