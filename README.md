# LoliHost

Created for use on https://l-o.li (now runs on a private fork from this repo). A file host server with security in mind, which stores everything locally. Intended for private/small group use, and for things like screenshots (though you could use it for something else as well).

## Features
    
- Excellent compatibility with image capture suites like ShareX/MagicCap/etc.
- Customizable rate limits
- Runs with [fasthttp](https://github.com/vayala/fasthttp) for maximum performance and built-in anti-DoS features
- Whitelist/blacklist file types, and check them based on their headers, not the extension
- Sanitize files to prevent against phishing attacks
- Public/private mode (private by default)
- Zero-width file IDs in URLs
- File ID collision checking

## Setup

Make sure you have a Redis instance set up somewhere; preferably in the same environment as the file hoster.

- Put your config in `conf/` as `config.yml` using `conf/example.yml` as a reference.   
- Optionally, you can replace `favicon.ico` with your own icon! (It must have the same name)
- If you're using this with ShareX, check `example/lolihost.sxcu` for a template sxcu file.

Now you can choose to either run LoliHost with Docker or as a service on your system.

### Option 1: systemd/other service manager (Recommended)

- Download the binary to the same directory where `conf/` is located.
  - Alternatively you can build it.
- Mark it as executable with `chmod 0744 <binary file>`.
- Copy `example/lolihost.service` to `/lib/systemd/system` (if it's not there already).
- Edit the WorkingDirectory and ExecFile to match the locations of the binary file.
- Run `systemctl daemon-reload`.

At this point, you can run the program using `systemctl start lolihost`. You can check on its status by running `systemctl status lolihost`.

If anything goes wrong, you can check `journalctl -u lolihost` and find out what happened.


### Option 2: Docker

**Note that you will have to attach external volumes yourself should you use this method.**

- Build the image with `docker build -t lolihost .`
- Make sure to bind the port you choose (default is `3030`) to other ports on your system. Here's an example of how you would run it, after building the image.  
  
`docker container run -d -p 127.0.0.1:3030:3030 lolihost`  

### How to Upload

Create a POST request to `/upload` with a file in the field "file". Put the key in `Authorization` header

Set `?omitdomain=1`, if you don't want the host's original domain appended before the file name in the response. E.g: `a.png` instead of `https://a.com/a.png`

Add `?zerowidth=1` and set it to `1` to make your image URLs appear "zero-width". If you don't get what that means, try it, and see what happens.