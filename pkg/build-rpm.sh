#!/bin/bash
set -euo pipefail

RHEL="$(rpm --eval %rhel)"
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
fi

# Release assets are named with the full tag version (e.g. 1.0.0-beta2) and
# live under the tag. DOCLOADER_VERSION is the plain RPM version (no
# pre-release suffix), so download by TAG, then stage the tarball under the
# spec's expected <version> basename. Mirrors build-deb.sh; avoids a 404 on
# pre-release tags (e.g. v1.0.0-test1) where v${DOCLOADER_VERSION} wouldn't
# resolve.
TAG_VERSION="${DOCLOADER_BRANCH#v}"
# The workflow stages the GoReleaser tarball + LICENCE here; prefer them over
# wget so simulate_tag runs (no published release) still build.
ARTIFACT_DIR="${ARTIFACT_DIR:-$(pwd)/release-artifacts}"

prepare() {
  setup_dnf_build_env

  echo "Copying packaging files..."
  cp ${COMPONENT_NAME}/rpm/docloader.spec ~/rpmbuild/SPECS/

  echo "Staging source tarball and docs ..."
  # Stage under the basename the spec's Source0 expects (docloader_version),
  # regardless of the tag's pre-release suffix. Prefer the workflow-staged
  # tarball (so simulate_tag builds work) and fall back to the GitHub release.
  SRC_TARBALL=~/rpmbuild/SOURCES/pgedge-docloader_${DOCLOADER_VERSION}_Linux_${ARCH}.tar.gz
  if [ -f "${ARTIFACT_DIR}/docloader.tar.gz" ]; then
    echo "Using staged tarball ${ARTIFACT_DIR}/docloader.tar.gz"
    cp "${ARTIFACT_DIR}/docloader.tar.gz" "${SRC_TARBALL}"
  else
    echo "Downloading source tarball from release ${DOCLOADER_BRANCH}"
    wget -q "https://github.com/pgEdge/pgedge-docloader/releases/download/${DOCLOADER_BRANCH}/pgedge-docloader_${TAG_VERSION}_Linux_${ARCH}.tar.gz" -O "${SRC_TARBALL}"
  fi
  cp ${COMPONENT_NAME}/common/config.yml ~/rpmbuild/SOURCES/
  if [ -f "${ARTIFACT_DIR}/LICENCE.md" ]; then
    cp "${ARTIFACT_DIR}/LICENCE.md" ~/rpmbuild/SOURCES/
  else
    wget -q "https://raw.githubusercontent.com/pgEdge/pgedge-docloader/${DOCLOADER_BRANCH}/LICENCE.md" -O ~/rpmbuild/SOURCES/LICENCE.md
  fi

  # This function is for debugging purpose if you have your own keys. GH workflow does not need it.
  #import_gpg_keys

  echo "🔧 Installing RPM build dependencies..."
  dnf builddep -y \
    --define "docloader_version ${DOCLOADER_VERSION}" \
    --define "docloader_buildnum ${DOCLOADER_BUILDNUM}" \
    --define "arch ${ARCH}" \
    ~/rpmbuild/SPECS/docloader.spec
}

build() {
  echo "Building RPM and SRPM..."
  QA_RPATHS=$(( 0xffff )) rpmbuild -ba ~/rpmbuild/SPECS/docloader.spec \
    --define "docloader_version ${DOCLOADER_VERSION}" \
    --define "docloader_buildnum ${DOCLOADER_BUILDNUM}" \
    --define "arch ${ARCH}"
}

post_build() {
  echo "📤 Copying built RPMs to /output..."
  mkdir -p /output
  cp -v ~/rpmbuild/RPMS/*/*.rpm /output/ || echo "No binary RPMs found"
  cp -v ~/rpmbuild/SRPMS/*.src.rpm /output/ || echo "No SRPM found"

  sign_rpms /output/*.rpm
  validate_signatures /output/*.rpm
}

