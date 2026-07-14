#!/usr/bin/env bash
set -euo pipefail

# Environment variables
BUILD_DIR="/tmp/pg_deb_build"

CWD="$(pwd)"

export DEBIAN_FRONTEND=noninteractive
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
fi

# Release assets are named with the full tag version (e.g. 1.0.0-beta2).
# DOCLOADER_VERSION may carry a '~beta…' suffix for DEB ordering, so it must
# NOT be used to build the download URL — use the tag instead.
TAG_VERSION="${DOCLOADER_BRANCH#v}"
SRC_DIR="${BUILD_DIR}/pgedge-docloader-${TAG_VERSION}"
# The workflow stages the GoReleaser tarball + LICENCE here; prefer them over
# wget so simulate_tag runs (no published release) still build.
ARTIFACT_DIR="${ARTIFACT_DIR:-${CWD}/release-artifacts}"

prepare() {

  setup_apt_build_env

  # This function is for debugging purpose if you have your own keys. GH workflow does not need it.
  #import_gpg_keys
  
  echo "Cloning docloader Debian packaging repo..."
  rm -rf "$SRC_DIR"
  mkdir -p $SRC_DIR

  echo "Fetching docloader source code"
  if [ -f "${ARTIFACT_DIR}/docloader.tar.gz" ]; then
    echo "Using staged tarball ${ARTIFACT_DIR}/docloader.tar.gz"
    cp "${ARTIFACT_DIR}/docloader.tar.gz" "${BUILD_DIR}/docloader.tar.gz"
  else
    echo "Downloading source tarball from release ${DOCLOADER_BRANCH}"
    wget -q "https://github.com/pgEdge/pgedge-docloader/releases/download/${DOCLOADER_BRANCH}/pgedge-docloader_${TAG_VERSION}_Linux_${ARCH}.tar.gz" -O "${BUILD_DIR}/docloader.tar.gz"
  fi
  tar -C "$SRC_DIR" -xzf "${BUILD_DIR}/docloader.tar.gz"

  echo "Moving Debian packaging into source directory..."
  cp -rp "${CWD}/${COMPONENT_NAME}/deb/debian" "$SRC_DIR/"
  cp ${COMPONENT_NAME}/common/config.yml "$SRC_DIR/debian/"

  if [ -f "${ARTIFACT_DIR}/LICENCE.md" ]; then
    cp "${ARTIFACT_DIR}/LICENCE.md" "$SRC_DIR/"
  else
    wget -q "https://raw.githubusercontent.com/pgEdge/pgedge-docloader/${DOCLOADER_BRANCH}/LICENCE.md" -O "$SRC_DIR/LICENCE.md"
  fi

  echo "Installing build dependencies..."
  cd "$SRC_DIR"
  sudo apt-get update
  sudo apt-get build-dep -y .
}

build() {

  cd "$SRC_DIR"
  echo "Building Debian package..."
  DISTRO=$(lsb_release -cs)
  rm -f debian/changelog
cat > debian/changelog <<EOF
pgedge-docloader (${DOCLOADER_VERSION}-${DOCLOADER_BUILDNUM}.${DISTRO}) stable; urgency=medium

  * Initial pgedge-docloader package.

 -- Muhammad Aqeel <muhammad.aqeel@pgedge.com>  $(date -R)
EOF

  dpkg-buildpackage -us -uc -b
}

post_build() {
  echo "Copying .deb packages to output..."
  sudo mkdir -p "/output"
  # Rename .ddeb files to .deb files
  rename_ddeb_packages $BUILD_DIR
  sudo cp "$BUILD_DIR"/*.deb "/output" || echo "No .deb packages found."
}

