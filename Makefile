.PHONY: test install

run: install
	sudo build/myduct

run-quiet: install
	sudo build/myduct --quiet

run-silent: install
	sudo build/myduct --silent

install:
	go build -o build/myduct
