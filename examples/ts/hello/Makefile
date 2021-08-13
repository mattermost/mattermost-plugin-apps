NPM ?= npm

.PHONY: all
## all: builds and runs the service
all: lint dist

.PHONY: build
## build: build the app
build: node_modules
	$(NPM) run build

.PHONY: run
## run: runs the service
run: build
	LOCAL=true $(NPM) run start

.PHONY: lint
## lint: Run eslint and check types
lint: node_modules
	$(NPM) run lint
	$(NPM) run check-types

SCRATCH = ./dist/scratch

.PHONY: dist-aws
## dist-aws: creates the bundle file for an AWS Lambda deployment with appsctl.
dist-aws: build
	mkdir -p $(SCRATCH)
	cp ./dist/*.js $(SCRATCH)
	cp -r ./node_modules $(SCRATCH)
	cp -r ./static $(SCRATCH)
	cd $(SCRATCH) && zip -qrm hello.zip *
	cp ./package.json $(SCRATCH)
	cp ./static/manifest.json $(SCRATCH)
	cp -r ./static $(SCRATCH)
	cd $(SCRATCH) && zip -rm ../aws-bundle.zip *
	rm -rf $(SCRATCH)

.PHONY: dist-kubeless
## dist-kubeless: creates the bundle file for a Kubeless deployment with appsctl.
dist-kubeless: build
	mkdir -p $(SCRATCH)
	cp ./dist/*.js $(SCRATCH)
	cp -r ./static $(SCRATCH)
	cd $(SCRATCH) && zip -qrm hello.zip *
	cp ./package.json $(SCRATCH)
	cp ./static/manifest.json $(SCRATCH)
	cd $(SCRATCH) && zip -rm ../kubeless-bundle.zip *
	rm -rf $(SCRATCH)

.PHONY: dist
## dist: creates bundle files for all supported distributions.
dist: dist-aws dist-kubeless

## build: build the app when changed
.PHONY: watch
watch: node_modules
	$(NPM) run build:watch

.PHONY: clean
## clean: deletes all
clean:
	$(NPM) run clean

## node_modules: ensures NPM dependencies are installed without having to run this all the time
node_modules: $(wildcard package.json)
	$(NPM) install
	touch $@
	
.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'