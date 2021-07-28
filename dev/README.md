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

To upgrade the Mattermost server or Apps framework plugin, you can edit [docker-compose.yml](docker-compose.yml) to configure your target versions. If you need to make changes to the Apps plugin, you can redeploy the plugin after making changes by setting these environment variables:

```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8066
export MM_ADMIN_USERNAME=(your Mattermost admin username)
export MM_ADMIN_PASSWORD=(your Mattermost admin password)
```

Then run `make deploy` to compile the Apps plugin from source and automatically redeploy the plugin to your server.
