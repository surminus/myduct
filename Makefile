.PHONY: test install

run: install
	sudo build/myduct

install:
	go build -o build/myduct
