#!/bin/sh

# DO NOT EDIT. Generated with:
#
#    devctl@5.9.0
#

APPLICATION=$1
VERSION=$2

SIGNCODE_UTIL=quay.io/giantswarm/signcode-util:1.1.1

echo "APPLICATION=${APPLICATION}"
echo "VERSION=${VERSION}"
echo "PWD=${PWD}"

NO_CODE_SIGNING="Skipping Windows binary signing. In order to create a signed Windows binary, set the environment variables CODE_SIGNING_CERT_BUNDLE_PASSWORD and CODE_SIGNING_CERT_BUNDLE_BASE64."

if [ "${CODE_SIGNING_CERT_BUNDLE_PASSWORD}" = "" ]; then
	echo "Variable CODE_SIGNING_CERT_BUNDLE_PASSWORD not set. ${NO_CODE_SIGNING}"
	exit 0
fi;

if [ "${CODE_SIGNING_CERT_BUNDLE_BASE64}" = "" ]; then
	echo "Variable CODE_SIGNING_CERT_BUNDLE_BASE64 not set.  ${NO_CODE_SIGNING}"
	exit 0
fi;

echo "Signing the Windows binary"

mkdir -p certs

echo "${CODE_SIGNING_CERT_BUNDLE_BASE64}" | base64 -d > certs/code-signing.p12

mv "${APPLICATION}-v${VERSION}-windows-amd64.exe" "${APPLICATION}-v${VERSION}-windows-amd64-unsigned.exe"

docker pull --quiet ${SIGNCODE_UTIL}

docker run --rm \
	-v "${PWD}/certs:/mnt/certs" \
	-v "${PWD}:/mnt/binaries" \
	${SIGNCODE_UTIL} \
	sign \
	-pkcs12 /mnt/certs/code-signing.p12 \
	-n "Giant Swarm CLI tool ${APPLICATION}" \
	-i "https://github.com/giantswarm/${APPLICATION}" \
	-t http://timestamp.digicert.com -verbose \
	-in "/mnt/binaries/${APPLICATION}-v${VERSION}-windows-amd64-unsigned.exe" \
	-out "/mnt/binaries/${APPLICATION}-v${VERSION}-windows-amd64.exe" \
	-pass "${CODE_SIGNING_CERT_BUNDLE_PASSWORD}"

echo "Verifying the signed binary"

docker run --rm \
	-v "${PWD}:/mnt/binaries" \
	${SIGNCODE_UTIL} \
	verify \
	"/mnt/binaries/${APPLICATION}-v${VERSION}-windows-amd64.exe"
