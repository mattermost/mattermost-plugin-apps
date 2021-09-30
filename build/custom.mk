# Include custome targets and environment variables here
ifndef MM_RUDDER_WRITE_KEY
    MM_RUDDER_WRITE_KEY = 1d5bMvdrfWClLxgK1FvV3s4U1tg
endif

GO_BUILD_FLAGS += -ldflags '-X "github.com/mattermost/mattermost-plugin-api/experimental/telemetry.rudderWriteKey=$(MM_RUDDER_WRITE_KEY)"'

.PHONY: dev_server
dev_server:
	cd dev && docker-compose up mattermost db

.PHONY: hello_world
hello_world:
	cd examples/go/hello-world && go run hello.go
