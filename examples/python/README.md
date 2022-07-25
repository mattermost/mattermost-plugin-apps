### Pre-requisite
- Have python installed, preferably `>=3.0`
- Install the requirements mentioned in the `requirement.txt` with `pip3 install -r requirements.txt`
- Configure the following environment variables to run the app on custom port/url
  ```
  export PORT=8080
  export ROOT_URL=http://localhost:8080
  export HOST=0.0.0.0
  ```
- To run with `ngrok`
  1. Start the ngrok server on 8080 port, `ngrok http 8080`
  2. Export the ngrok url. (replace the ngrok url)
     ```
     export ROOT_URL=https://4492-103-161-231-165.in.ngrok.io
     ```

#### RUN the app 
* `python3 hello-world.py`
