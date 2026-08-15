# License and Reuse Boundaries

## Project License
- Rclone is licensed under the MIT License. (Source: `COPYING`, Lines: 1-21)
- The license permits reuse, modification, and commercial distribution provided the copyright notice and permission notice are included in all copies or substantial portions of the software.

## Reuse Strategy for `storage-bridge`

**A. Architectural ideas we can independently implement:**
- The unified `Fs` backend interface concept.
- Side-effect `init()` registration for plugins and commands using Cobra (though we will keep it much smaller).

**B. Interfaces we can redesign:**
- Rclone's internal file system interfaces (`fs.Fs`) are complex due to supporting 70+ diverse backends. We can design simpler interfaces tailored only for our specific target backends.
- Path resolution and filtering can be simplified.

**C. Code that may potentially be reused:**
- Core generic utilities that don't depend on global state, if absolutely necessary.
- If we copy any code directly, we MUST include the rclone MIT license header and `COPYING` text in those specific files to comply with the license.

**D. Code/deps that require careful review:**
- Large SDK dependencies in `go.mod`. We should only import what we explicitly need to minimize binary size and audit surface. 
- Rclone's custom wrappers around API clients (e.g., `backend/s3/s3.go` wrapping `aws-sdk-go-v2`). We should use official SDKs directly where possible to keep maintenance low.

**E. Things we should NOT copy without review:**
- `cmd/all/all.go` and `backend/all/all.go` paradigms, which bloat the binary by importing everything unconditionally.
- Rclone's configuration parsing (`fs/config`), which has a huge surface area for all backends. We should build our own simplified, strict config logic.
