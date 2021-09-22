.PHONY: all
## all: builds and runs the app locally
all: dist run

.PHONY: run
## run: runs the app locally
run: 
	MODE=http go run .


.PHONY: dist-aws
## dist-aws: creates the bundle file for AWS Lambda deployments
dist-aws: 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/hello-serverless . 
	rm -rf dist/aws && mkdir -p dist/aws
	cp app/data/manifest.json dist/aws
	cp -r app/static/ dist/aws/static
	cd dist/aws ; \
		zip -q hello-serverless ../hello-serverless ; \
		zip -rmq ../bundle-aws.zip hello-serverless.zip manifest.json static
	rm -r dist/aws

.PHONY: dist-openfaas
## dist-openfaas: creates the bundle file for OpenFaaS deployments
dist-openfaas:
	mkdir -p dist
	rm -f dist/bundle-openfaas.zip
	rm -rf dist/openfaas && mkdir -p dist/openfaas
	cp app/data/manifest.json dist/openfaas
	cp go.mod dist/openfaas
	cp go.sum dist/openfaas
	cp openfaas-golang-middleware.yml dist/openfaas/manifest.yml
	cp -r app dist/openfaas
	cp -r template dist/openfaas
	cd dist/openfaas ; \
		zip -rmq ../bundle-openfaas.zip *

.PHONY: dist
## dist: creates all bundles
dist: dist-aws dist-openfaas

.PHONY: clean
## clean: deletes all
clean:
	rm -rf dist/

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
