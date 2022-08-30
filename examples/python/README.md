### Pre-requisite
1. Have [python installed](https://www.python.org/downloads/), preferably `>=3.0`
2. Change working directory to `~/mattermost-plugin-apps/examples/python`
3. Install the requirements mentioned in the `requirement.txt` with `pip3 install -r requirements.txt`
3. Configure the following environment variables to run the app on custom port/url
  ```
  export PORT=8080
  export ROOT_URL=http://localhost:8080
  export HOST=0.0.0.0
  ```
- To run with [ngrok](https://ngrok.com/download)
  1. Start the ngrok server on 8080 port, `ngrok http 8080`
  2. Export the ngrok url. (replace the ngrok url)
     ```
     export ROOT_URL=https://4492-103-161-231-165.in.ngrok.io
     ```

#### RUN the app 
1. Change directory to `~/mattermost-plugin-apps/examples/python/src`.
2. Run the app via `python3 hello-world.py`
