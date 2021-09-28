# Include custom targets and environment variables here

.PHONY: dev_server
dev_server:
	cd dev && docker-compose up mattermost db

.PHONY: hello_world
hello_world:
	cd examples/go/hello-world && go run hello.go
