.PHONY: test install

run:
	go build -o build/myduct && sudo build/myduct; rm build/myduct

install:
	go build -o ~/bin/myduct
