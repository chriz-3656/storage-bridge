# Provider Comparison in rclone

This document compares five representative rclone providers: Local filesystem, Amazon S3, Google Drive, Backblaze B2, and Dropbox.

## 1. Authentication
* **Local**: No authentication required. Path configuration is handled via standard local OS access. (Source: `backend/local/local.go`)
* **Amazon S3**: Uses access key ID and secret access key, or IAM roles. Configuration is stored securely. (Source: `backend/s3/s3.go`)
* **Google Drive**: Uses OAuth2. Tokens are stored in the rclone config file and refreshed automatically via `lib/oauthutil` and drive-specific logic. (Source: `backend/drive/drive.go`)
* **Backblaze B2**: Uses application key ID and application key. Acquires authorization tokens via the B2 API. (Source: `backend/b2/b2.go`)
* **Dropbox**: Uses OAuth2, token refresh is handled similarly to Google Drive. (Source: `backend/dropbox/dropbox.go`)

*Note: General configuration is managed via `fs/config/config.go`.*

## 2. Initialization
* Every backend registers itself using `fs.Register` which points to a `NewFs` function.
* **Local**: `NewFs` resolves absolute paths, checks if the path is a file or directory, and initializes a `Fs` struct. (Source: `backend/local/local.go`, Function: `NewFs`)
* **Cloud Providers**: `NewFs` reads config, initializes HTTP clients (often with `fs/fshttp/http.go`), performs an initial auth/stat check, and returns the constructed backend instance. (e.g., `backend/s3/s3.go`, `backend/drive/drive.go`)

## 3. Operations
Rclone abstracts operations into interfaces like `fs.Fs` and `fs.Object`.
* **List**: Implemented via `List` returning entries. (e.g., `backend/s3/s3.go`)
* **Stat**: Retrieves metadata. Often combined with object creation like `NewObject`. (e.g., `backend/drive/drive.go`)
* **Get/Download**: Generally handled via `Open` on `fs.Object`, returning an `io.ReadCloser`. (e.g., `backend/b2/b2.go`)
* **Put/Upload**: Handled via `Put` or `PutStream` on `fs.Fs`. (e.g., `backend/local/local.go`)
* **Delete**: `Remove` on `fs.Object`. (e.g., `backend/dropbox/dropbox.go`)
* **Copy/Move**: Some backends support server-side copy (`Copy` on `fs.Copier`) or move (`Move` on `fs.Mover`). Otherwise, it falls back to download+upload in `fs/operations/operations.go`.
* **Mkdir/Rmdir**: Handled natively by backends.
* **Streaming**: Backends indicate support for streaming uploads without known sizes via `PutStream`. If not supported, rclone might buffer. (Source: `fs/operations/operations.go`)

## Comparison Table

| Provider | Auth Type | Server-Side Copy | PutStream (Unknown Size Upload) |
|---|---|---|---|
| **Local** | None | N/A | Yes |
| **S3** | Keys/IAM | Yes | Yes (Multipart) |
| **Google Drive** | OAuth2 | Yes | Yes |
| **Backblaze B2** | Keys | Yes | Yes |
| **Dropbox** | OAuth2 | Yes | Yes |

*Note: Concurrency and multipart uploads are often configurable per backend (e.g., `chunk_size` in S3 and Drive).*
