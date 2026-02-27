#!/usr/bin/env bash
# One-time script: publish placeholder npm packages to reserve names
# and enable OIDC trusted publishing configuration.
#
# Uses NPM_TOKEN from environment (or .env file).
# @ancla org must already exist on npmjs.com.
#
# After running this, go to npmjs.com and configure trusted publishing
# on each package (ancla + all @ancla/* packages):
#   Settings → Trusted Publisher → GitHub Actions
#   Org: SideQuest-Group | Repo: ancla-client | Workflow: release.yml | Env: npm

set -euo pipefail

cd "$(dirname "$0")/.."

# Load .env if present
if [ -f .env ]; then
  set -a; source .env; set +a
fi

if [ -z "${NPM_TOKEN:-}" ]; then
  echo "NPM_TOKEN not set. Export it or add to .env"
  exit 1
fi

# Auth via token
echo "//registry.npmjs.org/:_authToken=${NPM_TOKEN}" > .npmrc_tmp
export npm_config_userconfig="$(pwd)/.npmrc_tmp"

echo "==> Checking npm auth..."
npm whoami || { rm -f .npmrc_tmp; echo "NPM_TOKEN is invalid"; exit 1; }

PACKAGES=(
  "npm/ancla-linux-x64"
  "npm/ancla-linux-arm64"
  "npm/ancla-darwin-x64"
  "npm/ancla-darwin-arm64"
  "npm/ancla-win32-x64"
)

# Publish platform packages first (0.0.1 placeholder)
for pkg_dir in "${PACKAGES[@]}"; do
  echo "==> Publishing ${pkg_dir} (placeholder)..."
  cd "$pkg_dir"
  npm version 0.0.1 --no-git-tag-version --allow-same-version 2>/dev/null
  npm publish --access public || echo "    (may already exist, continuing)"
  cd - > /dev/null
done

# Publish meta-package last
echo "==> Publishing npm/ancla (placeholder)..."
cd npm/ancla
node -e "
  const pkg = require('./package.json');
  pkg.version = '0.0.1';
  for (const dep of Object.keys(pkg.optionalDependencies)) {
    pkg.optionalDependencies[dep] = '0.0.1';
  }
  require('fs').writeFileSync('package.json', JSON.stringify(pkg, null, 2) + '\n');
"
npm publish --access public || echo "    (may already exist, continuing)"
cd - > /dev/null

# Cleanup
rm -f .npmrc_tmp

echo ""
echo "==> Done! Now configure trusted publishing on npmjs.com for each package:"
echo "    ancla, @ancla/linux-x64, @ancla/linux-arm64, @ancla/darwin-x64,"
echo "    @ancla/darwin-arm64, @ancla/win32-x64"
echo ""
echo "    Settings → Trusted Publisher → GitHub Actions"
echo "    Org: SideQuest-Group | Repo: ancla-client | Workflow: release.yml | Env: npm"
