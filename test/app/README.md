# Test app for Mattermost

This app is used for testing the Mattermost Apps framework.

To run,

```sh
[PORT=] [ROOT_URL=] [INCLUDE_INVALID=true] go run .
```

- `PORT` specifies a local port to listen on, default is 8081.
- `ROOT_URL` is the `root_url` to use in the manifest, if different from
  `http://localhost:$PORT`.
- `INCLUDE_INVALID=true` makes various invalid bindings and forms appear, will
  generate a lot of logs.
