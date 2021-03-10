# App Lifecycle
## Development
## Submit to Marketplace
## Provision
## Publish
## Install
Apps are installed with `apps install 
- Manifest -> Installed App
  - Consent to permissions, locations, OAuth app type
  - Create Bot+Access Token, OAuth App
  - HTTP: collect app’s JWT secret
- Invoke “OnInstall” callback on the App
  - Admin access token
- Also Uninstall/Enable/Disable per App

## Uninstall
## Upgrade/downgrade consideration
## appsctl 