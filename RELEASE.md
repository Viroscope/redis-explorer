# Release Guide

This document describes how to create and publish releases for Redis Explorer.

## Automated Release Process

The project uses GitHub Actions to automatically build and publish releases for multiple platforms.

### Creating a Release

1. **Update the version** in `internal/ui/dialogs.go`:
   ```go
   AppVersion = "1.0.1"  // Update this
   ```

2. **Commit the version change**:
   ```bash
   git add internal/ui/dialogs.go
   git commit -m "Bump version to v1.0.1"
   git push origin main
   ```

3. **Create and push a version tag**:
   ```bash
   git tag -a v1.0.1 -m "Release v1.0.1"
   git push origin v1.0.1
   ```

4. **GitHub Actions will automatically**:
   - Build binaries for:
     - Linux (x64)
     - macOS (Intel & Apple Silicon)
     - Windows (x64)
   - Create compressed archives (.tar.gz for Linux/macOS, .zip for Windows)
   - Create a GitHub release with all artifacts
   - Publish release notes

5. **The release will be available at**: `https://github.com/Viroscope/redis-explorer/releases`

## Manual Build Process

If you need to build a release manually for your current platform:

### Using the Build Script

```bash
./build.sh v1.0.0
```

This will:
- Build an optimized binary for your platform
- Create a compressed archive in `./releases/`
- Include the README and icon in the archive

### Manual Build Commands

#### Linux
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o redis-explorer-v1.0.0-linux-amd64 .
tar -czf redis-explorer-v1.0.0-linux-amd64.tar.gz redis-explorer-v1.0.0-linux-amd64 README.md icon.png
```

#### macOS
```bash
# On macOS with native tools
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o redis-explorer-v1.0.0-darwin-amd64 .
tar -czf redis-explorer-v1.0.0-darwin-amd64.tar.gz redis-explorer-v1.0.0-darwin-amd64 README.md icon.png

# For Apple Silicon
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-s -w" -o redis-explorer-v1.0.0-darwin-arm64 .
tar -czf redis-explorer-v1.0.0-darwin-arm64.tar.gz redis-explorer-v1.0.0-darwin-arm64 README.md icon.png
```

#### Windows
```bash
# On Windows with native tools
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=1
go build -ldflags="-s -w -H=windowsgui" -o redis-explorer-v1.0.0-windows-amd64.exe .
7z a redis-explorer-v1.0.0-windows-amd64.zip redis-explorer-v1.0.0-windows-amd64.exe README.md icon.png
```

## Build Flags Explained

- `-ldflags="-s -w"`: Strip debugging information to reduce binary size
  - `-s`: Omit the symbol table
  - `-w`: Omit the DWARF symbol table
- `-H=windowsgui` (Windows only): Build as a GUI application (no console window)

## Platform-Specific Notes

### Linux
- Requires X11 development libraries
- CGO is required for OpenGL/GLFW support
- Builds on Ubuntu/Debian with: `libgl1-mesa-dev xorg-dev`

### macOS
- Requires Xcode Command Line Tools
- CGO is required
- Separate builds needed for Intel (amd64) and Apple Silicon (arm64)

### Windows
- Requires GCC toolchain (MinGW-w64)
- CGO is required
- The `-H=windowsgui` flag prevents a console window from appearing

## Cross-Compilation Limitations

Due to Fyne's dependency on CGO and platform-specific graphics libraries, cross-compilation is challenging. The automated workflow uses platform-specific runners (ubuntu-latest, windows-latest, macos-latest) to build native binaries.

For development builds, use the simple `go build` command on your target platform.

## Release Checklist

Before creating a release:

- [ ] Update version in `internal/ui/dialogs.go`
- [ ] Update CHANGELOG (if exists) or release notes
- [ ] Test the application on at least one platform
- [ ] Ensure all tests pass: `go test ./...`
- [ ] Commit all changes
- [ ] Create and push version tag
- [ ] Verify GitHub Actions workflow completes successfully
- [ ] Test downloaded binaries from the release
- [ ] Update release notes if needed

## Versioning

This project follows [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality in a backward compatible manner
- **PATCH** version for backward compatible bug fixes

Format: `vMAJOR.MINOR.PATCH` (e.g., `v1.0.0`, `v1.1.0`, `v1.1.1`)
