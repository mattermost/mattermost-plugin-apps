module github.com/mattermost/mattermost-plugin-apps/upstream/upkubeless

go 1.16

require (
	github.com/golang/glog v0.0.0-20210429001901-424d2337a529 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/hashicorp/go-getter v1.5.5
	github.com/kubeless/kubeless v1.0.8
	github.com/mattermost/mattermost-plugin-apps v0.7.0
	github.com/pkg/errors v0.9.1
	golang.org/x/net v0.0.0-20210716203947-853a461950ff // indirect
	k8s.io/api v0.0.0-20180308224125-73d903622b73
	k8s.io/apimachinery v0.0.0-20180228050457-302974c03f7e
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v8.0.0+incompatible
	github.com/mattermost/mattermost-plugin-apps => ../..
)
