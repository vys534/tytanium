[![Tytanium.png](https://i.postimg.cc/MKXpNtbF/Tytanium.png)](https://postimg.cc/SJF4z6P6)

# Tytanium

(formerly known as hexFS)

A file host server with server security in mind. Intended for private use.  

## Features

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

## Setup

Make sure you have a Google Cloud Storage service account's JSON key and a Redis instance set up somewhere.

- Put the GCS JSON key in `conf/` as `key.json`.
- Put your config in `conf/` as `config.yml` using `conf/example.yml` as a reference.   
- Optionally, you can replace `favicon.ico` with your own icon! (It must have the same name)
- If you're using this with ShareX, check `example/tytanium.sxcu` for a template sxcu file.

Now you can choose to either run Tytanium with Docker or as a service on your system.

- Choose Docker if **you already have docker setup on your server**.
- Choose systemd if **you want to run it without any virtualization or don't want to install Docker**.

### Option 1: Run with Docker

- Build the image with `docker build -t tytanium .`
- Make sure to bind the port you choose (default is `3030`) to other ports on your system. Here's an example of how you would run it, after building the image.  
  
`docker container run -d -p 127.0.0.1:3030:3030 tytanium`  

### Option 2: Run as a systemd service

- Download the binary to the same directory where `conf/` is located.
- Mark it as executable with `chmod 0744 <binary file>`.
- Copy `example/tytanium.service` to `/lib/systemd/system`.
- Edit the WorkingDirectory and ExecFile to match the locations of the binary file.
- Run `systemctl daemon-reload`.

At this point, you can run the program using `systemctl start tytanium`. You can check on its status by running `systemctl status tytanium`. 

If anything goes wrong, you can check `journalctl -u tytanium` and find out what happened.

### How to Upload

Create a POST request to `/upload` with a file in the field "file". You can also set `omitdomain` to 1 if you don't want the host's original domain appended before the file name in the response. E.g: `a.png` instead of `https://a.com/a.png`