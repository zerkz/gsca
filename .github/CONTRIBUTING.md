# Contributing to GSCA

Thank you for your interest in contributing to GSCA!

## Development Setup

1. **Fork and clone the repository**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/gsca.git
   cd gsca
   ```

2. **Install Go** (version 1.21 or later):
   - Download from https://go.dev/dl/

3. **Install dependencies**:
   ```bash
   go mod download
   ```

4. **Build the project**:
   ```bash
   go build -o gsca
   ```

5. **Run tests**:
   ```bash
   go test ./...
   ```

## Making Changes

1. **Create a new branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** and ensure:
   - Code is formatted: `go fmt ./...`
   - Tests pass: `go test ./...`
   - Linter is happy: `golangci-lint run`

3. **Write tests** for new functionality

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "Add feature: your feature description"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request** on GitHub

## Continuous Integration

When you open a pull request, the following checks will run automatically:

### Test Suite
- Tests run on Linux, macOS, and Windows
- Tests run on Go versions 1.21, 1.22, and 1.23
- Code coverage is measured and reported

### Linting
- Code is checked with `golangci-lint`
- Ensures code quality and consistency

### Build Verification
- Binaries are built for Linux, Windows, and macOS
- Ensures the code compiles on all platforms

All checks must pass before a PR can be merged.

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Add comments for exported functions and types
- Keep functions focused and concise
- Write descriptive commit messages

## Testing

- Write unit tests for new functionality
- Ensure tests are platform-independent when possible
- Use table-driven tests for multiple test cases
- Mock external dependencies (Steam paths, file I/O, etc.)

Example test:
```go
func TestParseSelection(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        max      int
        expected []int
    }{
        {"single number", "1", 5, []int{0}},
        {"range", "1-3", 5, []int{0, 1, 2}},
        {"wildcard", "*", 3, []int{0, 1, 2}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := parseSelection(tt.input, tt.max)
            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Release Process

Releases are automated through GitHub Actions:

1. **Tag a new version**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

2. **GitHub Actions automatically**:
   - Builds binaries for all platforms
   - Generates checksums
   - Creates a GitHub release
   - Attaches all binaries to the release

## Questions?

If you have questions or need help:
- Open an issue on GitHub
- Check existing issues and PRs
- Read the documentation in README.md
- Maintainer(s) reserve the right to work at the pace they chose, no one is owed anything.

Thank you for contributing!
