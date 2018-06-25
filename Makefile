compile:
	$(MAKE) -C ./contracts abigen
	go install ./...

start: compile
	@echo "Starting..."
	@./bin/shoreman

clean:
	rm -rf ~/.plasma

fresh-start: clean start