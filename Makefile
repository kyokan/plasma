compile:
	$(MAKE) -C ./contracts abigen
	go install ./...

start: compile
	@./bin/start

clean:
	$(MAKE) -C ./contracts clean
	rm -rf ~/.plasma

fresh-start: clean start