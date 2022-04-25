**TODO** update


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
cd ./examples/go/serverless && make dist-aws
go run ./cmd/appsctl aws deploy -v ./examples/go/serverless/dist/bundle-aws.zip
```
- Install the apps
```
/apps install listed example-serverless 
```
- Use the apps
```
/example-serverless send 
```

#### OpenFaaS/faasd
- Clean out functions
```
faas-cli list
```
- Build and deploy an app to faasd
```sh
cd ./examples/go/serverless && make dist-openfaas
go run ./cmd/appsctl openfaas deploy -v ./examples/go/serverless/dist/bundle-openfaas.zip
```
- Install the app
```
/apps install listed example-serverless
```
