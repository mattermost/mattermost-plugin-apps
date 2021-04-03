module github.com/mattermost/mattermost-plugin-apps

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/golang/mock v1.5.0
	github.com/google/go-cmp v0.5.5
	github.com/gorilla/mux v1.8.0
	github.com/mattermost/mattermost-plugin-api v0.0.15-0.20210303034931-22355254f0ea
	github.com/mattermost/mattermost-server/v5 v5.3.2-0.20210120031517-5a7759f4d63b
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	golang.org/x/oauth2 v0.0.0-20210313182246-cd4f82c27b84
	google.golang.org/api v0.42.0
)

replace github.com/mattermost/mattermost-server/v5 v5.3.2-0.20210120031517-5a7759f4d63b => /Users/catalintomai/go/src/github.com/mattermost/mattermost-server

replace github.com/mattermost/mattermost-plugin-api v0.0.15-0.20210303034931-22355254f0ea => /Users/catalintomai/go/src/github.com/mattermost/mattermost-plugin-api
