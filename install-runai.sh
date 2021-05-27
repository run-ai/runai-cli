#!/usr/bin/env bash
set -e

SCRIPT_NAME=runai
OLD_SCRIPT_FILES=/etc/runai

# If first argument is not empty,
# use that for the installation path
NEW_SCRIPT_FILES=${1:-/usr/local/runai}

SCRIPT_DIR="$(cd "$(dirname "$(readlink "$0" || echo "$0")")"; pwd)"

# Remove old version files
if [ -d "${OLD_SCRIPT_FILES}" ]; then
  rm -rf "${OLD_SCRIPT_FILES}"
fi

# Remove new version files
if [ -d "${NEW_SCRIPT_FILES}" ]; then
  rm -rf "${NEW_SCRIPT_FILES}"
fi

# Create copy destination if it doesn't exist to have directories copied under the folder.
if [ ! -d "${NEW_SCRIPT_FILES}" ]; then
  mkdir "${NEW_SCRIPT_FILES}"
fi

cp "${SCRIPT_DIR}"/runai "${NEW_SCRIPT_FILES}"
cp "${SCRIPT_DIR}"/VERSION "${NEW_SCRIPT_FILES}"
cp -R "${SCRIPT_DIR}"/charts "${NEW_SCRIPT_FILES}"

ln -sf "${NEW_SCRIPT_FILES}"/"${SCRIPT_NAME}" /usr/local/bin/"${SCRIPT_NAME}"