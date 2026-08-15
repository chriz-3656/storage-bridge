# Storage Bridge

[![Release](https://img.shields.io/github/v/release/chriz-3656/storage-bridge)](https://github.com/chriz-3656/storage-bridge/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**Storage Bridge** is a blazing-fast, AI-native CLI and daemon that unites AWS S3, Google Drive, Local filesystems, and more into a single, scriptable operating layer. Built natively in Go.

---

## ⚡ Features

- **Stateful Navigation**: Traverse your cloud storage just like a local terminal (`cd`, `ls`, `pwd`, `mkdir`).
- **Universal Provider API**: One interface to interact with Local, AWS S3, Google Drive, Dropbox, Backblaze B2, and Cloudflare R2.
- **Union Load Balancing**: Combine multiple upstreams (e.g., S3 + Local) into a single virtual pool with customizable routing policies.
- **Concurrent Sync Engine**: State-of-the-art GoRoutine sync system performing in-memory diffs and fast parallel data transfers.
- **AI-Native (MCP)**: Implements the **Model Context Protocol (JSON-RPC)**, allowing autonomous AI agents (like Claude or Antigravity) to manage your files natively.

## 📦 Installation

Download the single standalone binary for your architecture. Zero dependencies required.

### Download

Head to the [Downloads Page](https://chriz-3656.github.io/storage-bridge/download.html) to automatically get the correct artifact for Windows, macOS (Apple Silicon/Intel), or Linux.

Alternatively, fetch the latest release from the [GitHub Releases](https://github.com/chriz-3656/storage-bridge/releases).

### Build from Source

```bash
git clone https://github.com/chriz-3656/storage-bridge.git
cd storage-bridge
go build -o storage-bridge ./cmd/storage-bridge
```

## 🚀 Quick Start

### 1. Authenticate a Provider
```bash
storage-bridge auth login google
storage-bridge provider add my-s3 --type s3 --access-key <KEY> --secret-key <SECRET>
```

### 2. Stateful Navigation
```bash
$ storage-bridge cd my-s3:/backups
$ storage-bridge ls
$ storage-bridge upload local-file.zip
```

### 3. High-Speed Sync
```bash
storage-bridge sync local:./data my-s3:/backups --workers 8
```

## 🤖 AI Agent Integration (MCP)

Storage Bridge is the first cloud filesystem built for agents. Run the MCP server to expose your files to LLMs:

```bash
storage-bridge mcp --port 8080
```

Agents can now communicate with Storage Bridge using standard JSON-RPC:
```json
{"jsonrpc":"2.0","method":"list","params":{"remote":"local:Documents/"}}
```

## 🤝 Contributing

Contributions are what make the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

Please check out our [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---
Built by **Chriz (chriz-3656)** - Open Source Creator & Cloud Architect.
