# Include custom targets and environment variables here

.PHONY: dev_server
dev_server:
	cd dev && docker-compose up mattermost db
