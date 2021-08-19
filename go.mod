module github.com/mattermost/mattermost-plugin-apps

go 1.16

require (
	github.com/aws/aws-lambda-go v1.25.0
	github.com/aws/aws-sdk-go v1.38.67
	github.com/awslabs/aws-lambda-go-api-proxy v0.10.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/golang/mock v1.5.0
	github.com/google/go-cmp v0.5.6
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-getter v1.5.5
	github.com/kubeless/kubeless v1.0.8
	github.com/mattermost/mattermost-plugin-api v0.0.18
	github.com/mattermost/mattermost-server/v5 v5.3.2-0.20210714130822-54b0ef574b5d
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/oauth2 v0.0.0-20210402161424-2e8d93401602
	google.golang.org/api v0.44.0
	k8s.io/apimachinery v0.0.0-20180228050457-302974c03f7e
	k8s.io/client-go v7.0.0+incompatible
)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v8.0.0+incompatible
