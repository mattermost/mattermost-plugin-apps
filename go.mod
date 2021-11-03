module github.com/mattermost/mattermost-plugin-apps

go 1.16

require (
	github.com/Azure/go-autorest/autorest v0.11.21 // indirect
	github.com/aws/aws-sdk-go v1.41.17
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.6
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-getter v1.5.9
	github.com/hashicorp/go-multierror v1.1.1
	github.com/kubeless/kubeless v1.0.8
	github.com/mattermost/mattermost-plugin-api v0.0.22-0.20211103113715-7277517c2940
	github.com/mattermost/mattermost-server/v6 v6.0.0-20211103113238-be923223d26f
	github.com/nicksnyder/go-i18n/v2 v2.1.2
	github.com/openfaas/faas-cli v0.0.0-20211012083206-08e3e965c831
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/oauth2 v0.0.0-20211028175245-ba495a64dcb5
	google.golang.org/api v0.60.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v7.0.0+incompatible
)

// https://github.com/kubernetes/client-go/issues/874
replace k8s.io/client-go => k8s.io/client-go v0.22.3
