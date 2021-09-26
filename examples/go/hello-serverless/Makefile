.PHONY: all
## all: builds and runs the app locally
all: dist run

.PHONY: run
## run: runs the app locally
run: 
	cd http ; \
		go run . --manifest=../manifest.json --static=../static

.PHONY: dist-http
## dist-http: creates the bundle file for HTTP deployment from a bundle (manifest and static deployed to Mattermost)
dist-http:
	rm -rf dist/http && mkdir -p dist/http
	cp manifest.json dist/http
	cp -r static dist/http
	cd dist/http ; \
		zip -rm ../bundle-http.zip manifest.json static
	rm -r dist/http

.PHONY: dist-aws
## dist-aws: creates the bundle file for AWS Lambda deployments
dist-aws: 
	rm -rf dist/aws && mkdir -p dist/aws
	cd aws ; \
	 	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../dist/aws/hello-serverless . 
	cp manifest.json dist/aws
	cp -r static dist/aws
	cd dist/aws ; \
		zip -m hello-serverless.zip hello-serverless ; \
		zip -rm ../bundle-aws.zip hello-serverless.zip manifest.json static
	rm -r dist/aws

.PHONY: dist-openfaas
## dist-openfaas: creates the bundle file for OpenFaaS deployments
dist-openfaas:
	rm -rf dist/openfaas && mkdir -p dist/openfaas
	cp manifest.json dist/openfaas
	cp -r static dist/openfaas
	cp -r function dist/openfaas
	cp openfaas/manifest.yml dist/openfaas
	cp -r openfaas/template dist/openfaas
	cd dist/openfaas ; \
		zip -rq ../bundle-openfaas.zip *
	rm -r dist/openfaas

.PHONY: dist
## dist: creates all bundles
dist: dist-http dist-aws dist-openfaas

.PHONY: clean
## clean: deletes all
clean:
	rm -rf dist/

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
