.PHONY: all
## all: builds and runs the app locally
all: dist run

.PHONY: run
## run: runs the app locally
run: 
	DEPLOY_TYPE=http go run .

.PHONY: dist-aws
## dist-aws: creates the bundle file for AWS Lambda deployments
dist-aws: 
	rm -rf dist/aws && mkdir -p dist/aws
	GOOS=linux GOARCH=amd64 go build -o dist/aws/hello-serverless . 
	cp ./hello/static/manifest.json dist/aws
	cp -r ./hello/static dist/aws
	cd dist/aws ; \
		zip -m hello-serverless.zip hello-serverless ; \
		zip -rm ../bundle-aws.zip hello-serverless.zip manifest.json static
	rm -r dist/aws

.PHONY: dist-openfaas
## dist-openfaas: creates the bundle file for OpenFaaS deployments
dist-openfaas:
	rm -rf dist/openfaas && mkdir -p dist/openfaas
	cp -r function dist/openfaas
	cp -r hello dist/openfaas
	cp hello/static/manifest.json dist/openfaas
	cp openfaas.yml dist/openfaas/manifest.yml
	cd dist/openfaas ; \
		faas-cli template store pull golang-middleware ; \
		zip -rq ../bundle-openfaas.zip *
	rm -r dist/openfaas

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
