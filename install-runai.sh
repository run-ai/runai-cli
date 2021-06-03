#!/usr/bin/env bash
set -e

SCRIPT_NAME=runai

# If first argument is not empty,
# use that for the installation path
NEW_SCRIPT_FILES=${1:-/usr/local/runai}

SCRIPT_DIR="$(cd "$(dirname "$(readlink "$0" || echo "$0")")"; pwd)"

# Create copy destination if it doesn't exist to have directories copied under the folder.
if [ ! -d "${NEW_SCRIPT_FILES}" ]; then
    if [ "$NEW_SCRIPT_FILES" == "/usr/local/runai" ]; then
        mkdir "${NEW_SCRIPT_FILES}"
    else
        echo "${NEW_SCRIPT_FILES} doesn't exist or is not a directory"
        ls "${NEW_SCRIPT_FILES}" 2> /dev/null
    fi
fi

cp "${SCRIPT_DIR}"/runai "${NEW_SCRIPT_FILES}"
cp "${SCRIPT_DIR}"/VERSION "${NEW_SCRIPT_FILES}"
cp -R "${SCRIPT_DIR}"/charts "${NEW_SCRIPT_FILES}"

if [ "$NEW_SCRIPT_FILES" == "/usr/local/runai" ] ; then
    ln -sf "${NEW_SCRIPT_FILES}"/"${SCRIPT_NAME}" /usr/local/bin/"${SCRIPT_NAME}"
else
    echo "Add ${NEW_SCRIPT_FILES} to your \$PATH: export PATH=\$PATH:${NEW_SCRIPT_FILES}"
fi

echo "Run:AI CLI installed successfully!"
