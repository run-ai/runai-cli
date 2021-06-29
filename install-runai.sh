#!/usr/bin/env bash
set -e

SCRIPT_NAME=runai

# If first argument is not empty,
# use that for the installation path
NEW_SCRIPT_PATH=${1:-/usr/local/runai}

SCRIPT_DIR="$(cd "$(dirname "$(readlink "$0" || echo "$0")")"; pwd)"

# Remove old version files
if [ -d "${NEW_SCRIPT_PATH}" ]; then
  rm -f "${NEW_SCRIPT_PATH}/runai"
  rm -f "${NEW_SCRIPT_PATH}/VERSION"
  rm -rf "${NEW_SCRIPT_PATH}/charts"
fi

# Create copy destination if it doesn't exist to have directories copied under the folder.
if [ ! -d "${NEW_SCRIPT_PATH}" ]; then
    if [ "${NEW_SCRIPT_PATH}" == "/usr/local/runai" ]; then
        mkdir "${NEW_SCRIPT_PATH}"
    else
        echo "${NEW_SCRIPT_PATH} doesn't exist or is not a directory"
        ls "${NEW_SCRIPT_PATH}" 2> /dev/null
    fi
fi

cp "${SCRIPT_DIR}"/runai "${NEW_SCRIPT_PATH}"
cp "${SCRIPT_DIR}"/VERSION "${NEW_SCRIPT_PATH}"
cp -R "${SCRIPT_DIR}"/charts "${NEW_SCRIPT_PATH}"

if [ "$NEW_SCRIPT_PATH" == "/usr/local/runai" ] ; then
    if [ ! -d "/usr/local/bin" ] ; then
        mkdir -p /usr/local/bin
        echo "Add \"/usr/local/bin\" to your PATH: export PATH=\$PATH:/usr/local/bin"
    fi
    ln -sf "${NEW_SCRIPT_PATH}"/"${SCRIPT_NAME}" /usr/local/bin/"${SCRIPT_NAME}"
else
    echo "Add ${NEW_SCRIPT_PATH} to your \$PATH: export PATH=\$PATH:${NEW_SCRIPT_PATH}"
fi

echo "Run:AI CLI installed successfully!"
