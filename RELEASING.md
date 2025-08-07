# Releasing

This document describes the release process for this project.

## Prerequisites

- You must have `goreleaser` installed. You can install it with:
  ```bash
  go install github.com/goreleaser/goreleaser@latest
  ```
- You must have a `GITHUB_TOKEN` environment variable with `repo` scope.

## Release Process

1. **Create a new git tag:**

   ```bash
   git tag -a v0.0.2 -m "Release v0.0.2"
   ```

2. **Push the tag to GitHub:**

   ```bash
   git push origin v0.0.2
   ```

3. **Run GoReleaser:**

   GoReleaser will automatically build the binaries, create a GitHub release, and upload the assets.

   ```bash
   goreleaser release --clean
   ```
