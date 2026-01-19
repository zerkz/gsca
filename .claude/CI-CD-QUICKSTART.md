# CI/CD Quick Start Guide

## What Was Set Up

GitHub Actions workflows for automated testing and releases:

### Continuous Integration (CI)
**File**: `.github/workflows/ci.yml`

Automatically runs on every push and pull request:
- **Tests**: Runs on Linux, macOS, Windows × Go 1.21, 1.22, 1.23
- **Linting**: Code quality checks with golangci-lint
- **Build**: Verifies cross-platform compilation

### Release Automation
**Files**: `.github/workflows/release.yml`, `.goreleaser.yaml`

Uses [GoReleaser](https://goreleaser.com/) for releases. Automatically triggers when you push a version tag:
- Builds binaries for 6 platforms (Linux, macOS, Windows × AMD64, ARM64)
- Creates archives (tar.gz for Linux/macOS, zip for Windows)
- Generates SHA256 checksums
- Creates GitHub release with all artifacts
- Auto-generates changelog from commit history

### Configuration
**File**: `.golangci.yml`

Linter configuration with sensible defaults for Go projects.

## How to Use

### For Development

1. **Make changes** to your code
2. **Run tests locally**:
   ```bash
   go test ./...
   ```
3. **Run linter** (optional):
   ```bash
   golangci-lint run
   ```
4. **Push to GitHub**:
   ```bash
   git push origin your-branch
   ```
5. **Open a Pull Request** - CI runs automatically!

### For Releases

1. **Tag a version**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **Wait 2-5 minutes** - GitHub Actions builds everything

3. **Check Releases page** - All binaries ready!
   ```
   https://github.com/zerkz/gsca/releases
   ```

## Status Badges

Added to README.md:
- ![CI](https://github.com/zerkz/gsca/workflows/CI/badge.svg) - Shows test status
- ![Release](https://github.com/zerkz/gsca/workflows/Release/badge.svg) - Shows release status
- [![Go Report Card](https://goreportcard.com/badge/github.com/zerkz/gsca)](https://goreportcard.com/report/github.com/zerkz/gsca) - Code quality score

## Viewing Workflow Status

1. Go to your repository on GitHub
2. Click the **Actions** tab
3. See all workflow runs and their status

## Files Created

```
.github/
├── workflows/
│   ├── ci.yml              # CI workflow
│   └── release.yml         # Release workflow (GoReleaser)
├── CONTRIBUTING.md         # Contributor guidelines
├── README.md              # GitHub directory info
└── WORKFLOWS.md           # Detailed workflow docs

.goreleaser.yaml           # GoReleaser configuration
.golangci.yml              # Linter configuration
CI-CD-QUICKSTART.md        # This file
```

## Next Steps

1. **Push to GitHub** to activate workflows
2. **Create a test PR** to see CI in action
3. **Tag v0.1.0** to test release workflow
4. **Monitor Actions tab** to see builds running

## Troubleshooting

### Tests fail in CI but pass locally?

Check:
- Platform-specific code (use `filepath.Join()` not `/` or `\`)
- Timing issues in tests
- Environment assumptions

### Release doesn't trigger?

Check:
- Tag format: Must start with `v` (v1.0.0, v2.1.0)
- Tag was pushed: `git push origin v1.0.0`
- Workflow file exists in main branch

### Want to test locally before pushing?

**Test GoReleaser config:**
```bash
# Validate config
goreleaser check

# Test build without releasing
goreleaser build --snapshot --clean
```

**Run GitHub Actions locally with [act](https://github.com/nektos/act):**
```bash
# Install act
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run CI workflow
act pull_request

# Run release workflow (requires tag)
act --job release
```

## Documentation

For more details, see:
- **[WORKFLOWS.md](.github/WORKFLOWS.md)** - Complete workflow documentation
- **[CONTRIBUTING.md](.github/CONTRIBUTING.md)** - Contribution guidelines
- **[GitHub Actions Docs](https://docs.github.com/en/actions)** - Official documentation

## Summary

**You now have**:
- Automated testing on 3 OSes × 3 Go versions
- Code quality checks with linting
- Automatic binary builds for 6 platforms
- One-command releases with `git tag`
- Professional CI/CD badges

**No manual builds ever again!**
