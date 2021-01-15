[![Tytanium.png](https://i.postimg.cc/MKXpNtbF/Tytanium.png)](https://postimg.cc/SJF4z6P6)

# Tytanium

(formerly known as hexFS)

A file host server with server security in mind. Intended for private use.  

### Features

- Excellent compatibility with image capture suites like ShareX/MagicCap/etc.
- Limit requests to a certain amount for an interval that you can choose
- Limit upload/download for an interval that you choose
- Store files in GCS with encryption at-rest
- Compatible with Docker
- Rate limiting for individual paths
- Runs with [fasthttp](https://github.com/vayala/fasthttp) for maximum performance and built-in anti-DoS features
- Whitelist/blacklist file types, and check them based on their headers, not the extension
- Sanitize files to prevent against phishing attacks
- Public/private mode (private by default)

### How to setup

Make sure you have a Google Cloud Storage service account's JSON key and a Redis instance set up somewhere.

- Put the GCS JSON key in `conf/` as `key.json`.
- Put your config in `conf/` as `config.yml` using `conf/example.yml` as a reference.   
- Optionally, you can replace `favicon.ico` with your own icon! (It must have the same name)
- If you do not want to set a rate limit, set the values on them to 0.
- Optional: if you're using this with ShareX, check `example/example.sxcu` for a template sxcu file.

### Run with Docker

- Ensure first that you have Docker installed on your system.
- Build the image with `docker build -t tytanium .`
- Make sure to bind the port you choose (default is `3030`) to other ports on your system. Here's an example of how you would run it, after building the image.  
  
`docker container run -d -p 127.0.0.1:3030:3030 tytanium`  

### Run with systemd/service file/etc.

Just build and run the executable.