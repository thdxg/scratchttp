# scratchttp

HTTP server from scratch in different languages.

- [x] go
- [ ] rust
- [ ] zig
- [ ] oCaml
- [ ] java
- [ ] odin

## Spec

| Method | Path     | Response                                                                       |
| ------ | -------- | ------------------------------------------------------------------------------ |
| any    | `/`      | `200 OK` with body `hi`                                                        |
| `POST` | `/echo`  | `200 OK` with the request body as the response body                            |
| `GET`  | `/store` | `200 OK` with every stored value, joined by `\n`                               |
| `POST` | `/store` | Stores the request body, then `200 OK` with every stored value, joined by `\n` |

Other methods on `/echo` or `/store` return `405 Method Not Allowed`. Other paths return `404 Not Found`.

**Requests.** Headers end at a blank line or end of input. The body is exactly `Content-Length` bytes (empty if absent). Header names are case-insensitive.

**400 Bad Request** for: a start line that isn't three space-separated parts, a header without `:`, an invalid or over-1024 `Content-Length`, or a body shorter than `Content-Length`.

**Responses** use `HTTP/1.1` and `\r\n`, with these headers in order:

    Content-Type: text/plain
    Location: http://<Host><path>
    Content-Length: <body length>

`<Host>` is the `Host` value without a trailing `/`, or empty. Error responses (400, 404, 405) have empty bodies. The connection closes after each response, and `/store` values live in memory while the server runs.

See [testdata.json](testdata.json) for exact requests and responses.
## Usage

Run the target http server listening on `localhost:$PORT`.

To test, run from project root:

```sh
mise run test
```

The `/store` tests expect an empty store, so restart the server before each run.
