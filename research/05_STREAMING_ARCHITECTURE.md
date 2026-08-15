# Streaming Architecture in rclone

This document outlines how rclone manages file data streaming, focusing on moving data between input and output with minimal memory footprint.

## 1. Input and Output Streams
* Rclone models all data transfers natively using Go's `io.Reader` and `io.Writer`.
* **Source (`fs.Object`)**: Downloads provide an `io.ReadCloser` via the `Open` method. (Source: `fs/fs.go`, Type: `Object`, Function: `Open`)
* **Destination (`fs.Fs`)**: Uploads consume an `io.Reader` via the `Put` or `PutStream` methods. (Source: `fs/fs.go`, Type: `Fs`, Function: `Put`)

## 2. Moving File Data
* The primary flow for moving data between endpoints happens in `fs/operations/operations.go` inside functions like `Copy`.
* `Copy` coordinates opening the source object, optionally wrapping it in accounting readers for stats, and passing it to the destination's `Put`. (Source: `fs/operations/operations.go`, Function: `Copy`)
* **Accounting/Stats**: Data flow is monitored by wrapping the `io.Reader` using `fs/accounting/accounting.go`.

## 3. Buffering and Asynchronous Readers
* Rclone employs asynchronous reading to decouple download and upload speeds, preventing slow uploads from blocking downloads and vice versa.
* Implemented using `fs/asyncreader/asyncreader.go` which provides a bounded buffer read-ahead.
* `fs/chunkedreader/chunkedreader.go` handles chunking streams for backends that require specific part sizes for multipart uploads.

## 4. Large-File Handling & Resumable Uploads
* Streaming avoids loading entire files into memory. Backends read in chunks.
* For backends supporting multipart (e.g., S3, B2), the stream is uploaded in pieces. If a chunk fails, it can sometimes be retried without restarting the whole file, though pure streaming limits this capability because streams cannot be arbitrarily rewound without a local buffer or disk spool.
* Temporary files are generally avoided unless explicitly configured (e.g., vfs cache). (Not established from source; requires further investigation for exact fallback locations).

## 5. Lessons for `storage-bridge`
To preserve a streaming-first architecture while remaining smaller:
1. **Zero Temporary Files**: Enforce `io.Reader` to `io.Writer` piping using bounded buffers (like `io.CopyBuffer` or custom async readers) instead of spooling to disk.
2. **Minimal Memory**: Keep chunk sizes configurable but small by default. Ensure streaming architectures do not eagerly buffer the whole payload.
3. **Pipelining**: Chain transformations (e.g., encryption, hashing) directly into the `io.Reader` chain using `io.TeeReader` and standard Go interfaces.
