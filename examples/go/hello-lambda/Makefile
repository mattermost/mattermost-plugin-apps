.PHONY: all
## all: builds and runs the app locally
all: dist run

.PHONY: build-linux
## build-linux: build the executable for linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/hello-lambda ./aws 

.PHONY: run
## run: runs the app locally
run: build
	go run ./http

.PHONY: dist
## dist: creates the bundle file
dist: build-linux
	mkdir -p dist
	cp manifest.json dist/
	cp -r static dist/
	cd dist ; zip -qrm hello-lambda hello-lambda ; zip -rm bundle.zip hello-lambda.zip manifest.json static

.PHONY: clean
## clean: deletes all
clean:
	rm -rf dist/

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
