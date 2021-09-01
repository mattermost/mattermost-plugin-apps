## Self-managed Apps Hosting demo

#### Setup:
- dev Mattermost Server + Apps plugin
- AWS: personal account
- Faasd installed on a Multipass Ubuntu VM
- Kubeless installed on Docker desktop Kubernetes

#### AWS Lambda
- Clean out functions and S3 bucket
- Init AWS
```sh
go run ./cmd/appsctl aws init --create
```
- Build and deploy 2 apps to AWS
```sh
cd ./examples/go/hello-serverless && make dist-aws
go run ./cmd/appsctl aws provision -v ./examples/go/hello-serverless/dist/bundle-aws.zip
cd ./examples/ts/hello && make dist-aws
go run ./cmd/appsctl aws provision -v ./examples/ts/hello/dist/aws-bundle.zip
```
- Install the apps
```
/apps install s3 hello-serverless demo
/apps install s3 hello-typescript 0.9.0
```
- Use the apps
```
/hello-serverless send 
/hello-typescript send 
```

#### OpenFaaS/faasd
- Clean out functions
```
faas-cli list
```
- Build and deploy an app to faasd
```sh
cd ./examples/go/hello-serverless && make dist-openfaas
go run ./cmd/appsctl openfaas provision -v ./examples/go/hello-serverless/dist/bundle-openfaas.zip
```
- Install the app
```
/apps install url http://192.168.64.3:8080/function/hello-serverless-demo-hello/manifest.json
```

#### Kubeless
- Clean out functions
```
kubeless function list --namespace mattermost-kubeless-apps
```
- Build and deploy an app to faasd
```sh
cd ./examples/ts/hello && make dist-kubeless
go run ./cmd/appsctl kubeless provision -v ./examples/ts/hello/dist/kubeless-bundle.zip
```
- Install the app
```
/apps install url https://raw.githubusercontent.com/levb/mattermost-plugin-apps/faasd/examples/ts/hello/static/manifest.json
```

#### Inside the App