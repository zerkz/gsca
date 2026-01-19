# CLAUDE.md

Project-specific instructions for Claude Code.

## Style Guidelines

- Do not use emojis in documentation or code comments
- Minimalistic approach while keeping good User Experience in mind.
- Keep documentation tight, nothing should be duplicated, Github markdown can link to other files.
- README.md should have this in order of importance: install instructions, usage, and then pointers to LICENSE/Contributing/technical background.

## Releasing

```bash
./scripts/release.sh 1.2.0
```

The script will:
1. Run tests and linter
2. Verify GoReleaser config
3. Update version in `PKGBUILD` and `com.github.zerkz.gsca.yaml`
4. Commit, tag, and push

GitHub Actions then builds binaries, creates the release, and uploads the Flatpak bundle.
