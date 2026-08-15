# Build and Binary Size Research

## Dependencies
- Rclone has a massive `go.mod` file, containing hundreds of dependencies because it natively supports over 70 storage providers. (Source: `go.mod`)
- Large SDKs such as `aws-sdk-go-v2`, `azure-sdk-for-go`, `dropbox-sdk-go-unofficial` are all imported simultaneously. (Source: `go.mod`)

## Backend Registration and Build Tags
- All backends are imported via side-effect in `backend/all/all.go` (e.g., `_ "github.com/rclone/rclone/backend/s3"`). (Source: `backend/all/all.go`)
- `rclone.go` imports `backend/all/all.go`, forcing the Go compiler to include all backends in the final binary. (Source: `rclone.go`)
- Build tags (`GOTAGS`) can be passed via the `Makefile` (Source: `Makefile`, Line: 42), but rclone doesn't use build tags by default to exclude standard backends—they are all compiled together into one monolithic binary.

## Binary Size Considerations
- Including all SDKs (AWS, Azure, GCP, Dropbox, Box, etc.) significantly increases the binary size.
- To achieve a smaller, single standalone binary for `storage-bridge` with no runtimes, we must ONLY import the specific backends we intend to support (e.g., just AWS S3, Azure, GCP).
- We should avoid the `backend/all/all.go` approach if we want modular builds. Instead, we should use build tags per provider or simply only import the needed ones to aggressively strip out unused backend logic.

## Cross Compilation
- Rclone uses standard Go cross-compilation. `Makefile` uses `GOOS` and `GOARCH`. (Source: `Makefile`, Line: 33)
- Windows resource compilation is done using `bin/resource_windows.go` to attach icons/metadata. (Source: `Makefile`, Line: 51)
- LDFLAGS are used to inject version variables (`-X github.com/rclone/rclone/fs.Version=$(TAG)`) and strip debugging symbols (`-s`) to reduce binary size. (Source: `Makefile`, Line: 46)
