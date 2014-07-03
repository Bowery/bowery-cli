#!/bin/bash
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/../" && pwd )"

# Change into that dir because we expect that
cd $DIR

BINTRAY_API_KEY="7683716684ff54cbebb896a92be8fa6749c17ba6"

# Determine the version that we're building based on the contents
# of cli/VERSION.txt.
VERSION=$(cat VERSION.txt)
VERSIONDIR="${VERSION}"
echo "Version: ${VERSION}"

# Determine the arch/os combos we're building for
XC_ARCH=${XC_ARCH:-"386 amd64"}
XC_OS=${XC_OS:-linux darwin windows}

echo "Arch: ${XC_ARCH}"
echo "OS: ${XC_OS}"

# Make sure that if we're killed, we kill all our subprocseses
trap "kill 0" SIGINT SIGTERM EXIT

# Make sure goxc is installed
go get github.com/laher/goxc

# This function builds whatever directory we're in...
goxc \
    -arch="$XC_ARCH" \
    -os="$XC_OS" \
    -d="${DIR}/pkg" \
    -pv="${VERSION}" \
    $XC_OPTS \
    go-install \
    xc

# Zip all the packages
mkdir -p ./pkg/${VERSIONDIR}/dist
for PLATFORM in $(find ./pkg/${VERSIONDIR} -mindepth 1 -maxdepth 1 -type d); do
    PLATFORM_NAME=$(basename ${PLATFORM})
    ARCHIVE_NAME="${VERSIONDIR}_${PLATFORM_NAME}"

    if [ $PLATFORM_NAME = "dist" ]; then
        continue
    fi

    pushd ${PLATFORM}
    zip ${DIR}/pkg/${VERSIONDIR}/dist/${ARCHIVE_NAME}.zip ./*
    popd
done

# Make the checksums
pushd ./pkg/${VERSIONDIR}/dist
shasum -a256 * > ./${VERSIONDIR}_SHA256SUMS
popd

for ARCHIVE in ./pkg/${VERSION}/dist/*; do
    ARCHIVE_NAME=$(basename ${ARCHIVE})

    echo Uploading: $ARCHIVE_NAME from $ARCHIVE
    curl \
        -T ${ARCHIVE} \
        -ustevekaliski:7683716684ff54cbebb896a92be8fa6749c17ba6 \
        "https://api.bintray.com/content/bowery/bowery/bowery/${VERSION}/${ARCHIVE_NAME}"
    echo
done

exit 0
