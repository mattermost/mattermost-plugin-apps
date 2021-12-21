# Mattermost Apps Framework development environment

When you're developing your own App, you need an actual Mattermost server to be running. The Apps Framework development environment helps accomplish this by setting up a minimalistic environment with just two containers. One is for the database Mattermost communicates with, and the other container runs the actual Mattermost server. The other containers present in the Mattermost development environment are unnecessary for the purposes of building Apps. So the advantage here is that there is just one dependency to start developing Apps.


Start the example App by running `docker-compose up`. This will spin up 3 docker containers:

- Mattermost Server
- Postgres
- Example App written in Node.js

Visit http://localhost:8066 to connect to the Mattermost instance. Once your account is set up, run the following slash command to install the example App:

`/apps install http http://node_app:4000/manifest.json`

Your App can be written in another language than JavaScript, and can be in a different directory. You'll just need to edit [docker-compose.override.yml](docker-compose.override.yml), and change the `volumes` to match the relative path to your app, and change the `command` to match your App's start command. Note how the environment variables are used in `src/app.ts`

Run the following commands in two different terminals to have your app run in its own terminal:

```sh
docker-compose up mattermost db

docker-compose up node_app
```

If you want to run your App outside of Docker, you will need to provide a way for the containers to access your server, such as using an [ngrok](https://ngrok.io) tunnel.

To upgrade the Mattermost server or Apps framework plugin, you can edit [docker-compose.yml](docker-compose.yml) to configure your target versions. If you need to make changes to the Apps plugin, you can redeploy the plugin after making changes by setting these environment variables:

```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8066
export MM_ADMIN_USERNAME=(your Mattermost admin username)
export MM_ADMIN_PASSWORD=(your Mattermost admin password)
```

Then run `make deploy` to compile the Apps plugin from source and automatically redeploy the plugin to your server.

## Mattermost development environment

Alternatively, you can setup a fully-fledged Mattermost development environment by following the steps [here](https://developers.mattermost.com/contribute/server/developer-setup/).

Some differences between the environments:

* The Apps Framework development environment has its own Mattermost server, and it's fully configured to start developing Apps. The config values are set correctly so no modifications need to be done there.

* The Apps Framework development environment also has a starter App built-in as a third container, but this can be ignored if the developer wishes to run their App outside of the dev environment, while still using it by communicating with it from outside of the containers.

* The Apps Framework e2e tests can't be run with the Apps Framework development environment.
