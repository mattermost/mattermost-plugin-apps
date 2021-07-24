.PHONY: dist
## dist: creates the bundle file
dist: 
	mkdir -p dist
	zip -r dist/bundle.zip manifest.json handler.go go.mod

.PHONY: clean
## clean: deletes all
clean:
	rm -rf dist/

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
