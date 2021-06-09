.PHONY: all
## all: builds and runs the app locally
all: run

.PHONY: build-linux
## build-linux: build the executable for linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/hello-lambda

.PHONY: build
## build: build the executable for the local OS
build:
	go build -o dist/hello-lambda

.PHONY: run
## run: runs the app locally
run: build
	LOCAL=true ./dist/hello-lambda

.PHONY: dist
## dist: creates the bundle file
dist: build-linux
	cp -r static dist; cp manifest.json dist/; cd dist/; zip -qr go-function hello-lambda; zip -r bundle.zip go-function.zip manifest.json static/

.PHONY: clean
## clean: deletes all
clean:
	rm -rf dist/

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
