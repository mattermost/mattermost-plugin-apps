.PHONY: dist
## dist: creates the bundle file
dist: 
	zip -r bundle.zip manifest.json handler.go go.mod
	mkdir -p dist
	mv bundle.zip ./dist

.PHONY: clean
## clean: deletes all
clean:
	rm -rf dist/

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
