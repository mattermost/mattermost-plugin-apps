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

.PHONY: dist-open-faas
## dist-open-faas: creates the bundle file for OpenFaaS deployments
dist-open-faas:
	mkdir -p dist
	rm -f dist/bundle-open-faas.zip
	rm -rf dist/open-faas && mkdir -p dist/open-faas
	cp app/data/manifest.json dist/open-faas
	cp go.mod dist/open-faas
	cp go.sum dist/open-faas
	cp open-faas-golang-middleware.yml dist/open-faas/manifest.yml
	cp -r app dist/open-faas
	cp -r template dist/open-faas
	cd dist/open-faas ; \
		zip -rmq ../bundle-open-faas.zip *
	rm -r dist/open-faas

.PHONY: dist
## dist: creates all bundles
dist: dist-aws dist-open-faas

.PHONY: clean
## clean: deletes all
clean:
	rm -rf dist/

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
