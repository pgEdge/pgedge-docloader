#!/usr/bin/env bash
# common.sh - Common environment variables

# Default PostgreSQL version and derived values
export PG_VERSION="${PG_VERSION:-17.7}"
export PG_MAJOR_VERSION="$(echo "$PG_VERSION" | cut -d. -f1)"

export DOCLOADER_REPO="https://github.com/pgEdge/pgedge-docloader.git"
export DOCLOADER_BRANCH="${COMPONENT_BRANCH:-v1.0.0}"
export DOCLOADER_VERSION=${COMPONENT_VERSION:-1.0.0}
export DOCLOADER_BUILDNUM=${COMPONENT_BUILDNUM:-1}

# DEB only: move a pre-release pretag (e.g. BUILDNUM='beta3_1') into the
# upstream VERSION with a leading '~' (1.0.0~beta3, BUILDNUM=1) so '~' sorts
# pre-releases BELOW stable in dpkg/reprepro. Downloads use the tag
# (DOCLOADER_BRANCH), not VERSION, so this never affects the source URL.
if command -v apt-get &>/dev/null; then
    if [[ "$DOCLOADER_BUILDNUM" == *_* ]]; then
        DOCLOADER_PRETAG="${DOCLOADER_BUILDNUM%%_*}"
        export DOCLOADER_VERSION="${DOCLOADER_VERSION}~${DOCLOADER_PRETAG}"
        DOCLOADER_BUILDNUM="${DOCLOADER_BUILDNUM#*_}"
    fi
fi

export REPO_TYPE="${REPO_TYPE:-daily}"
