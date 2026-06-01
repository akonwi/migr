#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=false
if [ "${1:-}" = "--dry-run" ]; then
  DRY_RUN=true
  shift
fi

if [ $# -ne 1 ]; then
  echo "Usage: $0 [--dry-run] <version>"
  echo "  e.g. $0 --dry-run v0.1.0"
  exit 1
fi
VERSION="$1"
SEMVER="${VERSION#v}"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUTDIR="$REPO_ROOT/ard-out/release"
mkdir -p "$OUTDIR"

PLATFORMS=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/arm64"
  "linux/amd64"
)

# Ensure working tree is clean
if ! git diff --quiet HEAD 2>/dev/null; then
  echo "Error: working tree has uncommitted changes. Commit or stash first."
  exit 1
fi

if git rev-parse "refs/tags/$VERSION" >/dev/null 2>&1; then
  echo "Error: tag $VERSION already exists"
  exit 1
fi

if [ "$DRY_RUN" = true ]; then
  echo "==> Dry run — skipping tag creation"
else
  echo "==> Tagging $VERSION …"
  git tag -a "$VERSION" -m "$VERSION"
fi

echo ""
echo "==> Building migr $VERSION"

for plat in "${PLATFORMS[@]}"; do
  GOOS="${plat%/*}"
  GOARCH="${plat#*/}"
  NAME="migr-${GOOS}-${GOARCH}"

  echo "  Building $NAME …"
  cd "$REPO_ROOT"
  GOOS="$GOOS" GOARCH="$GOARCH" ard build main.ard --out "$OUTDIR/migr" 2>&1 | sed 's/^/    /'

  echo "  Packing $NAME.tar.gz …"
  tar czf "$OUTDIR/$NAME.tar.gz" -C "$OUTDIR" migr
done

echo ""
echo "==> SHA256 checksums"
echo ""
cd "$OUTDIR"
sha256sum *.tar.gz | tee "$OUTDIR/SHA256SUMS.txt"
echo ""

echo "==> Create the release with:"
echo ""
echo "  gh release create $VERSION \\"
for plat in "${PLATFORMS[@]}"; do
  echo "    migr-${plat%/*}-${plat#*/}.tar.gz \\"
done
echo "    --title \"$VERSION\" \\"
echo "    --notes \"TODO: write release notes\""
echo ""

echo "==> Formula for homebrew-tap/Formula/migr.rb"
echo ""

DARWIN_ARM64_SHA=$(sha256sum migr-darwin-arm64.tar.gz | cut -d' ' -f1)
DARWIN_AMD64_SHA=$(sha256sum migr-darwin-amd64.tar.gz | cut -d' ' -f1)
LINUX_ARM64_SHA=$(sha256sum migr-linux-arm64.tar.gz | cut -d' ' -f1)
LINUX_AMD64_SHA=$(sha256sum migr-linux-amd64.tar.gz | cut -d' ' -f1)

cat <<FORMULA
class Migr < Formula
  desc "Database migration tool"
  homepage "https://github.com/akonwi/migr"
  version "$SEMVER"
  license "MIT"

  if OS.mac?
    if Hardware::CPU.arm?
      url "https://github.com/akonwi/migr/releases/download/$VERSION/migr-darwin-arm64.tar.gz"
      sha256 "$DARWIN_ARM64_SHA"
    else
      url "https://github.com/akonwi/migr/releases/download/$VERSION/migr-darwin-amd64.tar.gz"
      sha256 "$DARWIN_AMD64_SHA"
    end
  elsif OS.linux?
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/akonwi/migr/releases/download/$VERSION/migr-linux-arm64.tar.gz"
      sha256 "$LINUX_ARM64_SHA"
    else
      url "https://github.com/akonwi/migr/releases/download/$VERSION/migr-linux-amd64.tar.gz"
      sha256 "$LINUX_AMD64_SHA"
    end
  end

  def install
    bin.install "migr"
  end
end
FORMULA

echo ""
echo "==> All done."
echo ""
if [ "$DRY_RUN" = false ]; then
  echo "Push the tag:"
  echo "  git push origin $VERSION"
  echo ""
fi
echo "Publish the release:"
echo "  cd $OUTDIR"
echo "  gh release create $VERSION *.tar.gz --title \"$VERSION\" --notes \"...\""
echo ""
echo "Then update homebrew-tap:"
echo "  Copy the formula above into homebrew-tap/Formula/migr.rb"
echo "  cd ~/Developer/agent/homebrew-tap"
echo "  git add -A && git commit -m \"migr $VERSION\" && git push"
