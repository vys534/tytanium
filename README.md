![](https://i.imgur.com/AHHWH3z.png)

## Tytanium

Tytanium is a private file host program, meant for a single user or a small group. Here are the main highlights:

- SHA2-512 file encryption at rest, with an encryption key unique to each file
- Tune the server to exactly how you want with extensive customization options
- Works well with image capture suites, such as ShareX/MagicCap
- Good on system resources (<1MiB memory usage when idle)
- Limit how many requests/second to certain paths to prevent DoS attacks or an overloaded server
- Not written in Javascript! 

*Please note that files are NOT encrypted client-side; encryption is done on the server.*

### Setup

1. Download the binary in the Releases tab, or build the code from source.
2. Rename `example.yml` to `config.yml` and set the values you want, or create a `config.yml` from scratch.
3. Mark the binary as executable (this can be done with `chmod`).

### How to Upload

1. Create a POST request to `/upload` with a file in the field "file". Put the key in the `Authorization` header.
2. Set `?omitdomain=1`, if you don't want the host's original domain appended before the file name in the response. For example: `a.png` instead of `https://a.com/a.png`. This is useful if you have vanity/proxy domains you want to use.
3. The server will respond with JSON with fields `uri` and `encryption_key`. `uri` will be just the file name if `?omitdomain=1` was specified.

### Optional stuff

- You can use the [Size Checker](https://github.com/vysiondev/size-checker) program to make the `/stats` path produce values other than 0 for file count and total size used. Just tell it to check your files directory. You can run it as a cron job or run it manually whenever you want to update it. (If you choose not to use it, `/stats` will always return 0 for some fields.)
- If you want to change the favicon, replace `routes/favicon.ico` with your own image.

### License

[MIT License](LICENSE)