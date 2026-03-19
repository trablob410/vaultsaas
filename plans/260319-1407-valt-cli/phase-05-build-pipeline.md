# Phase 05: Build Pipeline + Install Script

**Priority:** P2
**Status:** COMPLETED — 2026-03-19

## Goal
Cross-compile `valt` binary for Mac/Linux/Windows, publish to GitHub Releases, and provide a one-line install script.

## Files

| File | Change |
|------|--------|
| `valt-cli/.goreleaser.yml` | Create |
| `scripts/install-valt-cli.sh` | Create |
| `.github/workflows/release-valt-cli.yml` | Create |

---

## Task 1: .goreleaser.yml

**Files:**
- Create: `valt-cli/.goreleaser.yml`

- [ ] Write:

```yaml
version: 2

project_name: valt

before:
  hooks:
    - go mod tidy

builds:
  - id: valt
    dir: valt-cli
    main: .
    binary: valt
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - id: valt
    builds: [valt]
    name_template: "valt_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: "checksums.txt"

release:
  github:
    owner: valt-dev
    name: valt

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^chore:"
```

- [ ] Test locally:
```bash
cd valt-cli && goreleaser build --snapshot --clean
ls dist/
```
Expected: binaries for linux_amd64, darwin_amd64, darwin_arm64, windows_amd64

- [ ] Commit:
```bash
git add valt-cli/.goreleaser.yml
git commit -m "chore(valt-cli): add goreleaser config for cross-platform builds"
```

---

## Task 2: Install script

**Files:**
- Create: `scripts/install-valt-cli.sh`

- [ ] Write:

```bash
#!/usr/bin/env bash
set -euo pipefail

REPO="valt-dev/valt"
BINARY="valt"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Fetch latest release tag
LATEST=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release"
  exit 1
fi

ARCHIVE="${BINARY}_${OS}_${ARCH}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ARCHIVE}.tar.gz"

echo "Installing valt ${LATEST} for ${OS}/${ARCH}..."
TMP=$(mktemp -d)
trap "rm -rf $TMP" EXIT

curl -sSL "$URL" -o "${TMP}/${ARCHIVE}.tar.gz"
tar -xzf "${TMP}/${ARCHIVE}.tar.gz" -C "$TMP"
chmod +x "${TMP}/${BINARY}"
mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

echo "✓ valt installed to ${INSTALL_DIR}/${BINARY}"
echo "  Run: valt setup"
```

- [ ] Make executable: `chmod +x scripts/install-valt-cli.sh`

- [ ] Commit:
```bash
git add scripts/install-valt-cli.sh
git commit -m "chore(valt-cli): add one-line install script for GitHub Releases"
```

---

## Task 3: GitHub Actions release workflow

**Files:**
- Create: `.github/workflows/release-valt-cli.yml`

- [ ] Write:

```yaml
name: Release valt CLI

on:
  push:
    tags:
      - 'cli/v*'   # trigger on tags like cli/v0.1.0

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
          workdir: valt-cli
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] Commit:
```bash
git add .github/workflows/release-valt-cli.yml
git commit -m "ci: add GitHub Actions release workflow for valt CLI"
```

---

## Task 4: README for valt-cli

**Files:**
- Create: `valt-cli/README.md`

- [ ] Write minimal usage doc:

```markdown
# valt CLI

Secret vault CLI for the Valt platform.

## Install

```bash
curl -sSL https://raw.githubusercontent.com/valt-dev/valt/main/scripts/install-valt-cli.sh | sh
```

## Quick start

```bash
valt setup                      # configure + login + pick project
valt mcp install --ide claude   # install MCP server for Claude Desktop
valt list                       # list accessible secrets
valt get DB_PASSWORD            # get a secret value
valt run -- node server.js      # inject secrets + run command
valt request DB_PASSWORD --reason "fixing prod" --duration 2h
valt status <request-id>
```
```

- [ ] Commit:
```bash
git add valt-cli/README.md
git commit -m "docs(valt-cli): add CLI README with install and usage instructions"
```

---

## Success Criteria
- `goreleaser build --snapshot --clean` produces binaries for all 4 targets
- Install script downloads and installs binary from GitHub Releases
- `valt --version` prints correct version after install
- GitHub Actions workflow triggers on `cli/v*` tags and publishes release
