# AUR Packaging Guide for GSCA

This guide explains how to package and publish GSCA to the Arch User Repository (AUR).

## Package Variants

Three PKGBUILD variants are commonly provided for AUR packages:

1. **`gsca`** - Stable release from tagged versions
2. **`gsca-git`** - Latest development version from git HEAD
3. **`gsca-bin`** - Pre-compiled binary (optional, for faster installation)

This repository includes:
- `PKGBUILD` - Stable release package
- `PKGBUILD-git` - Git development package

## Prerequisites

1. **Arch Linux** (or Arch-based distro)
2. **AUR account** - Register at https://aur.archlinux.org
3. **SSH key** configured with your AUR account
4. **Base development tools**:
   ```bash
   sudo pacman -S --needed base-devel git
   ```

## Publishing to AUR

### Initial Setup

1. **Clone the AUR repository** (after creating the package on AUR):
   ```bash
   git clone ssh://aur@aur.archlinux.org/gsca.git aur-gsca
   cd aur-gsca
   ```

2. **Copy the PKGBUILD**:
   ```bash
   cp /path/to/gsca/PKGBUILD .
   ```

3. **Generate checksums**:
   ```bash
   updpkgsums
   # Or manually with:
   makepkg -g
   ```

4. **Generate .SRCINFO**:
   ```bash
   makepkg --printsrcinfo > .SRCINFO
   ```

5. **Test the build locally**:
   ```bash
   makepkg -si
   ```

6. **Commit and push to AUR**:
   ```bash
   git add PKGBUILD .SRCINFO
   git commit -m "Initial commit: gsca v1.0.0"
   git push
   ```

### Updating the Package

When releasing a new version:

1. Update `pkgver` in PKGBUILD
2. Reset `pkgrel` to 1
3. Update checksums: `updpkgsums`
4. Regenerate .SRCINFO: `makepkg --printsrcinfo > .SRCINFO`
5. Test build: `makepkg -si`
6. Commit and push:
   ```bash
   git add PKGBUILD .SRCINFO
   git commit -m "Update to v1.1.0"
   git push
   ```

## Publishing the -git Variant

For `gsca-git`:

1. Clone AUR repo:
   ```bash
   git clone ssh://aur@aur.archlinux.org/gsca-git.git aur-gsca-git
   cd aur-gsca-git
   ```

2. Copy PKGBUILD-git as PKGBUILD:
   ```bash
   cp /path/to/gsca/PKGBUILD-git PKGBUILD
   ```

3. Follow the same steps as above (no need for updpkgsums since source is git)

## PKGBUILD Best Practices

### Required Fields
- `pkgname` - Package name (lowercase, no spaces)
- `pkgver` - Version number
- `pkgrel` - Release number (starts at 1, increment for packaging changes)
- `pkgdesc` - Short description (max 80 chars)
- `arch` - Supported architectures
- `url` - Project homepage/repository
- `license` - Software license
- `makedepends` - Build dependencies
- `source` - Source files to download
- `sha256sums` - Checksums for source files

### Go-Specific Requirements

The PKGBUILD includes these required Go build flags per Arch guidelines:

```bash
export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
```

- **`-buildmode=pie`** - Position-independent executable (security)
- **`-trimpath`** - Reproducible builds
- **`-mod=readonly`** - Prevent module modification during build
- **`-modcacherw`** - Make module cache writable after build

### CGO Flags

If your Go application uses CGO (C bindings):

```bash
export CGO_CPPFLAGS="${CPPFLAGS}"
export CGO_CFLAGS="${CFLAGS}"
export CGO_CXXFLAGS="${CXXFLAGS}"
export CGO_LDFLAGS="${LDFLAGS}"
```

## Automated Publishing with GoReleaser

You can automate AUR releases with GoReleaser:

1. Install GoReleaser:
   ```bash
   yay -S goreleaser
   ```

2. Add to `.goreleaser.yml`:
   ```yaml
   aur:
     - name: gsca
       homepage: "https://github.com/zerkz/gsca"
       description: "CLI tool to manage Steam game launch options"
       maintainers:
         - "Your Name <your.email@example.com>"
       license: "MIT"
       private_key: "{{ .Env.AUR_SSH_KEY }}"
       git_url: "ssh://aur@aur.archlinux.org/gsca.git"
   ```

3. Set up AUR SSH key as environment variable
4. Run `goreleaser release` on tagged commits

## Testing Locally

Before publishing:

1. **Build the package**:
   ```bash
   makepkg
   ```

2. **Install locally**:
   ```bash
   makepkg -si
   ```

3. **Test the installed binary**:
   ```bash
   gsca --help
   ```

4. **Remove test installation**:
   ```bash
   sudo pacman -R gsca
   ```

## AUR Package Naming Conventions

- **`pkgname`** - Stable version from source
- **`pkgname-git`** - Latest commit from git
- **`pkgname-bin`** - Pre-built binary
- **`pkgname-devel`** - Development version (alternative to -git)

## Maintenance

As the package maintainer, you should:

1. **Respond to comments** on the AUR page
2. **Update promptly** when new versions are released
3. **Test builds** before pushing updates
4. **Mark out-of-date** packages if you can't maintain them
5. **Orphan the package** if you no longer want to maintain it

## Additional Resources

- [Arch Wiki - AUR Guidelines](https://wiki.archlinux.org/title/Arch_User_Repository)
- [Arch Wiki - Go Package Guidelines](https://wiki.archlinux.org/title/Go_package_guidelines)
- [Arch Wiki - PKGBUILD](https://wiki.archlinux.org/title/PKGBUILD)
- [AUR Official Site](https://aur.archlinux.org)
- [GoReleaser AUR Documentation](https://goreleaser.com/customization/aur/)

## License Considerations

Make sure you have a LICENSE file in your repository. Common licenses for CLI tools:

- MIT - Simple and permissive
- GPL-3.0 - Copyleft
- Apache-2.0 - Permissive with patent grant

The PKGBUILD currently assumes MIT license. Update if different.
