#!/bin/bash -eu

# USAGE:
# ./utils.sh

# creates usefull functions for other scripts to use.

# fpChanged verifies the modified files in the finality provider registry and returns it
# as a list ready to be iterated.
function fpChanged {
  originUrl=$(git remote get-url origin)
  currBranch=$(git rev-parse --abbrev-ref HEAD)
  if [ "main" == "$currBranch" ] && [[ "$originUrl" =~ .*"babylonchain/networks.git".* ]]; then
    git diff HEAD~1 --name-only **/finality-providers/registry/
  else
    if ! git remote get-url bbnRepo &> /dev/null; then
      git remote add bbnRepo https://github.com/babylonchain/networks.git
    fi
    git fetch bbnRepo main --force
    git diff --diff-filter=AM bbnRepo/main --name-only **/finality-providers/registry/
  fi
}

function checkCommandJq {
  if ! command -v jq &> /dev/null; then
    echo "⚠️ jq command could not be found!"
    echo "Install it by checking https://stedolan.github.io/jq/download/"
    exit 1
  fi
}