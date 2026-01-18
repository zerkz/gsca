# GitHub Actions Workflows

This repository uses GitHub Actions for continuous integration and automated releases.

## Workflows

### CI Workflow (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main` or `master` branches
- Pull requests targeting `main` or `master` branches

**Jobs:**

#### 1. Test
- **Matrix**: Tests run on 3 operating systems × 3 Go versions = 9 combinations
  - OS: Ubuntu, macOS, Windows
  - Go versions: 1.21, 1.22, 1.23
- **Steps**:
  - Checkout code
  - Setup Go with specified version
  - Cache Go modules for faster builds
  - Download and verify dependencies
  - Run tests with race detection and coverage
  - Upload coverage to Codecov (Ubuntu + Go 1.21 only)

#### 2. Lint
- **Platform**: Ubuntu only
- **Steps**:
  - Checkout code
  - Setup Go 1.21
  - Run `golangci-lint` with configuration from `.golangci.yml`
  - Checks for code quality issues, formatting, and best practices

#### 3. Build
- **Platform**: Ubuntu only
- **Steps**:
  - Builds binaries for Linux, Windows, and macOS
  - Verifies that code compiles for all target platforms
  - Uses `-ldflags="-s -w"` to strip debug info and reduce binary size

### Release Workflow (`.github/workflows/release.yml`)

**Triggers:**
- Push of tags matching `v*` (e.g., `v1.0.0`, `v2.1.3`)

**Jobs:**

#### Build and Release
- **Platform**: Ubuntu
- **Steps**:
  1. Checkout code with full history
  2. Extract version from git tag
  3. Build binaries for 6 platform/architecture combinations:
     - Linux AMD64
     - Linux ARM64
     - Windows AMD64
     - Windows ARM64
     - macOS AMD64 (Intel)
     - macOS ARM64 (Apple Silicon)
  4. Generate SHA256 checksums for all binaries
  5. Generate changelog from git commits since last tag
  6. Create GitHub release with:
     - All binaries
     - Checksums file
     - Auto-generated changelog
     - Installation instructions

**Build Flags:**
- `-s -w`: Strip debug symbols and DWARF info (smaller binaries)
- `-X main.version=<version>`: Embed version string in binary

## Configuration Files

### `.golangci.yml`
Linter configuration with enabled checks:
- `errcheck`: Check for unchecked errors
- `gosimple`: Suggest code simplifications
- `govet`: Report suspicious constructs
- `ineffassign`: Detect ineffectual assignments
- `staticcheck`: Advanced static analysis
- `unused`: Find unused code
- `gofmt`: Check formatting
- `goimports`: Check import formatting
- `misspell`: Find spelling mistakes
- `unconvert`: Remove unnecessary type conversions
- `goconst`: Find repeated strings that could be constants
- `gocyclo`: Calculate cyclomatic complexity
- `dupl`: Find duplicated code

## Using the Workflows

### For Contributors

1. **Fork the repository**
2. **Create a feature branch**
3. **Make changes and commit**
4. **Push to your fork**
5. **Open a Pull Request**

GitHub Actions will automatically:
- Run all tests on multiple platforms
- Check code quality with linters
- Verify builds for all platforms

You'll see status checks in your PR showing pass/fail status.

### For Maintainers

#### Creating a Release

1. **Ensure all tests pass on main branch**

2. **Tag the release**:
   ```bash
   git checkout main
   git pull origin main
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

3. **Monitor the release workflow**:
   - Go to Actions tab on GitHub
   - Watch the "Release" workflow
   - Wait for completion (~2-5 minutes)

4. **Verify the release**:
   - Go to Releases page
   - Check that all binaries are attached
   - Verify checksums
   - Edit release notes if needed

#### Debugging Failed Workflows

1. **Go to Actions tab**
2. **Click on the failed workflow**
3. **Click on the failed job**
4. **Expand the failed step**
5. **Read the error logs**

Common issues:
- **Test failures**: Fix tests locally first
- **Lint errors**: Run `golangci-lint run` locally
- **Build errors**: Test cross-compilation locally:
  ```bash
  GOOS=windows GOARCH=amd64 go build
  GOOS=darwin GOARCH=arm64 go build
  ```

## Status Badges

Add these to your README to show workflow status:

```markdown
![CI](https://github.com/zerkz/gsca/workflows/CI/badge.svg)
![Release](https://github.com/zerkz/gsca/workflows/Release/badge.svg)
[![codecov](https://codecov.io/gh/zerkz/gsca/branch/main/graph/badge.svg)](https://codecov.io/gh/zerkz/gsca)
```

## Manual Workflow Triggers

You can also trigger workflows manually from the Actions tab:

1. Go to **Actions** tab
2. Select a workflow
3. Click **Run workflow**
4. Choose branch and parameters

## Secrets and Permissions

### Required Secrets
- `GITHUB_TOKEN`: Automatically provided by GitHub
  - Used for creating releases
  - No manual setup needed

### Optional Secrets
- `CODECOV_TOKEN`: For uploading coverage reports (optional)
  - Sign up at https://codecov.io
  - Add token to repository secrets

### Permissions
The release workflow requires `contents: write` permission to create releases.
This is already configured in `release.yml`.

## Cost and Resource Usage

GitHub Actions is free for public repositories with these limits:
- **Linux/Windows runners**: Unlimited minutes
- **macOS runners**: 2,000 minutes/month (then paid)

Current usage per workflow run:
- **CI**: ~5-10 minutes (all jobs combined)
- **Release**: ~2-5 minutes

Estimated monthly usage:
- 10 PRs/month × 10 min = 100 minutes
- 4 releases/month × 5 min = 20 minutes
- **Total**: ~120 minutes/month

Well within free tier limits.

## Optimization Tips

1. **Cache Go modules**: Already implemented, saves ~30 seconds per run
2. **Matrix strategy**: Tests multiple versions in parallel
3. **Skip redundant runs**: Use `paths` filter if needed
4. **Build only on merge**: Separate quick checks from full builds

Example path filter:
```yaml
on:
  push:
    paths:
      - '**.go'
      - 'go.mod'
      - 'go.sum'
```

## Troubleshooting

### Tests Pass Locally But Fail in CI

Possible causes:
- **Platform-specific code**: Use build tags or conditional compilation
- **Timing issues**: Increase timeouts for flaky tests
- **File path separators**: Use `filepath.Join()` instead of hardcoded `/` or `\`
- **Environment differences**: Check for assumptions about file locations

### Release Workflow Doesn't Trigger

Check:
- Tag format matches `v*` pattern
- Tag was pushed to origin: `git push origin v1.0.0`
- Workflow file is in `main` branch (not in tag)

### Binaries Are Too Large

Try:
- Using `-ldflags="-s -w"` (already done)
- Using UPX compression (not recommended for Go)
- Removing unused dependencies
- Building with `CGO_ENABLED=0` for static linking

## Further Reading

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go GitHub Actions](https://github.com/actions/setup-go)
- [golangci-lint GitHub Action](https://github.com/golangci/golangci-lint-action)
- [Cross-compilation in Go](https://go.dev/doc/install/source#environment)
