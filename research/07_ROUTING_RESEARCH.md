# Routing and Union Research

This document outlines how `rclone` handles union, overlay, and routing of multiple remotes, and what concepts can be adapted for `storage-bridge`.

## rclone Union Architecture
The `union` package implements a virtual provider that joins multiple existing remotes ("upstreams") into a single filesystem.
- Source: [backend/union/union.go](file:///home/demo/rclone/backend/union/union.go#L1) (Package union)
- Source: [backend/union/union.go](file:///home/demo/rclone/backend/union/union.go#L71) (`Fs` struct which maintains a slice of `*upstream.Fs`)

### Combining Remotes
Remotes are provided as a space-separated list to the `upstreams` parameter when initializing the union backend.
- Source: [backend/union/union.go](file:///home/demo/rclone/backend/union/union.go#L37) (upstreams Option)

### Routing Decisions & Policies
Routing decisions in `rclone` are governed by "policies" categorized by operation type:
1. `Action`: Modifying files/directories (e.g., Delete)
2. `Create`: Creating files/directories
3. `Search`: Accessing files/directories
- Source: [backend/union/policy/policy.go](file:///home/demo/rclone/backend/union/policy/policy.go#L18) (`Policy` interface)

**Available Policies:**
rclone provides numerous policies:
- `epall`, `epff`, `eplfs`, `eplno`, `eplus`, `epmfs`, `eprand`: Path-preserving (ep) policies.
- `ff` (First Found), `lfs` (Least Free Space), `mfs` (Most Free Space), `rand` (Random), `newest` (Newest file).
- Source: [backend/union/policy/](file:///home/demo/rclone/backend/union/policy/) (List of policy implementations)

### Fallback and Failure Handling
Failure handling is built into the policies and multithreading functions in the union package (e.g., returning the first successful remote or aggregating errors).
- Not established from source; requires further investigation into specific retry loops within union package.

## Useful Concepts for storage-bridge
A lightweight routing layer can be independently implemented for `storage-bridge` using the following ideas:
- **Routing Categories**: Separate read (Search) from write (Create/Action) operations.
- **Policies**: Implement a simple subset of policies:
  - `automatic` / `round-robin` / `least-used` (equivalent to lfs/mfs or random)
  - `priority` (similar to First Found, trying remotes in order)
  - `path-based` (custom routing based on path prefixes)
- **Abstraction**: Have a router component that delegates the actual FS operations to the underlying providers based on the selected policy.
