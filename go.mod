module github.com/mattermost/mattermost-plugin-apps

go 1.16

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/golang/mock v1.5.0
	github.com/google/go-cmp v0.5.5
	github.com/gorilla/mux v1.8.0
	github.com/mattermost/mattermost-plugin-api v0.0.15
	github.com/mattermost/mattermost-plugin-apps/upstream/upaws v0.0.0-00010101000000-000000000000
	github.com/mattermost/mattermost-server/v5 v5.3.2-0.20210503144558-5c16de58a020
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602
	google.golang.org/api v0.44.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v8.0.0+incompatible
	github.com/mattermost/mattermost-plugin-apps => ./
	github.com/mattermost/mattermost-plugin-apps/upstream/upaws => ./upstream/upaws
	github.com/mattermost/mattermost-plugin-apps/upstream/upkubeless => ./upstream/upkubeless
)
