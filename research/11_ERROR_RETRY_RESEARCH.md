# Error and Retry Research in rclone

Rclone features a robust engine for handling various failure modes during cloud storage operations. This document breaks down its error handling and rate-limiting architecture.

## 1. Error Classification
Rclone classifies errors to determine whether an operation should be retried.
* Defined in `fs/fserrors/error.go` and `fs/fserrors/retriable_errors.go`.
* **Retriable Errors**: Network timeouts, connection resets, and HTTP 5xx errors. Functions like `fserrors.IsRetriable` determine this. (Source: `fs/fserrors/error.go`, Function: `IsRetriable`)
* **Fatal Errors**: Authentication failures (e.g., HTTP 401/403 not related to rate limits), invalid configurations, or missing files. (Source: `fs/fserrors/error.go`, Function: `IsFatalError`)
* **Temporary Errors**: Wraps standard Go `net.Error` temporary checks.

## 2. Retry Logic
* Handled primarily at the high-level operation layer.
* In `fs/operations/operations.go`, operations like `Copy` or `Sync` have retry loops for the entire operation if a retriable error occurs. (Source: `fs/operations/operations.go`)
* Lower-level retries (e.g., chunk retries) are sometimes handled within specific backends (e.g., S3 multipart retries).

## 3. Rate Limiting and Pacing
* Cloud APIs often rate-limit clients (HTTP 429 Too Many Requests).
* Rclone uses a "Pacer" mechanism to throttle requests and implement exponential backoff.
* Located at `fs/pacer.go`. The pacer monitors operations, and if a rate limit error is encountered, it sleeps and increases the delay for future operations. (Source: `fs/pacer.go`)

## 4. Minimum Reliable Implementation for `storage-bridge`
To maintain robustness while keeping the footprint small, `storage-bridge` should implement:
1. **Error Categorization Interface**: A simple way to classify `IsRetriable(err)` vs fatal errors.
2. **Exponential Backoff Engine**: A basic sleep-and-retry loop wrapped around API calls for HTTP 429 and 5xx errors.
3. **Transport-Level Retries**: Implement retries as a decorator pattern over the HTTP client transport, ensuring that connection drops are automatically retried before they bubble up to the streaming logic, simplifying the higher-level code.
