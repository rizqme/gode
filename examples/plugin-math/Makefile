PLUGIN_NAME = math.so

build:
	go build -buildmode=plugin -o $(PLUGIN_NAME) main.go

clean:
	rm -f $(PLUGIN_NAME)

test: build
	gode run test.js

.PHONY: build clean test