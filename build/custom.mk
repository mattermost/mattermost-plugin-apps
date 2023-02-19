# Include custome targets and environment variables here
default: all

ifndef MM_RUDDER_WRITE_KEY
    MM_RUDDER_WRITE_KEY = 1d5bMvdrfWClLxgK1FvV3s4U1tg
endif

LDFLAGS += -X "github.com/mattermost/mattermost-plugin-api/experimental/telemetry.rudderWriteKey=$(MM_RUDDER_WRITE_KEY)"

BUILD_DATE = $(shell date -u)
BUILD_HASH = $(shell git rev-parse HEAD)
BUILD_HASH_SHORT = $(shell git rev-parse --short HEAD)
LDFLAGS += -X "github.com/mattermost/mattermost-plugin-apps/server/config.BuildDate=$(BUILD_DATE)"
LDFLAGS += -X "github.com/mattermost/mattermost-plugin-apps/server/config.BuildHash=$(BUILD_HASH)"
LDFLAGS += -X "github.com/mattermost/mattermost-plugin-apps/server/config.BuildHashShort=$(BUILD_HASH_SHORT)"
GO_BUILD_FLAGS += -ldflags '$(LDFLAGS)'
GO_TEST_FLAGS += -ldflags '$(LDFLAGS)'

MM_SERVER_PATH ?= ${PWD}/../mattermost-server
export MM_SERVER_PATH

## Generates mock golang interfaces for testing
.PHONY: mock
mock:
ifneq ($(HAS_SERVER),)
	go install github.com/golang/mock/mockgen@v1.6.0
	mockgen -destination server/mocks/mock_proxy/mock_expand_getter.go github.com/mattermost/mattermost-plugin-apps/server/proxy ExpandGetter
	mockgen -destination server/mocks/mock_upstream/mock_upstream.go github.com/mattermost/mattermost-plugin-apps/upstream Upstream
	mockgen -destination server/mocks/mock_store/mock_appstore.go github.com/mattermost/mattermost-plugin-apps/server/store AppStore
	mockgen -destination server/mocks/mock_store/mock_session.go github.com/mattermost/mattermost-plugin-apps/server/store SessionStore
	mockgen -destination server/mocks/mock_store/mock_app.go github.com/mattermost/mattermost-plugin-apps/server/store AppStore
endif

## Generates mock golang interfaces for testing
.PHONY: clean_mock
clean_mock:
ifneq ($(HAS_SERVER),)
	rm -rf ./server/mocks
endif

## Run Go REST API system tests
.PHONY: test-rest-api
test-rest-api: dist
	@echo Running REST API tests
ifneq ($(RUN),)
	PLUGIN_BUNDLE=$(shell pwd)/dist/$(BUNDLE_NAME) $(GO) test -v $(GO_TEST_FLAGS) ./test/restapitest --run $(RUN)
else
	PLUGIN_BUNDLE=$(shell pwd)/dist/$(BUNDLE_NAME) $(GO) test -v $(GO_TEST_FLAGS) ./test/restapitest
endif


## Extract new translation messages
.PHONY: i18n-extract-server
i18n-extract-server:
	@goi18n extract -format json -outdir assets/i18n/ server/ utils/ apps/ cmd/ upstream/
	@for x in assets/i18n/active.*.json; do echo $$x | sed 's/active/translate/' | sed 's/^/touch /' | bash; done
	@goi18n merge -format json -outdir assets/i18n/ assets/i18n/active.*.json
	@echo "Please update your assets/i18n/translate.*.json files and execute \"make i18n-merge-server\""
	@echo "If you don't want to change any locale file, simply remove the assets/i18n/translate.??.json file before calling \"make i18n-merge-server\""
	@echo "If you want to add a new language (for example french), simply run \"touch assets/i18n/active.fr.json\" and then run the \"make i18n-extract-server\" again"

## Merge translated messages
.PHONY: i18n-merge-server
i18n-merge-server:
	@goi18n merge -format json -outdir assets/i18n/ assets/i18n/active.*.json assets/i18n/translate.*.json
	@rm -f assets/i18n/translate.*.json
	@echo "Translations merged, please verify your "git diff" before you submit the changes"

## Run a simple Mattermost Server
.PHONY: dev_server
dev_server:
	cd dev && docker-compose up mattermost db

## Run the hello-world app
.PHONY: run-example-hello-4000
run-example-hello-4000:
	cd test/hello-world && go run .

## Run the test app
.PHONY: run-test-app-8081
run-test-app-8081:
	cd test/app && go run .
