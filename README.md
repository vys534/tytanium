![](https://i.imgur.com/jkB0XlM.png)

# Tytanium

A simple file host server, intended for personal or small group use.

## Features

- Tune the server to exactly how you want with extensive customization options
- Built with [fasthttp](https://github.com/vayala/fasthttp) for performance and built-in anti-DoS features
- Whitelist/blacklist checks done via file headers rather than extensions
- Sanitize files to prevent against phishing attacks (Change their Content-Type to text/plain)
- Zero-width file IDs in URLs - paste invisible but functional links!
- File ID collision checking
- Works well with image capture suites, such as ShareX/MagicCap
- Not written in Javascript! 

### Setup

- Download the binary or build this program
- Rename `example.yml` to `config.yml` and set the values you want
- Start the binary
- Done
- **Optional stuff** 
  - You can use the [Size Checker](https://github.com/vysiondev/size-checker) program to make the `/stats` path produce values other than 0 for file count and total size used. Just tell it to check your files directory. You can run it as a cron job or run it manually whenever you want to update it. (If you choose not to use it, `/stats` will always return 0 for every field.)
  - If you want to change the favicon, replace `routes/favicon.ico` with your own image.

### How to Upload

Create a POST request to `/upload` with a file in the field "file". Put the key in the `Authorization` header.

Set `?omitdomain=1`, if you don't want the host's original domain appended before the file name in the response. For example: `a.png` instead of `https://a.com/a.png`

Add `?zerowidth=1` and set it to `1` to make your image URLs appear "zero-width". If you don't get what that means, try it, and see what happens.