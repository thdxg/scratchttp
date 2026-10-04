# scratchttp

HTTP server from scratch in different languages.

- [x] go
- [ ] rust
- [ ] zig
- [ ] oCaml
- [ ] java
- [ ] odin

## Spec

Every implementation must serve these endpoints. Any other path returns `400 Bad Request` with an empty body.

| Method | Path     | Response                                                                 |
| ------ | -------- | ------------------------------------------------------------------------ |
| any    | `/`      | `200 OK` with body `hi`                                                  |
| `POST` | `/echo`  | `200 OK` with the request body as the response body                      |
| `GET`  | `/store` | `200 OK` with every stored value, joined by `\n`                         |
| `POST` | `/store` | Stores the request body, then `200 OK` with every stored value, joined by `\n` |

Any other method on `/echo` or `/store` returns `405 Method Not Allowed` with an empty body.

Every response:

- Uses `HTTP/1.1` and `\r\n` line endings.
- Has the header `Content-Type: plain/text`.
- Has the header `Location: http://<Host><path>`. `<Host>` is the value of the `Host` header without a trailing `/`, or empty if there is no `Host` header. The header name match is case-sensitive.
- Ends its headers with a blank line, followed by the body.

A malformed start line (not three space-separated parts) or a header line without a `:` gives a `400 Bad Request` status.

Values in `/store` are kept in memory for as long as the server runs. See [testdata.json](testdata.json) for exact requests and responses.

## Usage

Run the target http server listening on `localhost:$PORT`.

To test, run from project root:

```sh
mise run test
```

The `/store` tests expect an empty store, so restart the server before each run.
