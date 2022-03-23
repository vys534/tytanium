![](https://i.imgur.com/AHHWH3z.png)

# Tytanium

A durable and secure private file host solution. Intended for personal or small group use.

## Features

- Tune the server to exactly how you want with extensive customization options
- Built with [fasthttp](https://github.com/vayala/fasthttp) for optimal performance instead of the native http module
- File whitelist/blacklist type checks done via file headers rather than extensions
- Sanitize file type from rendering in HTML/other types to mitigate phishing attacks (Change their Content-Type to text/plain)
- Option to return a zero-width file IDs in a URL after upload - paste invisible but functional links!
- Works well with image capture suites, such as ShareX/MagicCap
- Good on system resources (<1MiB memory usage when idle)
- Limit how many requests/second to certain paths to prevent DoS attacks or an overloaded server
- Not written in Javascript! 

### Setup

1. Download the binary in the Releases tab, or build the code from source.
2. Rename `example.yml` to `config.yml` and set the values you want, or create a `config.yml` from scratch.
3. Start the binary with your method of choice.
4. Done! 

### How to Use

1. Create a POST request to `/upload` with a file in the field "file". Put the key in the `Authorization` header.
2. Set `?omitdomain=1`, if you don't want the host's original domain appended before the file name in the response. For example: `a.png` instead of `https://a.com/a.png`. This is useful if you have vanity/proxy domains you want to use.
3. Add `?zerowidth=1` and set it to `1` to make your image URLs appear "zero-width". If you don't get what that means, try it, and see what happens.
4. The server will respond with a link to the file (or just the file name if you set `?omitdomain=1`). It will be just text so no need to parse any JSON.

### Optional stuff

- You can use the [Size Checker](https://github.com/vysiondev/size-checker) program to make the `/stats` path produce values other than 0 for file count and total size used. Just tell it to check your files directory. You can run it as a cron job or run it manually whenever you want to update it. (If you choose not to use it, `/stats` will always return 0 for some fields.)
- If you want to change the favicon, replace `routes/favicon.ico` with your own image.

### License

[MIT License](LICENSE)