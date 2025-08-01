# Go Version Upgrade Guide

This document outlines the process for upgrading the Go version used in the prometheus-rds-exporter project.

## Overview

The prometheus-rds-exporter project uses Go for building the application and relies on specific Go versions defined in multiple configuration files. When upgrading Go, all references must be updated consistently.

## Files to Update

When upgrading Go version, the following files need to be updated:

### 1. go.mod
- Update the `go` directive to specify the new Go version
- Update the `toolchain` directive to match the new Go version

**Example:**
```go
go 1.24.0

toolchain go1.24.0
```

### 2. GitHub Actions Workflows

Update the `go-version` in the `actions/setup-go@v5` step

**Example:**
```yaml
- uses: actions/setup-go@v5
  with:
    go-version: '1.24'
```

Update the Go version in all workflow files:
- `.github/workflows/build.yaml`
- `.github/workflows/test.yaml`
- `github/workflows/linter.yaml`

### 3. Local development environment

- Update golang base image version (`FROM`) in the `scripts/prometheus/Dockerfile`

## Upgrade Process

1. Check for any deprecated features or breaking changes in the new Go version
2. Update `go.mod` with the new Go version and toolchain
3. Run tests
   1. Run local tests to ensure compatibility:
      ```bash
      make test
      make lint
      make build
      ```
   2. Update Go dependencies if required:
      ```bash
      go mod tidy
      go mod download
      ```
4. Update GitHub Actions workflows with the specific Go version
5. Commit these changes
6. Verify GitHub Actions pass:
   - Push changes to a feature branch
   - Ensure all CI checks pass (linter, tests, build)
7. Update this document with any new considerations
8. Update README.md if it mentions specific Go version requirements

## Considerations

### Compatibility

- Ensure all dependencies support the new Go version
- Check for any breaking changes in the Go release notes
- Verify that the Docker base images support the new Go version (if using Go-based images)

### Testing

- Run the full test suite locally before pushing
- Monitor CI/CD pipelines for any failures
- Test the built binary in a staging environment

### Rollback Plan

If issues arise after the upgrade:
1. Revert the configuration files to the previous Go version
2. Run tests to ensure stability
3. Investigate and resolve the compatibility issues
4. Retry the upgrade process

## Useful Commands

```bash
# Overrides golang version for test if not installed locally
GOTOOLCHAIN=go1.24.0 make test

# Check current Go version
go version

# Update dependencies
go mod tidy

# Run tests
make test

# Build the application
make build

# Run linter
make lint

# Check for security issues
make checkcov
```

## References

- [Go Release Notes](https://golang.org/doc/devel/release.html)
- [Go Module Reference](https://golang.org/ref/mod)
- [GitHub Actions Go Setup](https://github.com/actions/setup-go)