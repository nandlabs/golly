# Cloud Backends for `vfs`

The core `vfs` package ships with a `local://` (and bare-path) provider only. Cloud
backends live in separate satellite modules so their SDK dependencies stay
isolated — consumers only pay for what they import.

## Available providers

| Scheme        | Module                                                                          | Backend       |
| ------------- | ------------------------------------------------------------------------------- | ------------- |
| `s3://`       | [`nandlabs/golly-aws/s3`](https://github.com/nandlabs/golly-aws/tree/main/s3)   | AWS S3        |
| `gs://`       | [`nandlabs/golly-gcp/gs`](https://github.com/nandlabs/golly-gcp/tree/main/gs)   | Google Cloud Storage |
| _(planned)_   | `vfs/azblob` (not yet implemented)                                              | Azure Blob Storage |

Each provider blank-imports register itself with `vfs.Manager`, so usage looks like:

```go
import (
    "oss.nandlabs.io/golly/vfs"
    _ "oss.nandlabs.io/golly-aws/s3"   // registers s3://
    _ "oss.nandlabs.io/golly-gcp/gs"   // registers gs://
)

func main() {
    f, _ := vfs.Open("s3://my-bucket/path/to/object")
    defer f.Close()
    // ...
}
```

## Implementing a new cloud provider

To add support for a new backend (e.g. Azure Blob Storage), follow the pattern
used by `golly-aws/s3` and `golly-gcp/gs`:

1. **Create a new Go module** under your own org (or as a satellite of golly) so the cloud SDK dependency stays out of the core `golly` module.
2. **Implement `VFile` and `VFileSystem`** from `oss.nandlabs.io/golly/vfs`. Both expose `io.ReadWriteSeeker` for streaming and a standard `Walk` / `List` / `Remove` / `Copy` / `Move` API — see the local provider in [`localfs.go`](localfs.go) for a reference shape.
3. **Auto-register on import** with `func init() { vfs.Manager.Register("scheme", &Provider{}) }` so a blank-import does the wiring.
4. **Add a README** with auth examples (see `golly-aws/s3/README.md` for the conventions used by the existing satellites).

## Status

- S3 and GCS are production-ready via the existing satellite modules.
- An Azure Blob Storage adapter (`vfs/azblob` or a future `golly-azure/blob`) is on the roadmap but not yet implemented. PRs welcome.

For broader context on the gap and roadmap, see [issue #180](https://github.com/nandlabs/golly/issues/180).
