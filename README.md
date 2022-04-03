![](https://i.imgur.com/AHHWH3z.png)

## Tytanium

Tytanium is a private file host program, meant for a single user or a small group. Here are the main highlights:

- SHA2-512 file encryption at rest, with an encryption key unique to each file
- Tune the server to exactly how you want with extensive customization options
- Works well with image capture suites, such as ShareX/MagicCap
- Good on system resources (<1MiB memory usage when idle)
- Limit how many requests/second to certain paths to prevent DoS attacks or an overloaded server
- Zero-width strings: make your links appear invisible! (Example: https://example.com/file.png?enc_key=X appears as https://example.com/)
- Not written in Javascript! 

*Please note that files are NOT encrypted client-side; encryption is done on the server.*

### Setup

1. Download the binary in the Releases tab, or build the code from source.
2. Rename `example.yml` to `config.yml` and set the values you want, or create a `config.yml` from scratch.
3. Mark the binary as executable (this can be done with `chmod`).

### Upload & Response

Create a POST request to `/upload` with a file in the field "file". Put the key in the `Authorization` header.

Query arguments you can pass:
- `?zerowidth=1`: Zero-width links. By enabling this, you can turn a URL's path invisible (See example above in the feature list).
  - *`uri` will be zero-width if you specified `?zerowidth=1`.*

Example response on success:

```json
{
   "status": 0,
   "data": {
      "uri": "https://example.com/file.png?enc_key=ABCDEF",
      "path": "file.png?enc_key=ABCDEF",
      "file_name": "file.png",
      "encryption_key": "ABCDEF"
   }
}
```

If there's any error, the response will look like this. Status code `1` means a generic error, `2` means something broke internally. `message` will contain the error message.

```json
{
   "status:": 1,
   "message": "Error message"
}
```

### Optional stuff

- You can use the [Size Checker](https://github.com/vysiondev/size-checker) program to make the `/stats` path produce values other than 0 for file count and total size used. Just tell it to check your files directory. You can run it as a cron job or run it manually whenever you want to update it. (If you choose not to use it, `/stats` will always return 0 for some fields.)
- If you want to change the favicon, replace `routes/favicon.ico` with your own image.

### License

[MIT License](LICENSE)