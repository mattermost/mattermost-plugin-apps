.PHONY: all
## all: builds and runs the app locally
all: run

.PHONY: run
## run: runs the app locally
run: 
	go run ./http

.PHONY: dist
## dist: creates the bundle file
dist: build-linux
	mkdir -p dist
	zip -qr dist/go-function app.go manifest.json bindings.json pong.json; 
	cp -r static dist; 
	cp manifest.json dist/; 
	cd dist/; 
	zip -qr go-function hello-lambda; 
	zip -r bundle.zip go-function.zip manifest.json

.PHONY: clean
## clean: deletes all
clean:
	rm -rf dist/

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
