## Self-managed Apps Hosting demo

#### Setup:
- dev Mattermost Server + Apps plugin
- AWS: personal account
- Faasd installed on a Multipass Ubuntu VM

#### AWS Lambda
- Clean out functions and S3 bucket
- Init AWS
```sh
go run ./cmd/appsctl aws init --create
```
- Build and deploy 2 apps to AWS
```sh
cd ./examples/go/hello-serverless && make dist-aws
go run ./cmd/appsctl aws deploy -v ./examples/go/hello-serverless/dist/bundle-aws.zip
```
- Install the apps
```
/apps install listed hello-serverless 
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
go run ./cmd/appsctl openfaas deploy -v ./examples/go/hello-serverless/dist/bundle-openfaas.zip
```
- Install the app
```
/apps install listed hello-serverless
```
