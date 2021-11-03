module github.com/mattermost/mattermost-plugin-apps

go 1.16

require (
	github.com/BurntSushi/toml v0.4.1 // indirect
	github.com/aws/aws-sdk-go v1.40.57
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.6
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-getter v1.5.5
	github.com/hashicorp/go-multierror v1.1.1
	github.com/kubeless/kubeless v1.0.8
	github.com/mattermost/mattermost-plugin-api v0.0.22-0.20211103113715-7277517c2940
	github.com/mattermost/mattermost-server/v6 v6.0.0-20211103113238-be923223d26f
	github.com/nicksnyder/go-i18n/v2 v2.1.2
	github.com/openfaas/faas-cli v0.0.0-20210705110531-a230119be00f
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	google.golang.org/api v0.51.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/apimachinery v0.0.0-20180228050457-302974c03f7e
	k8s.io/client-go v7.0.0+incompatible
)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v8.0.0+incompatible
