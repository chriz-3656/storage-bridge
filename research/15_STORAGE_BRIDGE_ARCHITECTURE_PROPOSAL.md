# Storage Bridge Architecture Proposal

## Architecture Overview

```text
       [ CLI ]        [ MCP/STDIO ]      [ HTTP (Optional) ]
          |                 |                     |
          +-----------------+---------------------+
                            |
                    [ Core Engine ]
               (Routing, Error Handling, 
               Concurrency, Streaming logic)
                            |
                  [ Provider Interface ]
                            |
          +-----------------+---------------------+
          |                 |                     |
  [ S3 Adapter ]     [ Local Adapter ]   [ Drive Adapter ] ...
          |                 |                     |
    [ AWS Cloud ]     [ File System ]    [ Google Cloud ]
```

## 1. Core Engine
The core engine acts as the brain. It is responsible for:
- Routing requests to the appropriate provider adapter.
- Enforcing global limits (concurrency, rate limiting).
- Handling retries and standardizing errors (e.g. wrapping HTTP 429 errors into standard retryable signals).
- Exposing simple, clean methods to the presentation layers (CLI, MCP, HTTP).

## 2. Provider Interface and Adapters
The Provider Interface will be a heavily simplified version of rclone's `fs.Fs` and `fs.Object`.
By utilizing a capability detection model rather than a massive interface, we can keep the core tight. 
Adapters simply implement the methods they can support. For example:
- **Core capabilities (Required)**: `Stat`, `List`, `Get` (stream), `Put` (stream), `Remove`.
- **Optional capabilities**: `Move`, `Copy`, `Mkdir`, `Quota`, `MultipartUpload`.

## 3. Configuration & Credential Manager
A dedicated sub-system that manages reading/writing to the local OS-specific config directories.
- **Configuration** defines aliases, routing, and preferences.
- **Credentials** defines secrets, safely isolated.

## 4. Presentation Layers
- **CLI**: Powered by `cobra` but leaner than rclone's implementation. Focuses on straightforward commands (`cp`, `mv`, `rm`, `ls`) mapping directly to Core Engine calls.
- **MCP/STDIO**: Listens on standard I/O for JSON-RPC messages, conforming to the Model Context Protocol, parsing requests, calling the Core Engine, and emitting JSON responses.

## 5. Streaming Data Flow
Following rclone's strengths, all data transfers are `io.Reader` and `io.Writer` driven. 
Large files are chunked into standard byte buffers that pool memory to avoid garbage collection spikes and RAM exhaustion. There is no intermediate disk caching for direct transfers.
