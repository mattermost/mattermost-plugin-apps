module github.com/mattermost/mattermost-plugin-apps/upstream/upaws

go 1.16

require (
	github.com/aws/aws-sdk-go v1.40.2
	github.com/mattermost/mattermost-plugin-apps v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)

replace (
	github.com/mattermost/mattermost-plugin-apps => ../..
	github.com/mattermost/mattermost-plugin-apps/upstream/upaws => ./
)
